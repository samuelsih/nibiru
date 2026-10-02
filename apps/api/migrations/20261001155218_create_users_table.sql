-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email VARCHAR(320) NOT NULL,
  password TEXT,
  email_verified_at TIMESTAMPTZ,
  first_name VARCHAR(255),
  last_name VARCHAR(255),
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX idx_users_email ON users(email);

-- +migrate Down
DROP TABLE IF EXISTS users;