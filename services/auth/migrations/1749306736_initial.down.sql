BEGIN;

DROP TABLE IF EXISTS two_factor_auth;
DROP TABLE IF EXISTS user_session;
DROP TABLE IF EXISTS client_identity;
DROP TABLE IF EXISTS "user";
DROP TABLE IF EXISTS passkey_credential;
DROP TYPE IF EXISTS two_fa_transport;

COMMIT;
