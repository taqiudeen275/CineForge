CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE workspace_kind AS ENUM ('personal', 'team');
CREATE TYPE membership_role AS ENUM ('owner', 'admin', 'editor', 'reviewer', 'viewer');
CREATE TYPE membership_status AS ENUM ('invited', 'active', 'suspended', 'removed');
CREATE TYPE project_status AS ENUM ('active', 'archived', 'trashed');
CREATE TYPE entity_status AS ENUM ('draft', 'approved', 'retired');
CREATE TYPE media_state AS ENUM ('initiated', 'uploaded', 'scanning', 'processing', 'ready', 'rejected', 'failed', 'trashed');

CREATE TABLE users (
  id uuid PRIMARY KEY,
  identity_provider_uid text NOT NULL UNIQUE,
  email text NOT NULL,
  email_verified boolean NOT NULL DEFAULT false,
  display_name text NOT NULL CHECK (char_length(display_name) BETWEEN 1 AND 120),
  avatar_url text,
  locale text NOT NULL DEFAULT 'en',
  timezone text NOT NULL DEFAULT 'UTC',
  personal_workspace_id uuid,
  deletion_scheduled_at timestamptz,
  deleted_at timestamptz,
  version bigint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE auth_sessions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash bytea NOT NULL UNIQUE,
  device_label text NOT NULL,
  user_agent text,
  ip_prefix text,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz
);
CREATE INDEX auth_sessions_user_active_idx ON auth_sessions(user_id, expires_at) WHERE revoked_at IS NULL;

CREATE TABLE workspaces (
  id uuid PRIMARY KEY,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
  slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
  kind workspace_kind NOT NULL,
  owner_user_id uuid NOT NULL REFERENCES users(id),
  version bigint NOT NULL DEFAULT 1,
  deleted_at timestamptz,
  purge_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE users ADD CONSTRAINT users_personal_workspace_fk
  FOREIGN KEY (personal_workspace_id) REFERENCES workspaces(id);

CREATE TABLE memberships (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id),
  role membership_role NOT NULL,
  status membership_status NOT NULL DEFAULT 'active',
  can_spend boolean NOT NULL DEFAULT false,
  monthly_limit_micros bigint CHECK (monthly_limit_micros IS NULL OR monthly_limit_micros >= 0),
  per_run_limit_micros bigint CHECK (per_run_limit_micros IS NULL OR per_run_limit_micros >= 0),
  library_publish boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  removed_at timestamptz,
  UNIQUE (workspace_id, user_id)
);
CREATE UNIQUE INDEX memberships_single_owner_idx ON memberships(workspace_id)
  WHERE role = 'owner' AND status = 'active';
CREATE INDEX memberships_user_active_idx ON memberships(user_id, workspace_id) WHERE status = 'active';

