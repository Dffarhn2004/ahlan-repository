-- +goose Up
-- PSB_ONLY: the number is a dedicated admissions (PSB) line, the assistant answers everything as before.
-- GENERAL: a shared number; the assistant first asks what the visitor needs, answers PSB itself and
-- hands other needs to the matching admin, who is also notified on WhatsApp.
ALTER TABLE whatsapp_channels ADD COLUMN IF NOT EXISTS service_mode text NOT NULL DEFAULT 'PSB_ONLY' CHECK (service_mode IN ('PSB_ONLY','GENERAL'));

-- Admins who receive a WhatsApp notification when the assistant routes a visitor to them. This is not the public
-- contact list (tenant_contacts): these numbers are never shown to visitors.
CREATE TABLE IF NOT EXISTS whatsapp_admins (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name text NOT NULL CHECK (length(name) BETWEEN 1 AND 80),
  phone text NOT NULL CHECK (phone ~ '^\+[0-9]{8,15}$'),
  notify_when text NOT NULL CHECK (length(notify_when) BETWEEN 1 AND 200),
  sort_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_whatsapp_admins_tenant ON whatsapp_admins(tenant_id, sort_order);
-- +goose Down
DROP TABLE IF EXISTS whatsapp_admins;
ALTER TABLE whatsapp_channels DROP COLUMN IF EXISTS service_mode;
