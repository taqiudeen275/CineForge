package store

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/ats-tech/cineforge/services/control-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionView struct {
	ID, DeviceLabel                  string
	UserAgent, IPPrefix              *string
	CreatedAt, LastSeenAt, ExpiresAt time.Time
}

func (s *Store) GetUser(ctx context.Context, userID string) (domain.User, error) {
	var u domain.User
	var personalID *string
	err := s.pool.QueryRow(ctx, `select id::text,identity_provider_uid,email,email_verified,display_name,avatar_url,locale,timezone,personal_workspace_id::text,deletion_scheduled_at,version,created_at,updated_at from users where id=$1 and deleted_at is null`, userID).
		Scan(&u.ID, &u.IdentityProviderUID, &u.Email, &u.EmailVerified, &u.DisplayName, &u.AvatarURL, &u.Locale, &u.Timezone, &personalID, &u.DeletionScheduledAt, &u.Version, &u.CreatedAt, &u.UpdatedAt)
	u.PersonalWorkspaceID = personalID
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return u, err
}

func (s *Store) GetUserByIdentityUID(ctx context.Context, uid string) (domain.User, error) {
	var id string
	if err := s.pool.QueryRow(ctx, `select id::text from users where identity_provider_uid=$1 and deleted_at is null`, uid).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return s.GetUser(ctx, id)
}

func (s *Store) UpdateUser(ctx context.Context, userID string, expected int64, displayName string, avatar *string, locale, timezone string) (domain.User, error) {
	tag, err := s.pool.Exec(ctx, `update users set display_name=$1,avatar_url=$2,locale=$3,timezone=$4,version=version+1,updated_at=now() where id=$5 and version=$6 and deletion_scheduled_at is null`, displayName, avatar, locale, timezone, userID, expected)
	if err != nil {
		return domain.User{}, err
	}
	if tag.RowsAffected() != 1 {
		return domain.User{}, ErrConflict
	}
	return s.GetUser(ctx, userID)
}

func (s *Store) ListSessions(ctx context.Context, userID string) ([]SessionView, error) {
	rows, err := s.pool.Query(ctx, `select id::text,device_label,user_agent,ip_prefix,created_at,last_seen_at,expires_at from auth_sessions where user_id=$1 and revoked_at is null and expires_at>now() order by last_seen_at desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SessionView{}
	for rows.Next() {
		var v SessionView
		if err := rows.Scan(&v.ID, &v.DeviceLabel, &v.UserAgent, &v.IPPrefix, &v.CreatedAt, &v.LastSeenAt, &v.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) ScheduleAccountDeletion(ctx context.Context, userID string) error {
	return s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var blocked bool
		if err := tx.QueryRow(ctx, `select exists(select 1 from workspaces w where w.owner_user_id=$1 and w.deleted_at is null and exists(select 1 from memberships m where m.workspace_id=w.id and m.status='active' and m.user_id<>$1))`, userID).Scan(&blocked); err != nil {
			return err
		}
		if blocked {
			return ErrOwnerBlocked
		}
		if _, err := tx.Exec(ctx, `update users set deletion_scheduled_at=now()+interval '30 days',updated_at=now(),version=version+1 where id=$1`, userID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update auth_sessions set revoked_at=now() where user_id=$1 and revoked_at is null`, userID); err != nil {
			return err
		}
		return audit(ctx, tx, "", userID, "account.deletion_scheduled", "user", userID, nil)
	})
}

func (s *Store) CancelAccountDeletion(ctx context.Context, userID string) error {
	tag, err := s.pool.Exec(ctx, `update users set deletion_scheduled_at=null,updated_at=now(),version=version+1 where id=$1 and deletion_scheduled_at>now()`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateTeamWorkspace(ctx context.Context, userID, name string) (domain.Workspace, error) {
	var ws domain.Workspace
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		wid := uuid.Must(uuid.NewV7()).String()
		slug, err := uniqueSlug(ctx, tx, name)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `insert into workspaces(id,name,slug,kind,owner_user_id) values($1,$2,$3,'team',$4)`, wid, name, slug, userID); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `insert into memberships(id,workspace_id,user_id,role,status,can_spend,library_publish) values($1,$2,$3,'owner','active',true,true)`, uuid.Must(uuid.NewV7()).String(), wid, userID); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `insert into workspace_policies(workspace_id,updated_by) values($1,$2)`, wid, userID); err != nil {
			return err
		}
		if err = audit(ctx, tx, wid, userID, "workspace.created", "workspace", wid, map[string]any{"kind": "team"}); err != nil {
			return err
		}
		return scanWorkspace(tx.QueryRow(ctx, workspaceSelect+` where id=$1`, wid), &ws)
	})
	return ws, err
}

