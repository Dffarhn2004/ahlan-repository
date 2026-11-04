-- +goose Up
CREATE INDEX IF NOT EXISTS idx_audit_occurred ON audit_logs(occurred_at DESC);
-- +goose Down
DROP INDEX IF EXISTS idx_audit_occurred;
