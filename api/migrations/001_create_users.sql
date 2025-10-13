CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name STRING NOT NULL,
    email STRING NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);
