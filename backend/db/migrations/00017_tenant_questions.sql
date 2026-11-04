-- +goose Up
-- Questions the assistant could not answer, collected so an admin can answer them or mark them resolved.
-- One row per distinct question (question_key is the normalised wording); ask_count says how often it came up.
--   OPEN      waiting for an admin
--   ANSWERED  the admin wrote an answer, which is given to the assistant (faq_question + answer)
--   RESOLVED  the admin fixed it another way (for example by uploading a document)
--   IGNORED   not worth answering (off-topic, spam)
CREATE TABLE IF NOT EXISTS tenant_questions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  question text NOT NULL CHECK (length(question) BETWEEN 1 AND 500),
  question_key text NOT NULL CHECK (length(question_key) BETWEEN 1 AND 200),
  ask_count integer NOT NULL DEFAULT 1,
  status text NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN','ANSWERED','RESOLVED','IGNORED')),
  faq_question text NOT NULL DEFAULT '' CHECK (length(faq_question) <= 200),
  answer text NOT NULL DEFAULT '' CHECK (length(answer) <= 1500),
  first_asked_at timestamptz NOT NULL DEFAULT now(),
  last_asked_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, question_key)
);
CREATE INDEX IF NOT EXISTS idx_tenant_questions_tenant ON tenant_questions(tenant_id, status, last_asked_at DESC);

-- +goose Down
DROP TABLE IF EXISTS tenant_questions;
