DROP INDEX IF EXISTS projects_source_template_idx;
ALTER TABLE projects DROP COLUMN IF EXISTS source_template_version;
ALTER TABLE projects DROP COLUMN IF EXISTS source_template_id;
