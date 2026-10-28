-- Phase 5 : réglages de l'application (profil actif, F-50) et sessions
-- abandonnées volontairement, pour ne proposer la reprise (NF-06) qu'après
-- une fermeture brutale.

-- +goose Up
CREATE TABLE app_settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
ALTER TABLE sessions ADD COLUMN abandoned_at TEXT;

-- +goose Down
ALTER TABLE sessions DROP COLUMN abandoned_at;
DROP TABLE app_settings;
