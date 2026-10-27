// Package store gère la base SQLite locale (modernc.org/sqlite, sans CGO)
// et l'application automatique des migrations goose embarquées (NF-05).
package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // pilote "sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationsTable est le nom de la table de suivi des migrations.
const MigrationsTable = "schema_migrations"

// Store encapsule la connexion à la base.
type Store struct {
	DB   *sql.DB
	Path string
}

// DefaultPath renvoie l'emplacement de la base dans le dossier utilisateur.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("dossier utilisateur introuvable : %w", err)
	}
	return filepath.Join(dir, "Dactylo", "dactylo.db"), nil
}

// Open ouvre (ou crée) la base au chemin donné et applique les migrations.
// Le chemin spécial ":memory:" ouvre une base en mémoire (tests).
func Open(ctx context.Context, path string) (*Store, error) {
	dsn := ":memory:"
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("création du dossier de données : %w", err)
		}
		dsn = "file:" + filepath.ToSlash(path)
	}
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout(5000)")
	if path != ":memory:" {
		q.Add("_pragma", "journal_mode(WAL)")
	}
	db, err := sql.Open("sqlite", dsn+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("ouverture de la base : %w", err)
	}
	// SQLite n'accepte qu'un écrivain ; une seule connexion évite aussi
	// qu'une base ":memory:" soit dupliquée par connexion.
	db.SetMaxOpenConns(1)

	s := &Store{DB: db, Path: path}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate(ctx context.Context) error {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	p, err := goose.NewProvider(goose.DialectSQLite3, s.DB, sub, goose.WithTableName(MigrationsTable))
	if err != nil {
		return fmt.Errorf("initialisation des migrations : %w", err)
	}
	if _, err := p.Up(ctx); err != nil {
		return fmt.Errorf("application des migrations : %w", err)
	}
	return nil
}

// SchemaVersion renvoie la version de la dernière migration appliquée.
func (s *Store) SchemaVersion(ctx context.Context) (int64, error) {
	var v int64
	err := s.DB.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version_id), 0) FROM "+MigrationsTable+" WHERE is_applied = 1").Scan(&v)
	return v, err
}

// Close ferme la connexion.
func (s *Store) Close() error {
	return s.DB.Close()
}
