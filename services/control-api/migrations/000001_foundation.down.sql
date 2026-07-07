DROP FUNCTION IF EXISTS app_workspace_member(uuid);
DROP FUNCTION IF EXISTS app_user_id();
DROP TABLE IF EXISTS idempotency_keys, audit_events, project_library_links, library_entity_media,
  media_asset_versions, media_assets, library_entity_versions, library_entities, libraries,
  project_template_versions, project_templates, project_versions, projects, workspace_policies,
  invitations, memberships, workspaces, auth_sessions, users CASCADE;
DROP FUNCTION IF EXISTS library_entity_version_search_vector();
DROP TYPE IF EXISTS media_state, entity_status, project_status, membership_status, membership_role, workspace_kind;
