-- Phase 3 : durée effective des sessions (temps total, F-33), forme
-- affichée des mots difficiles et détail des tentatives (F-41).

-- +goose Up
ALTER TABLE sessions ADD COLUMN duration_ms INTEGER;
ALTER TABLE words ADD COLUMN shown TEXT;
ALTER TABLE words ADD COLUMN last_error_at TEXT;
ALTER TABLE word_attempts ADD COLUMN error_type TEXT;
ALTER TABLE word_attempts ADD COLUMN slow INTEGER NOT NULL DEFAULT 0;
CREATE INDEX idx_sessions_profile_ended ON sessions (profile_id, ended_at);

-- +goose Down
DROP INDEX idx_sessions_profile_ended;
ALTER TABLE word_attempts DROP COLUMN slow;
ALTER TABLE word_attempts DROP COLUMN error_type;
ALTER TABLE words DROP COLUMN last_error_at;
ALTER TABLE words DROP COLUMN shown;
ALTER TABLE sessions DROP COLUMN duration_ms;
