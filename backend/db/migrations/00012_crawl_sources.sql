-- +goose Up
-- Additive only: one new column and three new tables. Nothing existing is altered or dropped.
ALTER TABLE tenant_profiles
  ADD COLUMN IF NOT EXISTS google_maps_url text
  CHECK (google_maps_url IS NULL OR (length(google_maps_url) <= 2048 AND google_maps_url ~ '^https://'));

CREATE TABLE IF NOT EXISTS crawl_sources (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  url text NOT NULL CHECK (length(url) <= 2048),
  kind text NOT NULL CHECK (kind IN ('SITE','SINGLE_PAGE','GOOGLE_DOC','GOOGLE_SHEET','GOOGLE_SLIDES','GOOGLE_DRIVE_FILE')),
  max_depth smallint NOT NULL DEFAULT 2 CHECK (max_depth BETWEEN 0 AND 3),
  max_pages integer NOT NULL DEFAULT 50 CHECK (max_pages BETWEEN 1 AND 200),
  schedule_mode text NOT NULL DEFAULT 'MANUAL' CHECK (schedule_mode IN ('MANUAL','INTERVAL')),
  interval_minutes integer CHECK (interval_minutes IS NULL OR interval_minutes >= 360),
  enabled boolean NOT NULL DEFAULT true,
  forced_disabled boolean NOT NULL DEFAULT false, -- set by a superadmin only
  next_run_at timestamptz,
  last_run_at timestamptz,
  last_status text,
  error_public_message text,
  created_by uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, url),
  CHECK (schedule_mode <> 'INTERVAL' OR interval_minutes IS NOT NULL)
);
CREATE INDEX IF NOT EXISTS idx_crawl_sources_due ON crawl_sources(next_run_at)
  WHERE enabled AND NOT forced_disabled AND schedule_mode = 'INTERVAL';

CREATE TABLE IF NOT EXISTS crawl_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL,
  source_id uuid NOT NULL,
  trigger text NOT NULL CHECK (trigger IN ('SCHEDULED','MANUAL')),
  triggered_by uuid REFERENCES users(id),
  status text NOT NULL DEFAULT 'QUEUED' CHECK (status IN ('QUEUED','RUNNING','SUCCEEDED','PARTIAL','FAILED','NEEDS_PUBLIC_SHARING')),
  pages_new integer NOT NULL DEFAULT 0,
  pages_changed integer NOT NULL DEFAULT 0,
  pages_unchanged integer NOT NULL DEFAULT 0,
  pages_failed integer NOT NULL DEFAULT 0,
  error_code text,
  error_public_message text,
  started_at timestamptz,
  finished_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, source_id) REFERENCES crawl_sources(tenant_id, id) ON DELETE CASCADE
);
-- At most one active run per source: a manual click during a scheduled run is a conflict, not a second crawl.
CREATE UNIQUE INDEX IF NOT EXISTS ux_active_crawl_run ON crawl_runs(source_id) WHERE status IN ('QUEUED','RUNNING');
CREATE INDEX IF NOT EXISTS idx_crawl_runs_source_time ON crawl_runs(source_id, created_at DESC);

CREATE TABLE IF NOT EXISTS crawled_pages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL,
  source_id uuid NOT NULL,
  url text NOT NULL,
  title text,
  content_sha256 char(64),
  file_id uuid,
  status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','GONE','EXCLUDED')),
  miss_count smallint NOT NULL DEFAULT 0,
  first_seen_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  last_changed_at timestamptz,
  UNIQUE (source_id, url),
  FOREIGN KEY (tenant_id, source_id) REFERENCES crawl_sources(tenant_id, id) ON DELETE CASCADE,
  FOREIGN KEY (tenant_id, file_id) REFERENCES files(tenant_id, id)
);
CREATE INDEX IF NOT EXISTS idx_crawled_pages_source ON crawled_pages(source_id, status);

-- +goose Down
DROP TABLE IF EXISTS crawled_pages;
DROP TABLE IF EXISTS crawl_runs;
DROP TABLE IF EXISTS crawl_sources;
ALTER TABLE tenant_profiles DROP COLUMN IF EXISTS google_maps_url;