CREATE TABLE invitations (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  email text NOT NULL,
  role membership_role NOT NULL CHECK (role <> 'owner'),
  token_hash bytea NOT NULL UNIQUE,
  invited_by uuid NOT NULL REFERENCES users(id),
  expires_at timestamptz NOT NULL,
  accepted_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX invitations_pending_email_idx ON invitations(workspace_id, lower(email))
  WHERE accepted_at IS NULL AND revoked_at IS NULL;

CREATE TABLE workspace_policies (
  workspace_id uuid PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
  currency char(3) NOT NULL DEFAULT 'USD',
  monthly_limit_micros bigint CHECK (monthly_limit_micros IS NULL OR monthly_limit_micros >= 0),
  per_run_approval_micros bigint CHECK (per_run_approval_micros IS NULL OR per_run_approval_micros >= 0),
  editor_can_publish_library boolean NOT NULL DEFAULT false,
  version bigint NOT NULL DEFAULT 1,
  updated_by uuid NOT NULL REFERENCES users(id),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE projects (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  slug text NOT NULL,
  status project_status NOT NULL DEFAULT 'active',
  privacy text NOT NULL DEFAULT 'private' CHECK (privacy = 'private'),
  current_version bigint NOT NULL DEFAULT 1,
  created_by uuid NOT NULL REFERENCES users(id),
  deleted_at timestamptz,
  purge_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (workspace_id, slug)
);

CREATE TABLE project_versions (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  version bigint NOT NULL,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 160),
  project_type text NOT NULL CHECK (project_type IN ('single', 'series')),
  production_format text NOT NULL CHECK (production_format IN ('short_film','feature','episodic','music_video','advertisement','trailer','other')),
  aspect_width integer NOT NULL CHECK (aspect_width > 0),
  aspect_height integer NOT NULL CHECK (aspect_height > 0),
  frame_rate_numerator integer NOT NULL CHECK (frame_rate_numerator > 0),
  frame_rate_denominator integer NOT NULL CHECK (frame_rate_denominator > 0),
  audio_language text NOT NULL,
  rating text NOT NULL CHECK (rating IN ('family','moderate','mature')),
  style_direction text NOT NULL DEFAULT '' CHECK (char_length(style_direction) <= 2000),
  quality_policy text NOT NULL CHECK (quality_policy IN ('draft','balanced','final')),
  cost_ceiling_micros bigint CHECK (cost_ceiling_micros IS NULL OR cost_ceiling_micros >= 0),
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (project_id, version)
);
CREATE INDEX projects_workspace_status_idx ON projects(workspace_id, status, updated_at DESC);

CREATE TABLE project_templates (
  id uuid PRIMARY KEY,
  workspace_id uuid REFERENCES workspaces(id) ON DELETE CASCADE,
  key text,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  current_version bigint NOT NULL DEFAULT 1,
  system boolean NOT NULL DEFAULT false,
  archived_at timestamptz,
  created_by uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE NULLS NOT DISTINCT (workspace_id, key)
);

CREATE TABLE project_template_versions (
  id uuid PRIMARY KEY,
  template_id uuid NOT NULL REFERENCES project_templates(id) ON DELETE CASCADE,
  version bigint NOT NULL,
  settings jsonb NOT NULL,
  library_links jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_by uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (template_id, version)
);

CREATE TABLE libraries (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
  type text NOT NULL CHECK (type IN ('story','brand')),
  version bigint NOT NULL DEFAULT 1,
  archived_at timestamptz,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (workspace_id, name)
);

CREATE TABLE library_entities (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  library_id uuid NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
  current_version_id uuid,
  archived_at timestamptz,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE library_entity_versions (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  entity_id uuid NOT NULL REFERENCES library_entities(id) ON DELETE CASCADE,
  version bigint NOT NULL,
  type text NOT NULL CHECK (type IN ('character','location','prop','faction','creature','vehicle','style_guide','voice_profile','other')),
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 160),
  aliases text[] NOT NULL DEFAULT '{}',
  summary text NOT NULL DEFAULT '',
  description text NOT NULL DEFAULT '',
  tags text[] NOT NULL DEFAULT '{}',
  attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
  status entity_status NOT NULL DEFAULT 'draft',
  change_summary text NOT NULL DEFAULT '',
  author_user_id uuid NOT NULL REFERENCES users(id),
  source_version_id uuid REFERENCES library_entity_versions(id),
  search_vector tsvector NOT NULL DEFAULT ''::tsvector,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (entity_id, version)
);
ALTER TABLE library_entities ADD CONSTRAINT library_entities_current_version_fk
  FOREIGN KEY (current_version_id) REFERENCES library_entity_versions(id);
CREATE FUNCTION library_entity_version_search_vector() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  NEW.search_vector := to_tsvector(
    'simple',
    concat_ws(' ', NEW.name, array_to_string(NEW.aliases, ' '), NEW.summary,
      NEW.description, array_to_string(NEW.tags, ' '))
  );
  RETURN NEW;
END
$$;
CREATE TRIGGER library_entity_version_search_vector_trigger
  BEFORE INSERT OR UPDATE OF name, aliases, summary, description, tags
  ON library_entity_versions FOR EACH ROW
  EXECUTE FUNCTION library_entity_version_search_vector();
CREATE INDEX library_entity_search_idx ON library_entity_versions USING GIN(search_vector);

CREATE TABLE media_assets (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('image','video','audio','voice','3d')),
  state media_state NOT NULL DEFAULT 'initiated',
  original_filename text NOT NULL,
  content_type text NOT NULL,
  size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
  quarantine_object text NOT NULL,
  created_by uuid NOT NULL REFERENCES users(id),
  deleted_at timestamptz,
  purge_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE media_asset_versions (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  asset_id uuid NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
  version bigint NOT NULL,
  storage_object text NOT NULL,
  preview_object text,
  thumbnail_object text,
  waveform_object text,
  checksum_sha256 text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  rejection_code text,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (asset_id, version)
);

CREATE TABLE library_entity_media (
  entity_version_id uuid NOT NULL REFERENCES library_entity_versions(id) ON DELETE CASCADE,
  media_asset_version_id uuid NOT NULL REFERENCES media_asset_versions(id),
  role text NOT NULL DEFAULT 'reference',
  sort_order integer NOT NULL DEFAULT 0,
  PRIMARY KEY (entity_version_id, media_asset_version_id)
);

CREATE TABLE project_library_links (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  entity_id uuid NOT NULL REFERENCES library_entities(id),
  entity_version_id uuid NOT NULL REFERENCES library_entity_versions(id),
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(project_id, entity_id)
);

CREATE TABLE audit_events (
  id uuid PRIMARY KEY,
  workspace_id uuid REFERENCES workspaces(id) ON DELETE SET NULL,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  action text NOT NULL,
  target_type text NOT NULL,
  target_id uuid,
  decision text NOT NULL DEFAULT 'success',
  source_context jsonb NOT NULL DEFAULT '{}'::jsonb,
  correlation_id text NOT NULL,
  occurred_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_events_workspace_time_idx ON audit_events(workspace_id, occurred_at DESC);

CREATE TABLE idempotency_keys (
  workspace_id uuid,
  actor_user_id uuid NOT NULL REFERENCES users(id),
  key text NOT NULL,
  route text NOT NULL,
  request_hash text NOT NULL,
  response_status integer,
  response_body jsonb,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE NULLS NOT DISTINCT(workspace_id, actor_user_id, key, route)
);

INSERT INTO project_templates (id, workspace_id, key, name, description, system, current_version, created_at, updated_at)
VALUES
  ('00000000-0000-7000-8000-000000000001', NULL, 'blank-film', 'Blank film', 'A clean 16:9 cinematic project.', true, 1, now(), now()),
  ('00000000-0000-7000-8000-000000000002', NULL, 'vertical-short', 'Vertical short', 'A 9:16 project for short-form platforms.', true, 1, now(), now()),
  ('00000000-0000-7000-8000-000000000003', NULL, 'episodic-series', 'Episodic series', 'A 16:9 series project.', true, 1, now(), now());

INSERT INTO project_template_versions (id, template_id, version, settings)
VALUES
  ('00000000-0000-7000-8000-000000000011','00000000-0000-7000-8000-000000000001',1,'{"projectType":"single","productionFormat":"short_film","aspectWidth":16,"aspectHeight":9,"frameRateNumerator":24,"frameRateDenominator":1,"rating":"moderate","qualityPolicy":"balanced"}'),
  ('00000000-0000-7000-8000-000000000012','00000000-0000-7000-8000-000000000002',1,'{"projectType":"single","productionFormat":"other","aspectWidth":9,"aspectHeight":16,"frameRateNumerator":30,"frameRateDenominator":1,"rating":"moderate","qualityPolicy":"balanced"}'),
  ('00000000-0000-7000-8000-000000000013','00000000-0000-7000-8000-000000000003',1,'{"projectType":"series","productionFormat":"episodic","aspectWidth":16,"aspectHeight":9,"frameRateNumerator":24,"frameRateDenominator":1,"rating":"moderate","qualityPolicy":"balanced"}');

ALTER TABLE memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE project_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE libraries ENABLE ROW LEVEL SECURITY;
ALTER TABLE library_entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE library_entity_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE media_assets ENABLE ROW LEVEL SECURITY;
ALTER TABLE media_asset_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE project_library_links ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspace_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;

CREATE FUNCTION app_user_id() RETURNS uuid LANGUAGE sql STABLE AS $$
  SELECT NULLIF(current_setting('app.user_id', true), '')::uuid
$$;

CREATE FUNCTION app_workspace_member(target uuid) RETURNS boolean LANGUAGE sql STABLE SECURITY DEFINER SET search_path = public AS $$
  SELECT EXISTS (SELECT 1 FROM memberships WHERE workspace_id = target AND user_id = app_user_id() AND status = 'active')
$$;

CREATE POLICY memberships_tenant ON memberships USING (app_workspace_member(workspace_id));
CREATE POLICY projects_tenant ON projects USING (app_workspace_member(workspace_id));
CREATE POLICY project_versions_tenant ON project_versions USING (app_workspace_member(workspace_id));
CREATE POLICY libraries_tenant ON libraries USING (app_workspace_member(workspace_id));
CREATE POLICY library_entities_tenant ON library_entities USING (app_workspace_member(workspace_id));
CREATE POLICY library_entity_versions_tenant ON library_entity_versions USING (app_workspace_member(workspace_id));
CREATE POLICY media_assets_tenant ON media_assets USING (app_workspace_member(workspace_id));
CREATE POLICY media_asset_versions_tenant ON media_asset_versions USING (app_workspace_member(workspace_id));
CREATE POLICY project_library_links_tenant ON project_library_links USING (app_workspace_member(workspace_id));
CREATE POLICY workspace_policies_tenant ON workspace_policies USING (app_workspace_member(workspace_id));
CREATE POLICY audit_events_tenant ON audit_events USING (workspace_id IS NULL OR app_workspace_member(workspace_id));
