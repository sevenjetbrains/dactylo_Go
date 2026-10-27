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
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Language   string   `json:"language"`
	Source     string   `json:"source"`     // paste, pdf, builtin
	SourceFile string   `json:"sourceFile"` // nom du fichier importé
	Pages      string   `json:"pages"`      // pages importées, ex. « 3-5 »
	Tags       []string `json:"tags"`
	CreatedAt  string   `json:"createdAt"`
}

// TextSummary est une ligne de liste (sans le contenu complet).
type TextSummary struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Language   string   `json:"language"`
	Source     string   `json:"source"`
	SourceFile string   `json:"sourceFile"`
	Pages      string   `json:"pages"`
	Tags       []string `json:"tags"`
	Length     int      `json:"length"` // en caractères
	Preview    string   `json:"preview"`
	CreatedAt  string   `json:"createdAt"`
	BestWPM    *float64 `json:"bestWpm"`  // meilleur MPM net du profil sur ce texte
	Position   int      `json:"position"` // progression mémorisée (F-24)

	// Contenu complet, pour la recherche ; jamais envoyé à l'interface.
	Content string `json:"-"`
}

// CreateText enregistre un texte et renvoie son id.
func (s *Store) CreateText(ctx context.Context, t Text) (int64, error) {
	t, tags, err := prepare(t)
	if err != nil {
		return 0, err
	}
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO texts (title, content, language, source, source_file, pages, tags)
		 VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?)`,
		t.Title, t.Content, t.Language, t.Source, t.SourceFile, t.Pages, tags)
	if err != nil {
		return 0, fmt.Errorf("enregistrement du texte : %w", err)
	}
	return res.LastInsertId()
}

// UpdateText modifie titre, contenu, langue et étiquettes d'un texte.
func (s *Store) UpdateText(ctx context.Context, t Text) error {
	t, tags, err := prepare(t)
	if err != nil {
		return err
	}
	res, err := s.DB.ExecContext(ctx,
		"UPDATE texts SET title = ?, content = ?, language = ?, tags = ? WHERE id = ?",
		t.Title, t.Content, t.Language, tags, t.ID)
	if err != nil {
		return fmt.Errorf("modification du texte : %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func prepare(t Text) (Text, string, error) {
	if t.Tags == nil {
		t.Tags = []string{}
	}
	tags, err := json.Marshal(t.Tags)
	if t.Language == "" {
		t.Language = "fr"
	}
	if t.Source == "" {
		t.Source = "paste"
	}
	return t, string(tags), err
}

// GetText renvoie un texte complet.
func (s *Store) GetText(ctx context.Context, id int64) (Text, error) {
	var t Text
	var tags string
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, title, content, language, source, COALESCE(source_file, ''), COALESCE(pages, ''), tags, created_at
		FROM texts WHERE id = ?`, id).
		Scan(&t.ID, &t.Title, &t.Content, &t.Language, &t.Source, &t.SourceFile, &t.Pages, &tags, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Text{}, ErrNotFound
	}
	if err != nil {
		return Text{}, err
	}
	t.Tags = decodeTags(tags)
	return t, nil
}

// ListTexts renvoie les textes, du plus récent au plus ancien, avec le
// meilleur score et la progression du profil.
func (s *Store) ListTexts(ctx context.Context, profileID int64) ([]TextSummary, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT t.id, t.title, t.language, t.source, COALESCE(t.source_file, ''), COALESCE(t.pages, ''),
		       t.tags, length(t.content), substr(t.content, 1, 120), t.created_at, t.content,
		       (SELECT MAX(wpm_net) FROM sessions s
		         WHERE s.text_id = t.id AND s.profile_id = ? AND s.ended_at IS NOT NULL),
		       COALESCE((SELECT position FROM text_progress p
		         WHERE p.text_id = t.id AND p.profile_id = ?), 0)
		FROM texts t ORDER BY t.created_at DESC, t.id DESC`, profileID, profileID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []TextSummary{}
	for rows.Next() {
		var t TextSummary
		var tags string
		var best sql.NullFloat64
		if err := rows.Scan(&t.ID, &t.Title, &t.Language, &t.Source, &t.SourceFile, &t.Pages, &tags,
			&t.Length, &t.Preview, &t.CreatedAt, &t.Content, &best, &t.Position); err != nil {
			return nil, err
		}
		t.Tags = decodeTags(tags)
		if best.Valid {
			t.BestWPM = &best.Float64
		}
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

// GetProgress renvoie la position mémorisée du profil dans un texte.
func (s *Store) GetProgress(ctx context.Context, profileID, textID int64) (int, error) {
	var pos int
	err := s.DB.QueryRowContext(ctx,
		"SELECT position FROM text_progress WHERE profile_id = ? AND text_id = ?", profileID, textID).Scan(&pos)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return pos, err
}

// SetProgress mémorise la position du profil dans un texte.
func (s *Store) SetProgress(ctx context.Context, profileID, textID int64, position int) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO text_progress (profile_id, text_id, position, last_opened)
		VALUES (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		ON CONFLICT (profile_id, text_id) DO UPDATE SET position = excluded.position, last_opened = excluded.last_opened`,
		profileID, textID, position)
	return err
}

func decodeTags(raw string) []string {
	tags := []string{}
	_ = json.Unmarshal([]byte(raw), &tags)
	return tags
}