func (s *Store) UpdateWorkspace(ctx context.Context, userID, wid, name string, expected int64) (domain.Workspace, error) {
	var ws domain.Workspace
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `update workspaces set name=$1,version=version+1,updated_at=now() where id=$2 and version=$3 and deleted_at is null`, name, wid, expected)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return ErrConflict
		}
		if err := audit(ctx, tx, wid, userID, "workspace.updated", "workspace", wid, map[string]any{"name": name}); err != nil {
			return err
		}
		return scanWorkspace(tx.QueryRow(ctx, workspaceSelect+` where id=$1`, wid), &ws)
	})
	return ws, err
}

func (s *Store) ListMembers(ctx context.Context, userID, wid string) ([]domain.Membership, error) {
	out := []domain.Membership{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor, domain.RoleReviewer, domain.RoleViewer); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `select m.id::text,m.workspace_id::text,m.user_id::text,u.email,u.display_name,m.role::text,m.status::text,m.can_spend,m.monthly_limit_micros,m.per_run_limit_micros,m.library_publish,m.created_at,m.updated_at,m.removed_at from memberships m join users u on u.id=m.user_id where m.workspace_id=$1 and m.status<>'removed' order by m.created_at`, wid)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var m domain.Membership
			if err := rows.Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Email, &m.DisplayName, &m.Role, &m.Status, &m.CanSpend, &m.MonthlyLimitMicros, &m.PerRunLimitMicros, &m.LibraryPublish, &m.CreatedAt, &m.UpdatedAt, &m.RemovedAt); err != nil {
				return err
			}
			out = append(out, m)
		}
		return rows.Err()
	})
	return out, err
}

func (s *Store) ListInvitations(ctx context.Context, userID, wid string) ([]domain.Invitation, error) {
	out := []domain.Invitation{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `select id::text,workspace_id::text,email,role::text,expires_at,created_at from invitations where workspace_id=$1 and accepted_at is null and revoked_at is null order by created_at desc`, wid)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var v domain.Invitation
			if err := rows.Scan(&v.ID, &v.WorkspaceID, &v.Email, &v.Role, &v.ExpiresAt, &v.CreatedAt); err != nil {
				return err
			}
			out = append(out, v)
		}
		return rows.Err()
	})
	return out, err
}

func (s *Store) RevokeInvitation(ctx context.Context, userID, wid, id string) error {
	return s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `update invitations set revoked_at=now() where id=$1 and workspace_id=$2 and accepted_at is null and revoked_at is null`, id, wid)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return ErrNotFound
		}
		return audit(ctx, tx, wid, userID, "invitation.revoked", "invitation", id, nil)
	})
}

func (s *Store) CreateInvitation(ctx context.Context, userID, wid, email string, role domain.Role, token string, expires time.Time) (string, error) {
	id := uuid.Must(uuid.NewV7()).String()
	sum := sha256.Sum256([]byte(token))
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		m, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin)
		if err != nil {
			return err
		}
		if role == domain.RoleOwner || (role == domain.RoleAdmin && m.Role != domain.RoleOwner) {
			return ErrForbidden
		}
		if _, err := tx.Exec(ctx, `insert into invitations(id,workspace_id,email,role,token_hash,invited_by,expires_at) values($1,$2,lower($3),$4,$5,$6,$7)`, id, wid, email, role, sum[:], userID, expires); err != nil {
			return err
		}
		return audit(ctx, tx, wid, userID, "invitation.created", "invitation", id, map[string]any{"role": role})
	})
	return id, err
}

