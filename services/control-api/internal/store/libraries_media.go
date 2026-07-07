package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ats-tech/cineforge/services/control-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateLibrary(ctx context.Context, userID, wid, name, kind string) (domain.Library, error) {
	var l domain.Library
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		id := uuid.Must(uuid.NewV7()).String()
		return tx.QueryRow(ctx, `insert into libraries(id,workspace_id,name,type,created_by) values($1,$2,$3,$4,$5) returning id::text,workspace_id::text,name,type,version,archived_at,created_at,updated_at`, id, wid, name, kind, userID).Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Type, &l.Version, &l.ArchivedAt, &l.CreatedAt, &l.UpdatedAt)
	})
	return l, err
}

func (s *Store) ListLibraries(ctx context.Context, userID, wid string) ([]domain.Library, error) {
	out := []domain.Library{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor, domain.RoleReviewer, domain.RoleViewer); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `select id::text,workspace_id::text,name,type,version,archived_at,created_at,updated_at from libraries where workspace_id=$1 order by updated_at desc`, wid)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var l domain.Library
			if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Type, &l.Version, &l.ArchivedAt, &l.CreatedAt, &l.UpdatedAt); err != nil {
				return err
			}
			out = append(out, l)
		}
		return rows.Err()
	})
	return out, err
}

type EntityInput struct {
	Name, Type, Summary, Description, ChangeSummary string
	Aliases, Tags                                   []string
	Attributes                                      map[string]any
}

func (s *Store) CreateLibraryEntity(ctx context.Context, userID, libraryID string, in EntityInput) (domain.LibraryEntity, error) {
	var out domain.LibraryEntity
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid string
		if err := tx.QueryRow(ctx, `select workspace_id::text from libraries where id=$1 and archived_at is null`, libraryID).Scan(&wid); err != nil {
			return err
		}
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		eid := uuid.Must(uuid.NewV7()).String()
		vid := uuid.Must(uuid.NewV7()).String()
		attrs, _ := json.Marshal(in.Attributes)
		if _, err := tx.Exec(ctx, `insert into library_entities(id,workspace_id,library_id,created_by) values($1,$2,$3,$4)`, eid, wid, libraryID, userID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into library_entity_versions(id,workspace_id,entity_id,version,type,name,aliases,summary,description,tags,attributes,status,change_summary,author_user_id) values($1,$2,$3,1,$4,$5,$6,$7,$8,$9,$10,'draft',$11,$12)`, vid, wid, eid, in.Type, in.Name, in.Aliases, in.Summary, in.Description, in.Tags, attrs, in.ChangeSummary, userID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update library_entities set current_version_id=$1 where id=$2`, vid, eid); err != nil {
			return err
		}
		if err := audit(ctx, tx, wid, userID, "library_entity.created", "library_entity", eid, nil); err != nil {
			return err
		}
		return loadEntity(ctx, tx, eid, &out)
	})
	return out, err
}

