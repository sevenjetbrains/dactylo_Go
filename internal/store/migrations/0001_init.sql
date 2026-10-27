-- Schéma initial (cahier des charges, section 3.3).
-- Ne jamais modifier ce fichier après publication : créer 0002_xxx.sql.

-- +goose Up
CREATE TABLE profiles (
    id              INTEGER PRIMARY KEY,
    name            TEXT    NOT NULL,
    keyboard_layout TEXT    NOT NULL DEFAULT 'azerty',
    settings        TEXT    NOT NULL DEFAULT '{}', -- JSON
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE texts (
    id          INTEGER PRIMARY KEY,
    title       TEXT    NOT NULL,
    content     TEXT    NOT NULL,
    language    TEXT    NOT NULL DEFAULT 'fr',
    source      TEXT    NOT NULL DEFAULT 'paste' CHECK (source IN ('paste', 'pdf', 'builtin')),
    source_file TEXT,
    pages       TEXT,                          -- ex. "3-5"
    tags        TEXT    NOT NULL DEFAULT '[]', -- JSON
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE text_progress (
    profile_id  INTEGER NOT NULL REFERENCES profiles (id) ON DELETE CASCADE,
    text_id     INTEGER NOT NULL REFERENCES texts (id) ON DELETE CASCADE,
    position    INTEGER NOT NULL DEFAULT 0,
    last_opened TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    PRIMARY KEY (profile_id, text_id)
);

CREATE TABLE sessions (
    id         INTEGER PRIMARY KEY,
    profile_id INTEGER NOT NULL REFERENCES profiles (id) ON DELETE CASCADE,
    text_id    INTEGER REFERENCES texts (id) ON DELETE SET NULL,
    mode       TEXT    NOT NULL CHECK (mode IN ('copy', 'dictation', 'review')),
    started_at TEXT    NOT NULL,
    ended_at   TEXT,
    wpm_gross  REAL,
    wpm_net    REAL,
    cpm        REAL,
    accuracy   REAL,
    settings   TEXT    NOT NULL DEFAULT '{}' -- JSON
);
CREATE INDEX idx_sessions_profile_started ON sessions (profile_id, started_at);

CREATE TABLE keystrokes (
    session_id INTEGER NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    idx        INTEGER NOT NULL,
    expected   TEXT    NOT NULL,
    typed      TEXT    NOT NULL,
    ts_ms      INTEGER NOT NULL,
    correct    INTEGER NOT NULL CHECK (correct IN (0, 1)),
    PRIMARY KEY (session_id, idx)
);

CREATE TABLE session_errors (
    id            INTEGER PRIMARY KEY,
    session_id    INTEGER NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    expected_word TEXT    NOT NULL,
    typed_word    TEXT    NOT NULL,
    error_type    TEXT    NOT NULL,
    position      INTEGER NOT NULL
);
CREATE INDEX idx_session_errors_session ON session_errors (session_id);

CREATE TABLE words (
    id           INTEGER PRIMARY KEY,
    profile_id   INTEGER NOT NULL REFERENCES profiles (id) ON DELETE CASCADE,
    word         TEXT    NOT NULL, -- normalisé NFC, minuscules
    encounters   INTEGER NOT NULL DEFAULT 0,
    errors       INTEGER NOT NULL DEFAULT 0,
    avg_time_ms  REAL,
    mastery      REAL    NOT NULL DEFAULT 0,
    next_review  TEXT,
    srs_params   TEXT    NOT NULL DEFAULT '{}', -- JSON
    UNIQUE (profile_id, word)
);
CREATE INDEX idx_words_next_review ON words (profile_id, next_review);

CREATE TABLE word_attempts (
    id          INTEGER PRIMARY KEY,
    word_id     INTEGER NOT NULL REFERENCES words (id) ON DELETE CASCADE,
    session_id  INTEGER REFERENCES sessions (id) ON DELETE SET NULL,
    typed_form  TEXT    NOT NULL,
    correct     INTEGER NOT NULL CHECK (correct IN (0, 1)),
    duration_ms INTEGER,
    attempted_at TEXT   NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX idx_word_attempts_word ON word_attempts (word_id);

CREATE TABLE dictation_profiles (
    id                INTEGER PRIMARY KEY,
    name              TEXT    NOT NULL,
    voice             TEXT,
    rate              REAL    NOT NULL DEFAULT 1.0,
    pitch             REAL    NOT NULL DEFAULT 1.0,
    volume            REAL    NOT NULL DEFAULT 1.0,
    chunking          TEXT    NOT NULL DEFAULT 'group' CHECK (chunking IN ('word', 'group', 'sentence', 'paragraph')),
    group_size        INTEGER NOT NULL DEFAULT 4,
    pause_ms          INTEGER NOT NULL DEFAULT 1000,
    speak_punctuation INTEGER NOT NULL DEFAULT 0 CHECK (speak_punctuation IN (0, 1))
);

-- +goose Down
DROP TABLE dictation_profiles;
DROP TABLE word_attempts;
DROP TABLE words;
DROP TABLE session_errors;
DROP TABLE keystrokes;
DROP TABLE sessions;
DROP TABLE text_progress;
DROP TABLE texts;
DROP TABLE profiles;
