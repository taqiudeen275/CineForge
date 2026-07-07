package store

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ats-tech/cineforge/services/control-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrOwnerBlocked = errors.New("ownership must be resolved first")
)

type Store struct{ pool *pgxpool.Pool }

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) withUserTx(ctx context.Context, userID string, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "select set_config('app.user_id', $1, true)", userID); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type Identity struct {
	UID, Email, DisplayName string
	EmailVerified           bool
	AvatarURL               *string
}

func (s *Store) UpsertUser(ctx context.Context, id Identity) (domain.User, error) {
	user := domain.User{}
	var personalID *string
	err := s.withUserTx(ctx, "", func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
INSERT INTO users (id, identity_provider_uid, email, email_verified, display_name, avatar_url)
VALUES ($1,$2,lower($3),$4,$5,$6)
ON CONFLICT (identity_provider_uid) DO UPDATE SET
  email=excluded.email, email_verified=excluded.email_verified,
  display_name=CASE WHEN users.display_name='' THEN excluded.display_name ELSE users.display_name END,
  avatar_url=COALESCE(users.avatar_url, excluded.avatar_url), updated_at=now(), version=users.version+1
RETURNING id::text, identity_provider_uid, email, email_verified, display_name, avatar_url,
  locale, timezone, personal_workspace_id::text, deletion_scheduled_at, version, created_at, updated_at`,
			uuid.Must(uuid.NewV7()).String(), id.UID, id.Email, id.EmailVerified, fallbackName(id.DisplayName, id.Email), id.AvatarURL)
		return row.Scan(&user.ID, &user.IdentityProviderUID, &user.Email, &user.EmailVerified, &user.DisplayName,
			&user.AvatarURL, &user.Locale, &user.Timezone, &personalID, &user.DeletionScheduledAt,
			&user.Version, &user.CreatedAt, &user.UpdatedAt)
	})
	user.PersonalWorkspaceID = personalID
	return user, err
}

func (s *Store) BootstrapPersonalWorkspace(ctx context.Context, user domain.User) (domain.Workspace, error) {
	if !user.EmailVerified {
		return domain.Workspace{}, ErrForbidden
	}
	var ws domain.Workspace
	err := s.withUserTx(ctx, user.ID, func(tx pgx.Tx) error {
		var existing *string
		if err := tx.QueryRow(ctx, "select personal_workspace_id::text from users where id=$1 for update", user.ID).Scan(&existing); err != nil {
			return err
		}
		if existing != nil {
			return scanWorkspace(tx.QueryRow(ctx, workspaceSelect+" where id=$1", *existing), &ws)
		}
		wid := uuid.Must(uuid.NewV7()).String()
		slug, err := uniqueSlug(ctx, tx, fallbackName(user.DisplayName, user.Email))
		if err != nil {
			return err
		}
		name := fallbackName(user.DisplayName, user.Email) + "'s workspace"
		if _, err := tx.Exec(ctx, `insert into workspaces(id,name,slug,kind,owner_user_id) values($1,$2,$3,'personal',$4)`, wid, name, slug, user.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into memberships(id,workspace_id,user_id,role,status,can_spend,library_publish) values($1,$2,$3,'owner','active',true,true)`, uuid.Must(uuid.NewV7()).String(), wid, user.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into workspace_policies(workspace_id,updated_by) values($1,$2)`, wid, user.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update users set personal_workspace_id=$1 where id=$2`, wid, user.ID); err != nil {
			return err
		}
		if err := audit(ctx, tx, wid, user.ID, "workspace.created", "workspace", wid, map[string]any{"kind": "personal"}); err != nil {
			return err
		}
		return scanWorkspace(tx.QueryRow(ctx, workspaceSelect+" where id=$1", wid), &ws)
	})
	return ws, err
}

const workspaceSelect = `select id::text,name,slug,kind::text,owner_user_id::text,version,deleted_at,purge_at,created_at,updated_at from workspaces`