func (s *Store) AcceptInvitation(ctx context.Context, user domain.User, token string) (domain.Workspace, error) {
	var ws domain.Workspace
	sum := sha256.Sum256([]byte(token))
	err := s.withUserTx(ctx, user.ID, func(tx pgx.Tx) error {
		var inviteID, wid, email string
		var role domain.Role
		if err := tx.QueryRow(ctx, `select id::text,workspace_id::text,email,role::text from invitations where token_hash=$1 and accepted_at is null and revoked_at is null and expires_at>now() for update`, sum[:]).Scan(&inviteID, &wid, &email, &role); err != nil {
			return err
		}
		if !stringsEqualFold(email, user.Email) {
			return ErrForbidden
		}
		if _, err := tx.Exec(ctx, `insert into memberships(id,workspace_id,user_id,role,status) values($1,$2,$3,$4,'active') on conflict(workspace_id,user_id) do update set role=excluded.role,status='active',removed_at=null,updated_at=now()`, uuid.Must(uuid.NewV7()).String(), wid, user.ID, role); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update invitations set accepted_at=now() where id=$1`, inviteID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update workspaces set kind='team',version=version+1,updated_at=now() where id=$1 and kind='personal'`, wid); err != nil {
			return err
		}
		if err := audit(ctx, tx, wid, user.ID, "invitation.accepted", "invitation", inviteID, nil); err != nil {
			return err
		}
		return scanWorkspace(tx.QueryRow(ctx, workspaceSelect+` where id=$1`, wid), &ws)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return ws, err
}

func (s *Store) UpdateMember(ctx context.Context, userID, wid, memberID string, role domain.Role, canSpend bool, monthly, perRun *int64, libraryPublish, remove bool) (domain.Membership, error) {
	var out domain.Membership
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		actor, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin)
		if err != nil {
			return err
		}
		var targetRole domain.Role
		if err := tx.QueryRow(ctx, `select role::text from memberships where id=$1 and workspace_id=$2 and status='active'`, memberID, wid).Scan(&targetRole); err != nil {
			return err
		}
		if targetRole == domain.RoleOwner || role == domain.RoleOwner {
			return ErrForbidden
		}
		if (targetRole == domain.RoleAdmin || role == domain.RoleAdmin) && actor.Role != domain.RoleOwner {
			return ErrForbidden
		}
		if actor.Role != domain.RoleOwner && (canSpend || monthly != nil || perRun != nil) {
			return ErrForbidden
		}
		if remove {
			_, err = tx.Exec(ctx, `update memberships set status='removed',removed_at=now(),updated_at=now() where id=$1 and workspace_id=$2`, memberID, wid)
		} else {
			_, err = tx.Exec(ctx, `update memberships set role=$1,can_spend=$2,monthly_limit_micros=$3,per_run_limit_micros=$4,library_publish=$5,updated_at=now() where id=$6 and workspace_id=$7`, role, canSpend, monthly, perRun, libraryPublish, memberID, wid)
		}
		if err != nil {
			return err
		}
		if err := audit(ctx, tx, wid, userID, "membership.updated", "membership", memberID, map[string]any{"role": role, "removed": remove}); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `select m.id::text,m.workspace_id::text,m.user_id::text,u.email,u.display_name,m.role::text,m.status::text,m.can_spend,m.monthly_limit_micros,m.per_run_limit_micros,m.library_publish,m.created_at,m.updated_at,m.removed_at from memberships m join users u on u.id=m.user_id where m.id=$1`, memberID).Scan(&out.ID, &out.WorkspaceID, &out.UserID, &out.Email, &out.DisplayName, &out.Role, &out.Status, &out.CanSpend, &out.MonthlyLimitMicros, &out.PerRunLimitMicros, &out.LibraryPublish, &out.CreatedAt, &out.UpdatedAt, &out.RemovedAt)
	})
	return out, err
}

func (s *Store) TransferOwnership(ctx context.Context, userID, wid, memberID string) error {
	return s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner); err != nil {
			return err
		}
		var targetID, targetUserID string
		if err := tx.QueryRow(ctx, `select id::text,user_id::text from memberships where workspace_id=$1 and id=$2 and status='active' and role<>'owner' for update`, wid, memberID).Scan(&targetID, &targetUserID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update memberships set role='admin',updated_at=now() where workspace_id=$1 and user_id=$2`, wid, userID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update memberships set role='owner',can_spend=true,updated_at=now() where id=$1`, targetID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update workspaces set owner_user_id=$1,kind='team',version=version+1,updated_at=now() where id=$2`, targetUserID, wid); err != nil {
			return err
		}
		return audit(ctx, tx, wid, userID, "workspace.ownership_transferred", "workspace", wid, map[string]any{"newOwnerUserId": targetUserID})
	})
}

func (s *Store) TrashWorkspace(ctx context.Context, userID, wid string) error {
	return s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update workspaces set deleted_at=now(),purge_at=now()+interval '30 days',version=version+1,updated_at=now() where id=$1`, wid); err != nil {
			return err
		}
		return audit(ctx, tx, wid, userID, "workspace.trashed", "workspace", wid, nil)
	})
}
func (s *Store) RestoreWorkspace(ctx context.Context, userID, wid string) error {
	return s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update workspaces set deleted_at=null,purge_at=null,version=version+1,updated_at=now() where id=$1 and purge_at>now()`, wid); err != nil {
			return err
		}
		return audit(ctx, tx, wid, userID, "workspace.restored", "workspace", wid, nil)
	})
}

func stringsEqualFold(a, b string) bool { return len(a) == len(b) && (a == b || equalFold(a, b)) }
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