func (s *Store) UpdateLibraryEntity(ctx context.Context, userID, entityID string, in EntityInput) (domain.LibraryEntity, error) {
	var out domain.LibraryEntity
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid, currentID string
		var current int64
		if err := tx.QueryRow(ctx, `select e.workspace_id::text,e.current_version_id::text,v.version from library_entities e join library_entity_versions v on v.id=e.current_version_id where e.id=$1 for update`, entityID).Scan(&wid, &currentID, &current); err != nil {
			return err
		}
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		vid := uuid.Must(uuid.NewV7()).String()
		attrs, _ := json.Marshal(in.Attributes)
		if _, err := tx.Exec(ctx, `insert into library_entity_versions(id,workspace_id,entity_id,version,type,name,aliases,summary,description,tags,attributes,status,change_summary,author_user_id,source_version_id) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'draft',$12,$13,$14)`, vid, wid, entityID, current+1, in.Type, in.Name, in.Aliases, in.Summary, in.Description, in.Tags, attrs, in.ChangeSummary, userID, currentID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update library_entities set current_version_id=$1,updated_at=now() where id=$2`, vid, entityID); err != nil {
			return err
		}
		return loadEntity(ctx, tx, entityID, &out)
	})
	return out, err
}

func (s *Store) SetLibraryEntityStatus(ctx context.Context, userID, entityID, status string) (domain.LibraryEntity, error) {
	var out domain.LibraryEntity
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid string
		if err := tx.QueryRow(ctx, `select workspace_id::text from library_entities where id=$1`, entityID).Scan(&wid); err != nil {
			return err
		}
		m, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor)
		if err != nil {
			return err
		}
		if m.Role == domain.RoleEditor && !m.LibraryPublish {
			var allowed bool
			if err := tx.QueryRow(ctx, `select editor_can_publish_library from workspace_policies where workspace_id=$1`, wid).Scan(&allowed); err != nil || !allowed {
				return ErrForbidden
			}
		}
		if _, err := tx.Exec(ctx, `update library_entity_versions set status=$1 where id=(select current_version_id from library_entities where id=$2)`, status, entityID); err != nil {
			return err
		}
		if err := audit(ctx, tx, wid, userID, "library_entity."+status, "library_entity", entityID, nil); err != nil {
			return err
		}
		return loadEntity(ctx, tx, entityID, &out)
	})
	return out, err
}

func (s *Store) ListLibraryEntities(ctx context.Context, userID, libraryID, query, kind, status string) ([]domain.LibraryEntity, error) {
	out := []domain.LibraryEntity{}
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid string
		if err := tx.QueryRow(ctx, `select workspace_id::text from libraries where id=$1`, libraryID).Scan(&wid); err != nil {
			return err
		}
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor, domain.RoleReviewer, domain.RoleViewer); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, entitySelect+` where e.library_id=$1 and e.archived_at is null and ($2='' or v.search_vector @@ plainto_tsquery('simple',$2)) and ($3='' or v.type=$3) and ($4='' or v.status::text=$4) order by e.updated_at desc limit 100`, libraryID, query, kind, status)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var e domain.LibraryEntity
			if err := scanEntity(rows, &e); err != nil {
				return err
			}
			out = append(out, e)
		}
		return rows.Err()
	})
	return out, err
}

const entitySelect = `select e.id::text,e.workspace_id::text,e.library_id::text,e.current_version_id::text,e.archived_at,e.created_at,e.updated_at,v.id::text,v.entity_id::text,v.version,v.type,v.name,v.aliases,v.summary,v.description,v.tags,v.attributes,v.status::text,v.change_summary,v.author_user_id::text,v.source_version_id::text,v.created_at from library_entities e join library_entity_versions v on v.id=e.current_version_id`

func scanEntity(row scanner, e *domain.LibraryEntity) error {
	var attrs []byte
	err := row.Scan(&e.ID, &e.WorkspaceID, &e.LibraryID, &e.CurrentVersionID, &e.ArchivedAt, &e.CreatedAt, &e.UpdatedAt, &e.CurrentVersion.ID, &e.CurrentVersion.EntityID, &e.CurrentVersion.Version, &e.CurrentVersion.Type, &e.CurrentVersion.Name, &e.CurrentVersion.Aliases, &e.CurrentVersion.Summary, &e.CurrentVersion.Description, &e.CurrentVersion.Tags, &attrs, &e.CurrentVersion.Status, &e.CurrentVersion.ChangeSummary, &e.CurrentVersion.AuthorUserID, &e.CurrentVersion.SourceVersionID, &e.CurrentVersion.CreatedAt)
	if err == nil {
		_ = json.Unmarshal(attrs, &e.CurrentVersion.Attributes)
	}
	return err
}
func loadEntity(ctx context.Context, tx pgx.Tx, id string, out *domain.LibraryEntity) error {
	return scanEntity(tx.QueryRow(ctx, entitySelect+` where e.id=$1`, id), out)
}

func (s *Store) LinkEntityToProject(ctx context.Context, userID, projectID, entityVersionID string) error {
	return s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid, entityID, status string
		if err := tx.QueryRow(ctx, `select p.workspace_id::text,v.entity_id::text,v.status::text from projects p join library_entity_versions v on v.workspace_id=p.workspace_id where p.id=$1 and v.id=$2`, projectID, entityVersionID).Scan(&wid, &entityID, &status); err != nil {
			return err
		}
		if status != "approved" {
			return ErrConflict
		}
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `insert into project_library_links(id,workspace_id,project_id,entity_id,entity_version_id,created_by) values($1,$2,$3,$4,$5,$6) on conflict(project_id,entity_id) do update set entity_version_id=excluded.entity_version_id,updated_at=now()`, uuid.Must(uuid.NewV7()).String(), wid, projectID, entityID, entityVersionID, userID)
		return err
	})
}

type UploadInput struct {
	Kind, Filename, ContentType string
	SizeBytes                   int64
}

type MediaUploadTarget struct {
	Object, ContentType string
	SizeBytes           int64
}

func (s *Store) MediaUploadTarget(ctx context.Context, userID, assetID string) (MediaUploadTarget, error) {
	var out MediaUploadTarget
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid string
		if err := tx.QueryRow(ctx, `select workspace_id::text,quarantine_object,content_type,size_bytes from media_assets where id=$1 and state='initiated'`, assetID).Scan(&wid, &out.Object, &out.ContentType, &out.SizeBytes); err != nil {
			return err
		}
		_, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor)
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return out, err
}

func (s *Store) CreateMediaUpload(ctx context.Context, userID, wid string, in UploadInput) (domain.MediaAsset, error) {
	var a domain.MediaAsset
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		if err := validateUpload(in); err != nil {
			return err
		}
		id := uuid.Must(uuid.NewV7()).String()
		object := "quarantine/" + wid + "/" + id
		return tx.QueryRow(ctx, `insert into media_assets(id,workspace_id,kind,state,original_filename,content_type,size_bytes,quarantine_object,created_by) values($1,$2,$3,'initiated',$4,$5,$6,$7,$8) returning id::text,workspace_id::text,kind,state::text,original_filename,content_type,size_bytes,null::text,'{}'::jsonb,null::text,created_at,updated_at`, id, wid, in.Kind, in.Filename, in.ContentType, in.SizeBytes, object, userID).Scan(&a.ID, &a.WorkspaceID, &a.Kind, &a.State, &a.OriginalFilename, &a.ContentType, &a.SizeBytes, &a.ChecksumSHA256, &a.Metadata, &a.RejectionCode, &a.CreatedAt, &a.UpdatedAt)
	})
	return a, err
}

func validateUpload(in UploadInput) error {
	limits := map[string]int64{"image": 50 << 20, "video": 2 << 30, "audio": 500 << 20, "voice": 500 << 20, "3d": 250 << 20}
	limit, ok := limits[in.Kind]
	if !ok || in.SizeBytes <= 0 || in.SizeBytes > limit {
		return ErrConflict
	}
	allowed := map[string]map[string]bool{"image": {"image/jpeg": true, "image/png": true, "image/webp": true}, "video": {"video/mp4": true, "video/quicktime": true, "video/webm": true}, "audio": {"audio/wav": true, "audio/mpeg": true, "audio/mp4": true, "audio/flac": true, "audio/ogg": true}, "voice": {"audio/wav": true, "audio/mpeg": true, "audio/mp4": true, "audio/flac": true, "audio/ogg": true}, "3d": {"model/gltf-binary": true, "application/octet-stream": true}}
	if !allowed[in.Kind][in.ContentType] {
		return ErrConflict
	}
	return nil
}

func (s *Store) CompleteMediaUpload(ctx context.Context, userID, assetID string) error {
	return s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var wid string
		if err := tx.QueryRow(ctx, `select workspace_id::text from media_assets where id=$1 and state='initiated'`, assetID).Scan(&wid); err != nil {
			return err
		}
		if _, err := requireRole(ctx, tx, userID, wid, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `update media_assets set state='uploaded',updated_at=now() where id=$1`, assetID)
		return err
	})
}

func (s *Store) PurgeDue(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `delete from projects where purge_at<=now(); delete from workspaces where purge_at<=now(); update users set email='deleted+'||id||'@invalid.local',display_name='Deleted user',avatar_url=null,deleted_at=now(),deletion_scheduled_at=null where deletion_scheduled_at<=now();`)
	return err
}

var _ = errors.Is
var _ = time.Now