func scanWorkspace(row pgx.Row, ws *domain.Workspace) error {
	return row.Scan(&ws.ID, &ws.Name, &ws.Slug, &ws.Kind, &ws.OwnerUserID, &ws.Version, &ws.DeletedAt, &ws.PurgeAt, &ws.CreatedAt, &ws.UpdatedAt)
}

func (s *Store) ListWorkspaces(ctx context.Context, userID string) ([]domain.Workspace, error) {
	result := []domain.Workspace{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `select w.id::text,w.name,w.slug,w.kind::text,w.owner_user_id::text,w.version,w.deleted_at,w.purge_at,w.created_at,w.updated_at from workspaces w join memberships m on m.workspace_id=w.id where m.user_id=$1 and m.status='active' and w.deleted_at is null order by w.updated_at desc`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var w domain.Workspace
			if err := rows.Scan(&w.ID, &w.Name, &w.Slug, &w.Kind, &w.OwnerUserID, &w.Version, &w.DeletedAt, &w.PurgeAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
				return err
			}
			result = append(result, w)
		}
		return rows.Err()
	})
	return result, err
}

func (s *Store) Membership(ctx context.Context, userID, workspaceID string) (domain.Membership, error) {
	var m domain.Membership
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select id::text,workspace_id::text,user_id::text,role::text,status::text,can_spend,monthly_limit_micros,per_run_limit_micros,library_publish,created_at,updated_at,removed_at from memberships where workspace_id=$1 and user_id=$2 and status='active'`, workspaceID, userID).Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.Status, &m.CanSpend, &m.MonthlyLimitMicros, &m.PerRunLimitMicros, &m.LibraryPublish, &m.CreatedAt, &m.UpdatedAt, &m.RemovedAt)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrForbidden
	}
	return m, err
}

func (s *Store) ListProjects(ctx context.Context, userID, workspaceID, status, query string) ([]domain.Project, error) {
	projects := []domain.Project{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, workspaceID, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor, domain.RoleReviewer, domain.RoleViewer); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, projectSelect+` where p.workspace_id=$1 and ($2='' or p.status::text=$2) and ($3='' or pv.name ilike '%'||$3||'%') order by p.updated_at desc limit 100`, workspaceID, status, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var p domain.Project
			if err := scanProject(rows, &p); err != nil {
				return err
			}
			projects = append(projects, p)
		}
		return rows.Err()
	})
	return projects, err
}

const projectSelect = `select p.id::text,p.workspace_id::text,p.slug,p.status::text,p.privacy,p.current_version,p.created_by::text,p.deleted_at,p.purge_at,p.created_at,p.updated_at,
pv.name,pv.project_type,pv.production_format,pv.aspect_width,pv.aspect_height,pv.frame_rate_numerator,pv.frame_rate_denominator,pv.audio_language,pv.rating,pv.style_direction,pv.quality_policy,pv.cost_ceiling_micros
from projects p join project_versions pv on pv.project_id=p.id and pv.version=p.current_version`

type scanner interface{ Scan(...any) error }

func scanProject(row scanner, p *domain.Project) error {
	return row.Scan(&p.ID, &p.WorkspaceID, &p.Slug, &p.Status, &p.Privacy, &p.CurrentVersion, &p.CreatedBy, &p.DeletedAt, &p.PurgeAt, &p.CreatedAt, &p.UpdatedAt, &p.Settings.Name, &p.Settings.ProjectType, &p.Settings.ProductionFormat, &p.Settings.AspectWidth, &p.Settings.AspectHeight, &p.Settings.FrameRateNumerator, &p.Settings.FrameRateDenominator, &p.Settings.AudioLanguage, &p.Settings.Rating, &p.Settings.StyleDirection, &p.Settings.QualityPolicy, &p.Settings.CostCeilingMicros)
}

func (s *Store) CreateProject(ctx context.Context, userID, workspaceID string, settings domain.ProjectSettings) (domain.Project, error) {
	var p domain.Project
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, workspaceID, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		id := uuid.Must(uuid.NewV7()).String()
		slug, err := uniqueProjectSlug(ctx, tx, workspaceID, settings.Name)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `insert into projects(id,workspace_id,slug,created_by) values($1,$2,$3,$4)`, id, workspaceID, slug, userID); err != nil {
			return err
		}
		if err = insertProjectVersion(ctx, tx, id, workspaceID, 1, userID, settings); err != nil {
			return err
		}
		if err = audit(ctx, tx, workspaceID, userID, "project.created", "project", id, map[string]any{"name": settings.Name}); err != nil {
			return err
		}
		return scanProject(tx.QueryRow(ctx, projectSelect+` where p.id=$1`, id), &p)
	})
	return p, err
}

func (s *Store) GetProject(ctx context.Context, userID, projectID string) (domain.Project, error) {
	var p domain.Project
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		return scanProject(tx.QueryRow(ctx, projectSelect+` where p.id=$1`, projectID), &p)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return p, err
}

func (s *Store) ListProjectVersions(ctx context.Context, userID, projectID string) ([]domain.ProjectVersion, error) {
	out := []domain.ProjectVersion{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid string
		if err := tx.QueryRow(ctx, `select workspace_id::text from projects where id=$1`, projectID).Scan(&wid); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor, domain.RoleReviewer, domain.RoleViewer); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `select id::text,project_id::text,workspace_id::text,version,name,project_type,production_format,aspect_width,aspect_height,frame_rate_numerator,frame_rate_denominator,audio_language,rating,style_direction,quality_policy,cost_ceiling_micros,created_by::text,created_at
from project_versions where project_id=$1 order by version desc`, projectID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var v domain.ProjectVersion
			if err := rows.Scan(&v.ID, &v.ProjectID, &v.WorkspaceID, &v.Version, &v.Settings.Name, &v.Settings.ProjectType, &v.Settings.ProductionFormat, &v.Settings.AspectWidth, &v.Settings.AspectHeight, &v.Settings.FrameRateNumerator, &v.Settings.FrameRateDenominator, &v.Settings.AudioLanguage, &v.Settings.Rating, &v.Settings.StyleDirection, &v.Settings.QualityPolicy, &v.Settings.CostCeilingMicros, &v.CreatedBy, &v.CreatedAt); err != nil {
				return err
			}
			out = append(out, v)
		}
		return rows.Err()
	})
	return out, err
}

func (s *Store) UpdateProject(ctx context.Context, userID, projectID string, expected int64, settings domain.ProjectSettings) (domain.Project, error) {
	var p domain.Project
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid string
		var current int64
		if err := tx.QueryRow(ctx, `select workspace_id::text,current_version from projects where id=$1 for update`, projectID).Scan(&wid, &current); err != nil {
			return err
		}
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		if current != expected {
			return ErrConflict
		}
		next := current + 1
		if err := insertProjectVersion(ctx, tx, projectID, wid, next, userID, settings); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update projects set current_version=$1,updated_at=now() where id=$2`, next, projectID); err != nil {
			return err
		}
		if err := audit(ctx, tx, wid, userID, "project.updated", "project", projectID, map[string]any{"version": next}); err != nil {
			return err
		}
		return scanProject(tx.QueryRow(ctx, projectSelect+` where p.id=$1`, projectID), &p)
	})
	return p, err
}

