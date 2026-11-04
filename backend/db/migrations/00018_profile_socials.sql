-- +goose Up
-- Social media pages shown on the public link page. Empty means the pesantren has none.
ALTER TABLE tenant_profiles
  ADD COLUMN IF NOT EXISTS instagram_url text CHECK (instagram_url IS NULL OR (length(instagram_url) <= 2048 AND instagram_url ~ '^https://')),
  ADD COLUMN IF NOT EXISTS facebook_url text CHECK (facebook_url IS NULL OR (length(facebook_url) <= 2048 AND facebook_url ~ '^https://')),
  ADD COLUMN IF NOT EXISTS tiktok_url text CHECK (tiktok_url IS NULL OR (length(tiktok_url) <= 2048 AND tiktok_url ~ '^https://')),
  ADD COLUMN IF NOT EXISTS youtube_url text CHECK (youtube_url IS NULL OR (length(youtube_url) <= 2048 AND youtube_url ~ '^https://'));

-- +goose Down
ALTER TABLE tenant_profiles
  DROP COLUMN IF EXISTS instagram_url,
  DROP COLUMN IF EXISTS facebook_url,
  DROP COLUMN IF EXISTS tiktok_url,
  DROP COLUMN IF EXISTS youtube_url;
