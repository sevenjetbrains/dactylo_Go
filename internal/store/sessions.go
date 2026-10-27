package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dactylo/internal/core/stats"
	"dactylo/internal/core/typing"
)

// Session est une séance enregistrée.
type Session struct {
	ID        int64
	ProfileID int64
	TextID    sql.NullInt64
	Mode      string
	StartedAt string
	EndedAt   sql.NullString
	Settings  string // JSON
}

// Averages est la moyenne des dernières sessions terminées (F-32).
type Averages struct {
	Sessions int     `json:"sessions"`
	WPMNet   float64 `json:"wpmNet"`
	Accuracy float64 `json:"accuracy"`
}

// CreateSession ouvre une session et renvoie son id.
func (s *Store) CreateSession(ctx context.Context, profileID, textID int64, mode, settings string, startedAt time.Time) (int64, error) {
	if settings == "" {
		settings = "{}"
	}
	res, err := s.DB.ExecContext(ctx,
		"INSERT INTO sessions (profile_id, text_id, mode, started_at, settings) VALUES (?, ?, ?, ?, ?)",
		profileID, textID, mode, formatTime(startedAt), settings)
	if err != nil {
		return 0, fmt.Errorf("création de la session : %w", err)
	}
	return res.LastInsertId()
}

// GetSession renvoie une session.
func (s *Store) GetSession(ctx context.Context, id int64) (Session, error) {
	var se Session
	err := s.DB.QueryRowContext(ctx,
		"SELECT id, profile_id, text_id, mode, started_at, ended_at, settings FROM sessions WHERE id = ?", id).
		Scan(&se.ID, &se.ProfileID, &se.TextID, &se.Mode, &se.StartedAt, &se.EndedAt, &se.Settings)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	return se, err
}

// SaveKeystrokes enregistre un lot de frappes. Un lot renvoyé deux fois
// (nouvel essai après une erreur réseau interne) remplace les mêmes lignes.
func (s *Store) SaveKeystrokes(ctx context.Context, sessionID int64, ks []typing.Scored) error {
	if len(ks) == 0 {
		return nil
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO keystrokes (session_id, idx, position, expected, typed, ts_ms, correct)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()
	for _, k := range ks {
		if _, err := stmt.ExecContext(ctx, sessionID, k.Seq, k.Position, k.Expected, k.Typed, k.TimeMs, boolInt(k.Correct)); err != nil {
			return fmt.Errorf("enregistrement des frappes : %w", err)
		}
	}
	return tx.Commit()
}

// Keystrokes renvoie les frappes brutes d'une session, dans l'ordre.
func (s *Store) Keystrokes(ctx context.Context, sessionID int64) ([]typing.Keystroke, error) {
	rows, err := s.DB.QueryContext(ctx,
		"SELECT idx, position, typed, ts_ms FROM keystrokes WHERE session_id = ? ORDER BY idx", sessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []typing.Keystroke
	for rows.Next() {
		var k typing.Keystroke
		if err := rows.Scan(&k.Seq, &k.Position, &k.Typed, &k.TimeMs); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// FinishSession enregistre les mesures finales et les mots fautifs.
func (s *Store) FinishSession(ctx context.Context, id int64, endedAt time.Time, sum stats.Summary, errs []typing.WordError) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE sessions SET ended_at = ?, wpm_gross = ?, wpm_net = ?, cpm = ?, accuracy = ?
		WHERE id = ? AND ended_at IS NULL`,
		formatTime(endedAt), sum.WPMGross, sum.WPMNet, sum.CPM, sum.Accuracy, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("session %d absente ou déjà terminée", id)
	}
	for _, e := range errs {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO session_errors (session_id, expected_word, typed_word, error_type, position) VALUES (?, ?, ?, ?, ?)",
			id, e.Expected, e.Typed, e.Type, e.Position); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// RecentAverages calcule la moyenne des `limit` dernières sessions terminées
// du profil dans ce mode, hors session `excludeID`.
func (s *Store) RecentAverages(ctx context.Context, profileID int64, mode string, excludeID int64, limit int) (Averages, error) {
	var a Averages
	var wpm, acc sql.NullFloat64
	err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*), AVG(wpm_net), AVG(accuracy) FROM (
			SELECT wpm_net, accuracy FROM sessions
			WHERE profile_id = ? AND mode = ? AND id <> ? AND ended_at IS NOT NULL
			ORDER BY ended_at DESC LIMIT ?)`,
		profileID, mode, excludeID, limit).Scan(&a.Sessions, &wpm, &acc)
	a.WPMNet, a.Accuracy = wpm.Float64, acc.Float64
	return a, err
}

func formatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
