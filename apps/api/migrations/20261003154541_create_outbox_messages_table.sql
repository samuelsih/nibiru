-- +migrate Up
CREATE TABLE IF NOT EXISTS outbox_messages (
  id UUID PRIMARY KEY,
  type VARCHAR(100) NOT NULL,
  metadata JSONB,
  payload JSONB NOT NULL,
  error TEXT,
  delivered_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_messages_undelivered
ON outbox_messages (created_at, delivered_at)
INCLUDE (id, type, payload)
WHERE delivered_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS outbox_messages;
