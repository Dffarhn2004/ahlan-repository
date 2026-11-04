-- +goose Up
-- Contact persons a pesantren registers. Every contact saved here is handed to people who ask the assistant for a
-- contact, so the dashboard warns the admin before saving. Phone is stored normalised (+<country><number>).
CREATE TABLE IF NOT EXISTS tenant_contacts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name text NOT NULL CHECK (length(name) BETWEEN 1 AND 80),
  role text NOT NULL DEFAULT '' CHECK (length(role) <= 80),
  phone text NOT NULL CHECK (phone ~ '^\+[0-9]{8,15}$'),
  email text NOT NULL DEFAULT '' CHECK (length(email) <= 120),
  hours text NOT NULL DEFAULT '' CHECK (length(hours) <= 80),
  share_hint text NOT NULL DEFAULT '' CHECK (length(share_hint) <= 160),
  sort_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_tenant_contacts_tenant ON tenant_contacts(tenant_id, sort_order, created_at);

-- +goose Down
DROP TABLE IF EXISTS tenant_contacts;
