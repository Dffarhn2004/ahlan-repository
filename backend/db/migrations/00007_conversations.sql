-- +goose Up
CREATE TABLE IF NOT EXISTS telegram_conversations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  telegram_chat_id bigint NOT NULL,
  telegram_user_id bigint,
  username text,
  status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','RESOLVED','HANDOFF')),
  started_at timestamptz NOT NULL DEFAULT now(),
  last_message_at timestamptz NOT NULL DEFAULT now(),
  closed_at timestamptz,
  UNIQUE (tenant_id, telegram_chat_id)
);
CREATE TABLE IF NOT EXISTS telegram_messages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id uuid NOT NULL REFERENCES telegram_conversations(id) ON DELETE CASCADE,
  sender_type text NOT NULL CHECK (sender_type IN ('USER','BOT','ADMIN')),
  message_text text NOT NULL,
  telegram_message_id bigint,
  source_documents jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_telegram_conversations_tenant_time ON telegram_conversations(tenant_id, last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_telegram_messages_conversation_time ON telegram_messages(conversation_id, created_at ASC);
-- +goose Down
DROP TABLE IF EXISTS telegram_messages;
DROP TABLE IF EXISTS telegram_conversations;
