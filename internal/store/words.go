// Accès aux mots difficiles, à leurs tentatives et à leur révision.

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"dactylo/internal/core/srs"
	"dactylo/internal/core/typing"
)

// Word est un mot difficile suivi pour un profil (F-41).
type Word struct {
	ID          int64    `json:"id"`
	Word        string   `json:"word"`  // clé normalisée
	Shown       string   `json:"shown"` // forme à taper
	Encounters  int      `json:"encounters"`
	Errors      int      `json:"errors"`
	AvgMsPerChr *float64 `json:"avgMsPerChar"`
	Mastery     float64  `json:"mastery"`
	NextReview  string   `json:"nextReview"`
	LastErrorAt string   `json:"lastErrorAt"`
	Card        srs.Card `json:"card"`

	// Détail, rempli par ListWords
	Forms         []string `json:"forms"`         // formes erronées tapées
	DominantError string   `json:"dominantError"` // type de faute le plus fréquent
}

// WordFilter choisit les mots d'une liste.
type WordFilter struct {
	Status string `json:"status"` // due, learning, mastered, all
	Today  string `json:"-"`      // AAAA-MM-JJ local
}

// GetWord renvoie le mot suivi, ou ok = false.
func (s *Store) GetWord(ctx context.Context, profileID int64, word string) (Word, bool, error) {
	rows, err := s.queryWords(ctx, "WHERE profile_id = ? AND word = ?", profileID, word)
	if err != nil || len(rows) == 0 {
		return Word{}, false, err
	}
	return rows[0], true, nil
}

// SaveWord insère (ID = 0) ou met à jour un mot ; renvoie son id.
func (s *Store) SaveWord(ctx context.Context, profileID int64, w Word) (int64, error) {
	card, err := json.Marshal(w.Card)
	if err != nil {
		return 0, err
	}
	if w.ID == 0 {
		res, err := s.DB.ExecContext(ctx, `
			INSERT INTO words (profile_id, word, shown, encounters, errors, avg_time_ms, mastery, next_review, srs_params, last_error_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''))`,
			profileID, w.Word, w.Shown, w.Encounters, w.Errors, w.AvgMsPerChr, w.Mastery, w.NextReview, string(card), w.LastErrorAt)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
	_, err = s.DB.ExecContext(ctx, `
		UPDATE words SET shown = ?, encounters = ?, errors = ?, avg_time_ms = ?, mastery = ?, next_review = ?,
		       srs_params = ?, last_error_at = NULLIF(?, '')
		WHERE id = ? AND profile_id = ?`,
		w.Shown, w.Encounters, w.Errors, w.AvgMsPerChr, w.Mastery, w.NextReview, string(card), w.LastErrorAt, w.ID, profileID)
	return w.ID, err
}

// AddAttempt enregistre une tentative (sessionID = 0 : saisie manuelle).
func (s *Store) AddAttempt(ctx context.Context, wordID, sessionID int64, a typing.WordAttempt, at time.Time) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO word_attempts (word_id, session_id, typed_form, correct, duration_ms, attempted_at, error_type, slow)
		VALUES (?, NULLIF(?, 0), ?, ?, ?, ?, NULLIF(?, ''), ?)`,
		wordID, sessionID, a.Typed, boolInt(a.Correct), a.DurationMs, formatTime(at), a.ErrorType, boolInt(a.Slow))
	return err
}

// ListWords renvoie les mots suivis avec leur détail, les plus fautifs d'abord.
func (s *Store) ListWords(ctx context.Context, profileID int64, f WordFilter) ([]Word, error) {
	where := "WHERE profile_id = ?"
	args := []any{profileID}
	switch f.Status {
	case "due":
		where += " AND mastery < 1 AND next_review <= ?"
		args = append(args, f.Today)
	case "learning":
		where += " AND mastery < 1"
	case "mastered":
		where += " AND mastery >= 1"
	}
	words, err := s.queryWords(ctx, where+" ORDER BY errors DESC, encounters DESC, word", args...)
	if err != nil {
		return nil, err
	}
	for i := range words {
		if err := s.fillDetail(ctx, &words[i]); err != nil {
			return nil, err
		}
	}
	return words, nil
}

// DueCount compte les mots à réviser ce jour (F-43).
func (s *Store) DueCount(ctx context.Context, profileID int64, today string) (int, error) {
	var n int
	err := s.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM words WHERE profile_id = ? AND mastery < 1 AND next_review <= ?", profileID, today).Scan(&n)
	return n, err
}

// DeleteWord retire un mot de la liste (et ses tentatives).
func (s *Store) DeleteWord(ctx context.Context, profileID, id int64) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM words WHERE id = ? AND profile_id = ?", id, profileID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetWordByID renvoie un mot du profil.
func (s *Store) GetWordByID(ctx context.Context, profileID, id int64) (Word, error) {
	rows, err := s.queryWords(ctx, "WHERE profile_id = ? AND id = ?", profileID, id)
	if err != nil {
		return Word{}, err
	}
	if len(rows) == 0 {
		return Word{}, ErrNotFound
	}
	return rows[0], nil
}

// queryWords lit les mots difficiles selon la clause WHERE donnée.
func (s *Store) queryWords(ctx context.Context, where string, args ...any) ([]Word, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, word, COALESCE(shown, word), encounters, errors, avg_time_ms, mastery,
		       COALESCE(next_review, ''), srs_params, COALESCE(last_error_at, '')
		FROM words `+where, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Word{}
	for rows.Next() {
		var w Word
		var avg sql.NullFloat64
		var card string
		if err := rows.Scan(&w.ID, &w.Word, &w.Shown, &w.Encounters, &w.Errors, &avg, &w.Mastery,
			&w.NextReview, &card, &w.LastErrorAt); err != nil {
			return nil, err
		}
		if avg.Valid {
			w.AvgMsPerChr = &avg.Float64
		}
		_ = json.Unmarshal([]byte(card), &w.Card)
		w.Forms = []string{}
		out = append(out, w)
	}
	return out, rows.Err()
}

// fillDetail ajoute à un mot ses fautes les plus fréquentes.
func (s *Store) fillDetail(ctx context.Context, w *Word) error {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT typed_form FROM word_attempts WHERE word_id = ? AND correct = 0
		GROUP BY typed_form ORDER BY COUNT(*) DESC, MAX(attempted_at) DESC LIMIT 5`, w.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			_ = rows.Close()
			return err
		}
		w.Forms = append(w.Forms, f)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	err = s.DB.QueryRowContext(ctx, `
		SELECT error_type FROM word_attempts WHERE word_id = ? AND error_type IS NOT NULL
		GROUP BY error_type ORDER BY COUNT(*) DESC, error_type LIMIT 1`, w.ID).Scan(&w.DominantError)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}
