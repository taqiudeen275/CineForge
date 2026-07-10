ALTER TABLE projects ADD COLUMN source_template_id uuid REFERENCES project_templates(id);
ALTER TABLE projects ADD COLUMN source_template_version bigint;
CREATE INDEX projects_source_template_idx ON projects(source_template_id) WHERE source_template_id IS NOT NULL;
