package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrNotFound signale un enregistrement absent.
var ErrNotFound = errors.New("introuvable")

// Text est un texte de la bibliothèque.
type Text struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Language  string   `json:"language"`
	Source    string   `json:"source"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"createdAt"`
}

// TextSummary est une ligne de liste (sans le contenu complet).
type TextSummary struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Language  string   `json:"language"`
	Source    string   `json:"source"`
	Tags      []string `json:"tags"`
	Length    int      `json:"length"` // en caractères
	Preview   string   `json:"preview"`
	CreatedAt string   `json:"createdAt"`
}

// CreateText enregistre un texte et renvoie son id.
func (s *Store) CreateText(ctx context.Context, t Text) (int64, error) {
	if t.Tags == nil {
		t.Tags = []string{}
	}
	tags, err := json.Marshal(t.Tags)
	if err != nil {
		return 0, err
	}
	if t.Language == "" {
		t.Language = "fr"
	}
	if t.Source == "" {
		t.Source = "paste"
	}
	res, err := s.DB.ExecContext(ctx,
		"INSERT INTO texts (title, content, language, source, tags) VALUES (?, ?, ?, ?, ?)",
		t.Title, t.Content, t.Language, t.Source, string(tags))
	if err != nil {
		return 0, fmt.Errorf("enregistrement du texte : %w", err)
	}
	return res.LastInsertId()
}

// GetText renvoie un texte complet.
func (s *Store) GetText(ctx context.Context, id int64) (Text, error) {
	var t Text
	var tags string
	err := s.DB.QueryRowContext(ctx,
		"SELECT id, title, content, language, source, tags, created_at FROM texts WHERE id = ?", id).
		Scan(&t.ID, &t.Title, &t.Content, &t.Language, &t.Source, &tags, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Text{}, ErrNotFound
	}
	if err != nil {
		return Text{}, err
	}
	t.Tags = decodeTags(tags)
	return t, nil
}

// ListTexts renvoie les textes, du plus récent au plus ancien.
func (s *Store) ListTexts(ctx context.Context) ([]TextSummary, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, title, language, source, tags, length(content), substr(content, 1, 120), created_at
		FROM texts ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []TextSummary{}
	for rows.Next() {
		var t TextSummary
		var tags string
		if err := rows.Scan(&t.ID, &t.Title, &t.Language, &t.Source, &tags, &t.Length, &t.Preview, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Tags = decodeTags(tags)
		out = append(out, t)
	}
	return out, rows.Err()
}

// DeleteText supprime un texte ; ses sessions sont conservées (text_id NULL).
func (s *Store) DeleteText(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM texts WHERE id = ?", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func decodeTags(raw string) []string {
	tags := []string{}
	_ = json.Unmarshal([]byte(raw), &tags)
	return tags
}