func insertProjectVersion(ctx context.Context, tx pgx.Tx, id, wid string, version int64, userID string, s domain.ProjectSettings) error {
	_, err := tx.Exec(ctx, `insert into project_versions(id,project_id,workspace_id,version,name,project_type,production_format,aspect_width,aspect_height,frame_rate_numerator,frame_rate_denominator,audio_language,rating,style_direction,quality_policy,cost_ceiling_micros,created_by)
values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`, uuid.Must(uuid.NewV7()).String(), id, wid, version, s.Name, s.ProjectType, s.ProductionFormat, s.AspectWidth, s.AspectHeight, s.FrameRateNumerator, s.FrameRateDenominator, s.AudioLanguage, s.Rating, s.StyleDirection, s.QualityPolicy, s.CostCeilingMicros, userID)
	return err
}

func (s *Store) SetProjectStatus(ctx context.Context, userID, projectID, status string) (domain.Project, error) {
	var p domain.Project
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid string
		if err := tx.QueryRow(ctx, `select workspace_id::text from projects where id=$1 for update`, projectID).Scan(&wid); err != nil {
			return err
		}
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		var deleted, purge any
		if status == "trashed" {
			deleted = "now()"
			purge = "now() + interval '30 days'"
		} else {
			deleted = nil
			purge = nil
		}
		if status == "trashed" {
			_, _ = tx.Exec(ctx, `update projects set status='trashed',deleted_at=now(),purge_at=now()+interval '30 days',updated_at=now() where id=$1`, projectID)
		} else {
			_, _ = tx.Exec(ctx, `update projects set status=$2,deleted_at=null,purge_at=null,updated_at=now() where id=$1`, projectID, status)
		}
		_ = deleted
		_ = purge
		if err := audit(ctx, tx, wid, userID, "project."+status, "project", projectID, nil); err != nil {
			return err
		}
		return scanProject(tx.QueryRow(ctx, projectSelect+` where p.id=$1`, projectID), &p)
	})
	return p, err
}

