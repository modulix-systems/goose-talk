BEGIN;

CREATE TABLE IF NOT EXISTS "user" (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  email CITEXT NOT NULL UNIQUE,
  password BYTEA NOT NULL,
  first_name TEXT,
  last_name TEXT,
  photo_url TEXT,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  is_active BOOL DEFAULT true NOT NULL,
  birth_date TIMESTAMPTZ,
  about_me TEXT
);

CREATE TABLE IF NOT EXISTS client_identity (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  location TEXT NOT NULL,
  ip_addr INET NOT NULL,
  device_info TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_session (
  id TEXT PRIMARY KEY,
  user_id INT REFERENCES "user"(id),
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  last_seen_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deactived_at TIMESTAMPTZ,
  client_identity_id INT REFERENCES client_identity(id)
);


CREATE TYPE two_fa_transport AS ENUM ('telegram', 'email', 'sms', 'totp_app');

CREATE TABLE IF NOT EXISTS two_factor_auth (
  user_id INT PRIMARY KEY,
  transport two_fa_transport NOT NULL,
  contact TEXT,
  totp_secret BYTEA,
  enabled BOOL DEFAULT true NOT NULL
);

COMMIT;
