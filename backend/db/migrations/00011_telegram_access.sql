-- +goose Up
ALTER TABLE telegram_bots
  ADD COLUMN IF NOT EXISTS access_mode text NOT NULL DEFAULT 'TESTING' CHECK (access_mode IN ('TESTING','PUBLIC')),
  ADD COLUMN IF NOT EXISTS allowed_user_ids bigint[] NOT NULL DEFAULT '{}';
-- +goose Down
ALTER TABLE telegram_bots DROP COLUMN IF EXISTS allowed_user_ids, DROP COLUMN IF EXISTS access_mode;