func (s *Store) DuplicateProject(ctx context.Context, userID, projectID, name string) (domain.Project, error) {
	source, err := s.GetProject(ctx, userID, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	source.Settings.Name = name
	created, err := s.CreateProject(ctx, userID, source.WorkspaceID, source.Settings)
	if err != nil {
		return domain.Project{}, err
	}
	err = s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `insert into project_library_links(id,workspace_id,project_id,entity_id,entity_version_id,created_by) select gen_random_uuid(),workspace_id,$1,entity_id,entity_version_id,$2 from project_library_links where project_id=$3`, created.ID, userID, projectID); err != nil {
			return err
		}
		return audit(ctx, tx, source.WorkspaceID, userID, "project.duplicated", "project", created.ID, map[string]any{"sourceProjectId": projectID})
	})
	return created, err
}

func (s *Store) GetBudget(ctx context.Context, userID, wid string) (domain.BudgetPolicy, error) {
	var b domain.BudgetPolicy
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor, domain.RoleReviewer, domain.RoleViewer); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `select workspace_id::text,currency,monthly_limit_micros,per_run_approval_micros,editor_can_publish_library,version,updated_at from workspace_policies where workspace_id=$1`, wid).Scan(&b.WorkspaceID, &b.Currency, &b.MonthlyLimitMicros, &b.PerRunApprovalMicros, &b.EditorCanPublishLibrary, &b.Version, &b.UpdatedAt)
	})
	return b, err
}

func (s *Store) UpdateBudget(ctx context.Context, userID, wid string, expected int64, b domain.BudgetPolicy) (domain.BudgetPolicy, error) {
	var out domain.BudgetPolicy
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `update workspace_policies set currency=$1,monthly_limit_micros=$2,per_run_approval_micros=$3,editor_can_publish_library=$4,version=version+1,updated_by=$5,updated_at=now() where workspace_id=$6 and version=$7`, b.Currency, b.MonthlyLimitMicros, b.PerRunApprovalMicros, b.EditorCanPublishLibrary, userID, wid, expected)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return ErrConflict
		}
		if err := audit(ctx, tx, wid, userID, "budget.updated", "workspace", wid, nil); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `select workspace_id::text,currency,monthly_limit_micros,per_run_approval_micros,editor_can_publish_library,version,updated_at from workspace_policies where workspace_id=$1`, wid).Scan(&out.WorkspaceID, &out.Currency, &out.MonthlyLimitMicros, &out.PerRunApprovalMicros, &out.EditorCanPublishLibrary, &out.Version, &out.UpdatedAt)
	})
	return out, err
}

