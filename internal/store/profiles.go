package store

import (
	"context"
	"database/sql"
	"errors"
)

// DefaultProfileName est le nom du profil créé au premier lancement.
const DefaultProfileName = "Moi"

// Profile est un profil local (F-50).
type Profile struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	KeyboardLayout string `json:"keyboardLayout"` // azerty, qwerty, bepo
	Settings       string `json:"settings"`       // réglages de l'interface (JSON, F-51)
	CreatedAt      string `json:"createdAt"`
}

const profileColumns = "id, name, keyboard_layout, settings, created_at"

func scanProfile(row interface{ Scan(...any) error }) (Profile, error) {
	var p Profile
	err := row.Scan(&p.ID, &p.Name, &p.KeyboardLayout, &p.Settings, &p.CreatedAt)
	return p, err
}

// GetProfile renvoie un profil.
func (s *Store) GetProfile(ctx context.Context, id int64) (Profile, error) {
	p, err := scanProfile(s.DB.QueryRowContext(ctx, "SELECT "+profileColumns+" FROM profiles WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	return p, err
}

// ListProfiles renvoie les profils, du plus ancien au plus récent.
func (s *Store) ListProfiles(ctx context.Context) ([]Profile, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT "+profileColumns+" FROM profiles ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Profile{}
	for rows.Next() {
		p, err := scanProfile(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// CreateProfile ajoute un profil et renvoie son id.
func (s *Store) CreateProfile(ctx context.Context, name, layout string) (int64, error) {
	res, err := s.DB.ExecContext(ctx, "INSERT INTO profiles (name, keyboard_layout) VALUES (?, ?)", name, layout)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateProfile change le nom et la disposition de clavier d'un profil.
func (s *Store) UpdateProfile(ctx context.Context, id int64, name, layout string) error {
	res, err := s.DB.ExecContext(ctx, "UPDATE profiles SET name = ?, keyboard_layout = ? WHERE id = ?", name, layout, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// RenameProfile change le nom affiché d'un profil.
func (s *Store) RenameProfile(ctx context.Context, id int64, name string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE profiles SET name = ? WHERE id = ?", name, id)
	return err
}

// SaveProfileSettings enregistre les réglages (JSON) d'un profil.
func (s *Store) SaveProfileSettings(ctx context.Context, id int64, settings string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE profiles SET settings = ? WHERE id = ?", settings, id)
	return err
}

// DeleteProfile supprime un profil et toutes ses données (sessions, mots,
// progression : clés étrangères en cascade).
func (s *Store) DeleteProfile(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM profiles WHERE id = ?", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetSetting lit un réglage de l'application (ok = false s'il est absent).
func (s *Store) GetSetting(ctx context.Context, key string) (string, bool, error) {
	var v string
	err := s.DB.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return v, err == nil, err
}

// SetSetting enregistre un réglage de l'application.
func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO app_settings (key, value) VALUES (?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// LastProgress renvoie le texte commencé et non terminé ouvert le plus
// récemment (ok = false s'il n'y en a pas).
func (s *Store) LastProgress(ctx context.Context, profileID int64) (Text, int, bool, error) {
	var textID int64
	var pos int
	err := s.DB.QueryRowContext(ctx, `
		SELECT p.text_id, p.position FROM text_progress p JOIN texts t ON t.id = p.text_id
		WHERE p.profile_id = ? AND p.position > 0 AND p.position < length(t.content)
		ORDER BY p.last_opened DESC LIMIT 1`, profileID).Scan(&textID, &pos)
	if errors.Is(err, sql.ErrNoRows) {
		return Text{}, 0, false, nil
	}
	if err != nil {
		return Text{}, 0, false, err
	}
	t, err := s.GetText(ctx, textID)
	return t, pos, err == nil, err
}

// LastTextForMode renvoie le texte de la dernière session terminée dans ce
// mode, ou à défaut le texte le plus récent (ok = false si la bibliothèque est vide).
func (s *Store) LastTextForMode(ctx context.Context, profileID int64, mode string) (Text, bool, error) {
	var id int64
	err := s.DB.QueryRowContext(ctx, `
		SELECT text_id FROM sessions
		WHERE profile_id = ? AND mode = ? AND ended_at IS NOT NULL AND text_id IS NOT NULL
		ORDER BY ended_at DESC LIMIT 1`, profileID, mode).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		err = s.DB.QueryRowContext(ctx, "SELECT id FROM texts ORDER BY created_at DESC, id DESC LIMIT 1").Scan(&id)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return Text{}, false, nil
	}
	if err != nil {
		return Text{}, false, err
	}
	t, err := s.GetText(ctx, id)
	return t, err == nil, err
}

// EnsureDefaultProfile renvoie l'id du premier profil, en le créant s'il
// n'en existe aucun. Les profils multiples arrivent en Phase 5 (F-50).
func (s *Store) EnsureDefaultProfile(ctx context.Context) (int64, error) {
	var id int64
	err := s.DB.QueryRowContext(ctx, "SELECT id FROM profiles ORDER BY id LIMIT 1").Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	res, err := s.DB.ExecContext(ctx, "INSERT INTO profiles (name) VALUES (?)", DefaultProfileName)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
