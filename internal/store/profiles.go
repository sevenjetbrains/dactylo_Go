package store

import (
	"context"
	"database/sql"
	"errors"
)

// DefaultProfileName est le nom du profil créé au premier lancement.
const DefaultProfileName = "Moi"

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
