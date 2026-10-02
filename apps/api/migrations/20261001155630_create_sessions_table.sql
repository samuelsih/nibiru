-- +migrate Up
CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY,
  token VARCHAR(255) NOT NULL UNIQUE,
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMPTZ NOT NULL,
  metadata JSONB
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);

-- +migrate Down
DROP TABLE IF EXISTS sessions;