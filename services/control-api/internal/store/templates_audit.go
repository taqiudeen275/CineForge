package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ats-tech/cineforge/services/control-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListProjectTemplates(ctx context.Context, userID, workspaceID string) ([]domain.ProjectTemplate, error) {
	out := []domain.ProjectTemplate{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if workspaceID != "" {
			if _, err := requireRole(ctx, tx, userID, workspaceID, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor, domain.RoleReviewer, domain.RoleViewer); err != nil {
				return err
			}
		}
		rows, err := tx.Query(ctx, `select t.id::text,t.workspace_id::text,t.key,t.name,t.description,t.system,t.current_version,v.settings,v.library_links,t.created_at,t.updated_at from project_templates t join project_template_versions v on v.template_id=t.id and v.version=t.current_version where t.archived_at is null and (t.system or t.workspace_id=$1::uuid) order by t.system desc,t.name`, nilUUID(workspaceID))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item domain.ProjectTemplate
			var settings, links []byte
			if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.Key, &item.Name, &item.Description, &item.System, &item.CurrentVersion, &settings, &links, &item.CreatedAt, &item.UpdatedAt); err != nil {
				return err
			}
			_ = json.Unmarshal(settings, &item.Settings)
			_ = json.Unmarshal(links, &item.LibraryLinks)
			out = append(out, item)
		}
		return rows.Err()
	})
	return out, err
}

func (s *Store) SaveProjectTemplate(ctx context.Context, userID, workspaceID, projectID, templateID, name, description string, expected int64) (domain.ProjectTemplate, error) {
	var out domain.ProjectTemplate
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, workspaceID, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		var project domain.Project
		if err := scanProject(tx.QueryRow(ctx, projectSelect+` where p.id=$1 and p.workspace_id=$2`, projectID, workspaceID), &project); err != nil {
			return err
		}
		settings, _ := json.Marshal(project.Settings)
		var links []string
		rows, err := tx.Query(ctx, `select entity_version_id::text from project_library_links where project_id=$1 order by entity_id`, projectID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			links = append(links, id)
		}
		rows.Close()
		linkJSON, _ := json.Marshal(links)
		if templateID == "" {
			templateID = uuid.Must(uuid.NewV7()).String()
			if _, err := tx.Exec(ctx, `insert into project_templates(id,workspace_id,name,description,created_by) values($1,$2,$3,$4,$5)`, templateID, workspaceID, name, description, userID); err != nil {
				return err
			}
			expected = 0
		} else {
			var current int64
			if err := tx.QueryRow(ctx, `select current_version from project_templates where id=$1 and workspace_id=$2 and not system for update`, templateID, workspaceID).Scan(&current); err != nil {
				return err
			}
			if current != expected {
				return ErrConflict
			}
			if _, err := tx.Exec(ctx, `update project_templates set name=$1,description=$2,current_version=current_version+1,updated_at=now() where id=$3`, name, description, templateID); err != nil {
				return err
			}
		}
		next := expected + 1
		if _, err := tx.Exec(ctx, `insert into project_template_versions(id,template_id,version,settings,library_links,created_by) values($1,$2,$3,$4,$5,$6)`, uuid.Must(uuid.NewV7()).String(), templateID, next, settings, linkJSON, userID); err != nil {
			return err
		}
		if err := audit(ctx, tx, workspaceID, userID, "project_template.saved", "project_template", templateID, map[string]any{"version": next}); err != nil {
			return err
		}
		items, err := templateByID(ctx, tx, templateID)
		if err != nil {
			return err
		}
		out = items
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return out, err
}

func templateByID(ctx context.Context, tx pgx.Tx, id string) (domain.ProjectTemplate, error) {
	var item domain.ProjectTemplate
	var settings, links []byte
	err := tx.QueryRow(ctx, `select t.id::text,t.workspace_id::text,t.key,t.name,t.description,t.system,t.current_version,v.settings,v.library_links,t.created_at,t.updated_at from project_templates t join project_template_versions v on v.template_id=t.id and v.version=t.current_version where t.id=$1`, id).Scan(&item.ID, &item.WorkspaceID, &item.Key, &item.Name, &item.Description, &item.System, &item.CurrentVersion, &settings, &links, &item.CreatedAt, &item.UpdatedAt)
	_ = json.Unmarshal(settings, &item.Settings)
	_ = json.Unmarshal(links, &item.LibraryLinks)
	return item, err
}

func (s *Store) ListAuditEvents(ctx context.Context, userID, workspaceID string) ([]domain.AuditEvent, error) {
	out := []domain.AuditEvent{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, workspaceID, domain.RoleOwner, domain.RoleAdmin); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `select id::text,workspace_id::text,actor_user_id::text,action,target_type,target_id::text,decision,source_context,correlation_id,occurred_at from audit_events where workspace_id=$1 order by occurred_at desc,id desc limit 100`, workspaceID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item domain.AuditEvent
			var source []byte
			if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.ActorUserID, &item.Action, &item.TargetType, &item.TargetID, &item.Decision, &source, &item.CorrelationID, &item.OccurredAt); err != nil {
				return err
			}
			_ = json.Unmarshal(source, &item.SourceContext)
			out = append(out, item)
		}
		return rows.Err()
	})
	return out, err
}

func nilUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}
