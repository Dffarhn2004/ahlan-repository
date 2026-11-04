-- +goose Up
CREATE TABLE IF NOT EXISTS tenant_profiles (
  tenant_id uuid PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
  display_name text NOT NULL,
  description text,
  address text,
  phone text,
  email text,
  website text,
  updated_by uuid REFERENCES users(id),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS telegram_bots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
  telegram_bot_id bigint NOT NULL UNIQUE,
  username text NOT NULL,
  credential_ref text NOT NULL,
  status text NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','VERIFIED','ACTIVE','SUSPENDED','INVALID')),
  verified_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_tenant_time ON audit_logs(tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_action_time ON audit_logs(action, occurred_at DESC);
-- +goose Down
DROP TABLE IF EXISTS telegram_bots;
DROP TABLE IF EXISTS tenant_profiles;
