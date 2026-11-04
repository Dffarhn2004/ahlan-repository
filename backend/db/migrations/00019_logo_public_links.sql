-- +goose Up
-- Logo file (storage key under the tenant's folder) shown on the public link page.
ALTER TABLE tenant_profiles ADD COLUMN IF NOT EXISTS logo_key text;
-- A link the pesantren chose to show on the public page, grouped into a folder per kind. Off by default.
ALTER TABLE tenant_links ADD COLUMN IF NOT EXISTS show_public boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE tenant_links DROP COLUMN IF EXISTS show_public;
ALTER TABLE tenant_profiles DROP COLUMN IF EXISTS logo_key;
