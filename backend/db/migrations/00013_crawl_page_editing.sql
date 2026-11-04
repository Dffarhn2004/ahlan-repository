-- +goose Up
-- Additive only. Lets a pesantren see what each crawl changed and edit what the assistant reads.
--   body_text    what the assistant currently reads for this page (the stored file is built from it)
--   source_text  the latest text fetched from the web, kept so an edit can be reverted or compared
--   content_sha256 / source_hash  hash of the source text body_text is based on / of the newest fetch;
--                 they differ only when the site changed after the pesantren edited the page
ALTER TABLE crawled_pages
  ADD COLUMN IF NOT EXISTS ext text NOT NULL DEFAULT '.txt',
  ADD COLUMN IF NOT EXISTS body_text text,
  ADD COLUMN IF NOT EXISTS source_text text,
  ADD COLUMN IF NOT EXISTS source_hash char(64),
  ADD COLUMN IF NOT EXISTS user_edited boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS edited_at timestamptz,
  ADD COLUMN IF NOT EXISTS edited_by uuid REFERENCES users(id);

CREATE TABLE IF NOT EXISTS crawl_page_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL,
  source_id uuid NOT NULL,
  run_id uuid REFERENCES crawl_runs(id) ON DELETE CASCADE,
  page_id uuid REFERENCES crawled_pages(id) ON DELETE SET NULL,
  url text NOT NULL,
  title text,
  event text NOT NULL CHECK (event IN ('NEW','CHANGED','SOURCE_CHANGED','GONE')),
  lines_added integer NOT NULL DEFAULT 0,
  lines_removed integer NOT NULL DEFAULT 0,
  before_text text,
  after_text text,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, source_id) REFERENCES crawl_sources(tenant_id, id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_crawl_events_source_time ON crawl_page_events(source_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS crawl_page_events;
ALTER TABLE crawled_pages
  DROP COLUMN IF EXISTS edited_by, DROP COLUMN IF EXISTS edited_at, DROP COLUMN IF EXISTS user_edited,
  DROP COLUMN IF EXISTS source_hash, DROP COLUMN IF EXISTS source_text, DROP COLUMN IF EXISTS body_text, DROP COLUMN IF EXISTS ext;
