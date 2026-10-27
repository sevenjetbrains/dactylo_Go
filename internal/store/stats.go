package store

import (
	"context"
	"database/sql"
	"time"

	"dactylo/internal/core/typing"
)

// SessionPoint est une session terminée, pour les courbes et l'export.
type SessionPoint struct {
	ID         int64   `json:"id"`
	EndedAt    string  `json:"endedAt"`
	Mode       string  `json:"mode"`
	TextTitle  string  `json:"textTitle"`
	DurationMs int64   `json:"durationMs"`
	WPMGross   float64 `json:"wpmGross"`
	WPMNet     float64 `json:"wpmNet"`
	CPM        float64 `json:"cpm"`
	Accuracy   float64 `json:"accuracy"`
	Errors     int     `json:"errors"` // mots fautifs
}

// Totals résume l'activité d'un profil.
type Totals struct {
	Sessions int     `json:"sessions"`
	TotalMs  int64   `json:"totalMs"`
	BestWPM  float64 `json:"bestWpm"`
}

// SessionPoints renvoie les `limit` dernières sessions terminées, de la plus
// ancienne à la plus récente (limit <= 0 : toutes).
func (s *Store) SessionPoints(ctx context.Context, profileID int64, limit int) ([]SessionPoint, error) {
	if limit <= 0 {
		limit = -1
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT * FROM (
			SELECT s.id, s.ended_at, s.mode, COALESCE(t.title, ''), COALESCE(s.duration_ms, 0),
			       COALESCE(s.wpm_gross, 0), COALESCE(s.wpm_net, 0), COALESCE(s.cpm, 0), COALESCE(s.accuracy, 0),
			       (SELECT COUNT(*) FROM session_errors e WHERE e.session_id = s.id)
			FROM sessions s LEFT JOIN texts t ON t.id = s.text_id
			WHERE s.profile_id = ? AND s.ended_at IS NOT NULL
			ORDER BY s.ended_at DESC, s.id DESC LIMIT ?)
		ORDER BY 2, 1`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []SessionPoint{}
	for rows.Next() {
		var p SessionPoint
		if err := rows.Scan(&p.ID, &p.EndedAt, &p.Mode, &p.TextTitle, &p.DurationMs,
			&p.WPMGross, &p.WPMNet, &p.CPM, &p.Accuracy, &p.Errors); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SessionTotals renvoie nombre de sessions, temps total et record (MPM net).
func (s *Store) SessionTotals(ctx context.Context, profileID int64) (Totals, error) {
	var t Totals
	var best sql.NullFloat64
	err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(duration_ms), 0), MAX(wpm_net)
		FROM sessions WHERE profile_id = ? AND ended_at IS NOT NULL`, profileID).Scan(&t.Sessions, &t.TotalMs, &best)
	t.BestWPM = best.Float64
	return t, err
}

// SessionDays renvoie l'heure de fin de chaque session terminée.
func (s *Store) SessionDays(ctx context.Context, profileID int64) ([]time.Time, error) {
	rows, err := s.DB.QueryContext(ctx,
		"SELECT ended_at FROM sessions WHERE profile_id = ? AND ended_at IS NOT NULL", profileID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []time.Time
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
			out = append(out, t)
		}
	}
	return out, rows.Err()
}

// RecentKeystrokes renvoie les frappes corrigées des `sessions` dernières
// sessions terminées, une liste par session (F-34).
func (s *Store) RecentKeystrokes(ctx context.Context, profileID int64, sessions int) ([][]typing.Scored, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT k.session_id, k.idx, k.position, k.expected, k.typed, k.ts_ms, k.correct
		FROM keystrokes k
		JOIN (SELECT id FROM sessions WHERE profile_id = ? AND ended_at IS NOT NULL
		      ORDER BY ended_at DESC LIMIT ?) s ON s.id = k.session_id
		ORDER BY k.session_id, k.idx`, profileID, sessions)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out [][]typing.Scored
	var current int64 = -1
	for rows.Next() {
		var sid int64
		var k typing.Scored
		var correct int
		if err := rows.Scan(&sid, &k.Seq, &k.Position, &k.Expected, &k.Typed, &k.TimeMs, &correct); err != nil {
			return nil, err
		}
		k.Correct = correct == 1
		if sid != current {
			out = append(out, nil)
			current = sid
		}
		out[len(out)-1] = append(out[len(out)-1], k)
	}
	return out, rows.Err()
}
