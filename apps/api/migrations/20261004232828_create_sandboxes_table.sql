-- +migrate Up
CREATE TABLE IF NOT EXISTS sandboxes (
  id UUID PRIMARY KEY,
  owner_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
  name VARCHAR(255) NOT NULL,
  cpu INT NOT NULL,
  memory_gb INT NOT NULL,
  disk_gb INT NOT NULL,
  state VARCHAR(50) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sandboxes_owner_id_id ON sandboxes (owner_id, id DESC);

-- +migrate Down
DROP TABLE IF EXISTS sandboxes;