func (s *Store) CreateSession(ctx context.Context, userID string, token string, device, userAgent, ipPrefix string, expires time.Time) (string, error) {
	id := uuid.Must(uuid.NewV7()).String()
	sum := sha256.Sum256([]byte(token))
	_, err := s.pool.Exec(ctx, `insert into auth_sessions(id,user_id,token_hash,device_label,user_agent,ip_prefix,expires_at) values($1,$2,$3,$4,$5,$6,$7)`, id, userID, sum[:], device, userAgent, ipPrefix, expires)
	return id, err
}
func (s *Store) ValidateSession(ctx context.Context, userID, token string) error {
	sum := sha256.Sum256([]byte(token))
	tag, err := s.pool.Exec(ctx, `update auth_sessions set last_seen_at=now() where user_id=$1 and token_hash=$2 and revoked_at is null and expires_at>now()`, userID, sum[:])
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrForbidden
	}
	return nil
}
func (s *Store) RevokeSession(ctx context.Context, userID, sessionID string) error {
	tag, err := s.pool.Exec(ctx, `update auth_sessions set revoked_at=now() where id=$1 and user_id=$2 and revoked_at is null`, sessionID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
func (s *Store) RevokeSessionByToken(ctx context.Context, userID, token string) error {
	sum := sha256.Sum256([]byte(token))
	tag, err := s.pool.Exec(ctx, `update auth_sessions set revoked_at=now() where user_id=$1 and token_hash=$2 and revoked_at is null`, userID, sum[:])
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
func (s *Store) RevokeAllSessions(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `update auth_sessions set revoked_at=now() where user_id=$1 and revoked_at is null`, userID)
	return err
}

func requireRole(ctx context.Context, tx pgx.Tx, userID, wid string, allowed ...domain.Role) (domain.Membership, error) {
	var m domain.Membership
	err := tx.QueryRow(ctx, `select id::text,workspace_id::text,user_id::text,role::text,status::text,can_spend,monthly_limit_micros,per_run_limit_micros,library_publish,created_at,updated_at,removed_at from memberships where workspace_id=$1 and user_id=$2 and status='active'`, wid, userID).Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &m.Status, &m.CanSpend, &m.MonthlyLimitMicros, &m.PerRunLimitMicros, &m.LibraryPublish, &m.CreatedAt, &m.UpdatedAt, &m.RemovedAt)
	if err != nil {
		return m, ErrForbidden
	}
	for _, r := range allowed {
		if m.Role == r {
			return m, nil
		}
	}
	return m, ErrForbidden
}

func audit(ctx context.Context, tx pgx.Tx, wid, userID, action, targetType, targetID string, source map[string]any) error {
	if source == nil {
		source = map[string]any{}
	}
	b, _ := json.Marshal(source)
	_, err := tx.Exec(ctx, `insert into audit_events(id,workspace_id,actor_user_id,action,target_type,target_id,source_context,correlation_id) values($1,nullif($2,'')::uuid,nullif($3,'')::uuid,$4,$5,nullif($6,'')::uuid,$7,$8)`, uuid.Must(uuid.NewV7()).String(), wid, userID, action, targetType, targetID, b, uuid.NewString())
	return err
}

func uniqueSlug(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	base := slugify(name)
	if base == "" {
		base = "workspace"
	}
	for i := 0; i < 100; i++ {
		v := base
		if i > 0 {
			v = fmt.Sprintf("%s-%d", base, i+1)
		}
		var exists bool
		if err := tx.QueryRow(ctx, `select exists(select 1 from workspaces where slug=$1)`, v).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return v, nil
		}
	}
	return "", ErrConflict
}
func uniqueProjectSlug(ctx context.Context, tx pgx.Tx, wid, name string) (string, error) {
	base := slugify(name)
	if base == "" {
		base = "project"
	}
	for i := 0; i < 100; i++ {
		v := base
		if i > 0 {
			v = fmt.Sprintf("%s-%d", base, i+1)
		}
		var exists bool
		if err := tx.QueryRow(ctx, `select exists(select 1 from projects where workspace_id=$1 and slug=$2)`, wid, v).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return v, nil
		}
	}
	return "", ErrConflict
}
func slugify(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = nonSlug.ReplaceAllString(v, "-")
	v = strings.Trim(v, "-")
	if len(v) > 55 {
		v = strings.Trim(v[:55], "-")
	}
	return v
}
func fallbackName(name, email string) string {
	if strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return "Creator"
}
