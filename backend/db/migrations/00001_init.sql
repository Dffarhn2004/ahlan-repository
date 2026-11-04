-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL UNIQUE,
  password_hash text,
  is_platform_superadmin boolean NOT NULL DEFAULT false,
  mfa_enabled boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tenants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  lifecycle_status text NOT NULL DEFAULT 'DRAFT' CHECK (lifecycle_status IN ('DRAFT','ONBOARDING','ACTIVE','SUSPENDED','ARCHIVED','DELETING','DELETED')),
  storage_quota_bytes bigint NOT NULL DEFAULT 524288000 CHECK (storage_quota_bytes >= 0),
  config_revision integer NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE hermes_profiles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL UNIQUE REFERENCES tenants(id),
  profile_key text NOT NULL UNIQUE,
  hermes_host_id text NOT NULL,
  runtime_status text NOT NULL DEFAULT 'UNKNOWN' CHECK (runtime_status IN ('UNKNOWN','PARKED','HEALTHY','DEGRADED','OFFLINE')),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE provisioning_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  kind text NOT NULL,
  status text NOT NULL CHECK (status IN ('QUEUED','RUNNING','WAITING_INPUT','VERIFYING','SUCCEEDED','FAILED','CANCELLED')),
  attempt integer NOT NULL DEFAULT 0,
  locked_until timestamptz,
  error_code text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_active_provision_run ON provisioning_runs(tenant_id) WHERE status IN ('QUEUED','RUNNING','WAITING_INPUT','VERIFYING');

CREATE TABLE audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid REFERENCES tenants(id),
  actor_user_id uuid REFERENCES users(id),
  actor_type text NOT NULL,
  action text NOT NULL,
  target_type text,
  target_id text,
  request_id text,
  details_redacted jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE audit_logs, provisioning_runs, hermes_profiles, tenant_users, tenants, users;
