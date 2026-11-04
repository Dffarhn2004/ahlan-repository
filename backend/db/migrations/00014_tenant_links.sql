-- +goose Up
-- Official links a pesantren registers. Each can be shared by the assistant ("share"), learned from ("learn"), or both.
-- "Learn" is not a column: a link is learned exactly when it points at a crawl source, so it can never drift out of sync
-- with the Sumber Web menu.
CREATE TABLE IF NOT EXISTS tenant_links (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('PENDAFTARAN','BROSUR','JADWAL','BIAYA','WHATSAPP','INSTAGRAM','YOUTUBE','FACEBOOK','TIKTOK','WEBSITE','LAINNYA')),
  label text NOT NULL CHECK (length(label) BETWEEN 1 AND 80),
  url text NOT NULL CHECK (length(url) <= 2048 AND url ~ '^https://'),
  share_hint text NOT NULL DEFAULT '' CHECK (length(share_hint) <= 160),
  share boolean NOT NULL DEFAULT true,
  crawl_source_id uuid REFERENCES crawl_sources(id) ON DELETE SET NULL,
  sort_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, url)
);
CREATE INDEX IF NOT EXISTS idx_tenant_links_tenant ON tenant_links(tenant_id, sort_order, created_at);

-- +goose Down
DROP TABLE IF EXISTS tenant_links;
