-- +goose Up
CREATE TABLE IF NOT EXISTS hermes_profile_settings (
  tenant_id uuid PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
  description text NOT NULL DEFAULT '',
  soul_md text NOT NULL DEFAULT '',
  enabled_skills jsonb NOT NULL DEFAULT '[]'::jsonb,
  config_revision integer NOT NULL DEFAULT 1,
  updated_by uuid REFERENCES users(id),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS hermes_profile_settings;
