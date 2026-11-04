-- +goose Up
CREATE TABLE IF NOT EXISTS whatsapp_channels (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
  phone_number text NOT NULL,
  status text NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','ACTIVE','SUSPENDED')),
  access_mode text NOT NULL DEFAULT 'TESTING' CHECK (access_mode IN ('TESTING','PUBLIC')),
  allowed_numbers text[] NOT NULL DEFAULT '{}',
  linked_at timestamptz NOT NULL DEFAULT now(),
  verified_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT now()
);
-- Conversations are shared by both channels; WhatsApp chat IDs are the digits of the sender's number.
ALTER TABLE telegram_conversations ADD COLUMN IF NOT EXISTS channel text NOT NULL DEFAULT 'TELEGRAM' CHECK (channel IN ('TELEGRAM','WHATSAPP'));
ALTER TABLE telegram_conversations DROP CONSTRAINT IF EXISTS telegram_conversations_tenant_id_telegram_chat_id_key;
ALTER TABLE telegram_conversations ADD CONSTRAINT telegram_conversations_tenant_channel_chat_key UNIQUE (tenant_id, channel, telegram_chat_id);
-- +goose Down
ALTER TABLE telegram_conversations DROP CONSTRAINT IF EXISTS telegram_conversations_tenant_channel_chat_key;
DELETE FROM telegram_conversations WHERE channel <> 'TELEGRAM';
ALTER TABLE telegram_conversations ADD CONSTRAINT telegram_conversations_tenant_id_telegram_chat_id_key UNIQUE (tenant_id, telegram_chat_id);
ALTER TABLE telegram_conversations DROP COLUMN IF EXISTS channel;
DROP TABLE IF EXISTS whatsapp_channels;
