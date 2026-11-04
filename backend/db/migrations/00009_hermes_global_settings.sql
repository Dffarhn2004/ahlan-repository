-- +goose Up
CREATE TABLE IF NOT EXISTS hermes_global_settings (
  id smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  description text NOT NULL DEFAULT '',
  soul_md text NOT NULL DEFAULT '',
  enabled_skills jsonb NOT NULL DEFAULT '[]'::jsonb,
  config_revision integer NOT NULL DEFAULT 1,
  updated_by uuid REFERENCES users(id),
  updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO hermes_global_settings (description, soul_md, enabled_skills, config_revision, updated_by, updated_at)
SELECT description, soul_md, enabled_skills, config_revision, updated_by, updated_at
FROM hermes_profile_settings
ORDER BY updated_at DESC
LIMIT 1
ON CONFLICT (id) DO NOTHING;

DROP TABLE IF EXISTS hermes_profile_settings;

-- +goose Down
CREATE TABLE IF NOT EXISTS hermes_profile_settings (
  tenant_id uuid PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
  description text NOT NULL DEFAULT '',
  soul_md text NOT NULL DEFAULT '',
  enabled_skills jsonb NOT NULL DEFAULT '[]'::jsonb,
  config_revision integer NOT NULL DEFAULT 1,
  updated_by uuid REFERENCES users(id),
  updated_at timestamptz NOT NULL DEFAULT now()
);
DROP TABLE IF EXISTS hermes_global_settings;
