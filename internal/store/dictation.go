// Accès aux données de la dictée : profils de lecture et moyennes.

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// DictationProfile est un jeu de réglages de dictée enregistré (F-17).
type DictationProfile struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Voice            string  `json:"voice"`
	Rate             float64 `json:"rate"`
	Pitch            float64 `json:"pitch"`
	Volume           float64 `json:"volume"`
	Chunking         string  `json:"chunking"`
	GroupSize        int     `json:"groupSize"`
	PauseMs          int     `json:"pauseMs"`
	SpeakPunctuation bool    `json:"speakPunctuation"`
	Advance          string  `json:"advance"`    // adaptive ou auto
	MaxRepeats       int     `json:"maxRepeats"` // 0 = illimité
}

// ListDictationProfiles renvoie les profils, par nom.
func (s *Store) ListDictationProfiles(ctx context.Context) ([]DictationProfile, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, COALESCE(voice, ''), rate, pitch, volume, chunking, group_size, pause_ms,
		       speak_punctuation, advance, max_repeats
		FROM dictation_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []DictationProfile{}
	for rows.Next() {
		var p DictationProfile
		var punct int
		if err := rows.Scan(&p.ID, &p.Name, &p.Voice, &p.Rate, &p.Pitch, &p.Volume, &p.Chunking,
			&p.GroupSize, &p.PauseMs, &punct, &p.Advance, &p.MaxRepeats); err != nil {
			return nil, err
		}
		p.SpeakPunctuation = punct == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

// SaveDictationProfile insère (ID = 0) ou met à jour un profil ; renvoie son id.
func (s *Store) SaveDictationProfile(ctx context.Context, p DictationProfile) (int64, error) {
	args := []any{p.Name, p.Voice, p.Rate, p.Pitch, p.Volume, p.Chunking, p.GroupSize, p.PauseMs,
		boolInt(p.SpeakPunctuation), p.Advance, p.MaxRepeats}
	if p.ID == 0 {
		res, err := s.DB.ExecContext(ctx, `
			INSERT INTO dictation_profiles (name, voice, rate, pitch, volume, chunking, group_size, pause_ms,
			                                speak_punctuation, advance, max_repeats)
			VALUES (?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
		if err != nil {
			return 0, fmt.Errorf("enregistrement du profil : %w", err)
		}
		return res.LastInsertId()
	}
	res, err := s.DB.ExecContext(ctx, `
		UPDATE dictation_profiles SET name = ?, voice = NULLIF(?, ''), rate = ?, pitch = ?, volume = ?,
		       chunking = ?, group_size = ?, pause_ms = ?, speak_punctuation = ?, advance = ?, max_repeats = ?
		WHERE id = ?`, append(args, p.ID)...)
	if err != nil {
		return 0, fmt.Errorf("enregistrement du profil : %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return 0, ErrNotFound
	}
	return p.ID, nil
}

// DeleteDictationProfile supprime un profil.
func (s *Store) DeleteDictationProfile(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM dictation_profiles WHERE id = ?", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetSessionSettings remplace les réglages (JSON) d'une session : la dictée
// y conserve le texte tapé (données brutes, section 3.1).
func (s *Store) SetSessionSettings(ctx context.Context, id int64, settings string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE sessions SET settings = ? WHERE id = ?", settings, id)
	return err
}

// DictationAverages est la moyenne des dernières dictées (écran de correction).
type DictationAverages struct {
	Sessions int     `json:"sessions"`
	Accuracy float64 `json:"accuracy"`
	Faults   float64 `json:"faults"`
	WPM      float64 `json:"wpm"`
}

// RecentDictationAverages calcule la moyenne des `limit` dernières dictées
// terminées, hors session `excludeID`.
func (s *Store) RecentDictationAverages(ctx context.Context, profileID, excludeID int64, limit int) (DictationAverages, error) {
	var a DictationAverages
	var acc, faults, wpm sql.NullFloat64
	err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*), AVG(accuracy), AVG(faults), AVG(wpm_net) FROM (
			SELECT s.accuracy, s.wpm_net,
			       (SELECT COUNT(*) FROM session_errors e WHERE e.session_id = s.id) AS faults
			FROM sessions s
			WHERE s.profile_id = ? AND s.mode = 'dictation' AND s.id <> ? AND s.ended_at IS NOT NULL
			ORDER BY s.ended_at DESC LIMIT ?)`, profileID, excludeID, limit).
		Scan(&a.Sessions, &acc, &faults, &wpm)
	if errors.Is(err, sql.ErrNoRows) {
		return a, nil
	}
	a.Accuracy, a.Faults, a.WPM = acc.Float64, faults.Float64, wpm.Float64
	return a, err
}
