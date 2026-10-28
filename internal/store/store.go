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

// Backup écrit une copie cohérente de la base dans path (F-52), même en mode
// WAL et pendant l'utilisation.
func (s *Store) Backup(ctx context.Context, path string) error {
	_ = os.Remove(path) // VACUUM INTO refuse un fichier existant
	if _, err := s.DB.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return fmt.Errorf("sauvegarde : %w", err)
	}
	return nil
}

// CheckBackup vérifie qu'un fichier est une sauvegarde Dactylo lisible et
// pas plus récente que l'application ; renvoie sa version de schéma.
func CheckBackup(ctx context.Context, path string, maxVersion int64) (int64, error) {
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		return 0, err
	}
	defer func() { _ = db.Close() }()
	var v int64
	err = db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version_id), 0) FROM "+MigrationsTable+" WHERE is_applied = 1").Scan(&v)
	if err != nil {
		return 0, fmt.Errorf("ce fichier n'est pas une sauvegarde Dactylo : %w", err)
	}
	var n int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM profiles").Scan(&n); err != nil || n == 0 {
		return 0, fmt.Errorf("sauvegarde incomplète (aucun profil)")
	}
	if v > maxVersion {
		return 0, fmt.Errorf("sauvegarde créée par une version plus récente de Dactylo (schéma v%d)", v)
	}
	return v, nil
}

// Restore remplace la base par la sauvegarde `from` (F-52) : copie de
// sécurité de la base actuelle (`.bak`), puis réouverture et migrations
// (une sauvegarde plus ancienne est mise à niveau).
func (s *Store) Restore(ctx context.Context, from string) error {
	if s.Path == ":memory:" {
		return fmt.Errorf("restauration impossible sur une base en mémoire")
	}
	current, err := s.SchemaVersion(ctx)
	if err != nil {
		return err
	}
	if _, err := CheckBackup(ctx, from, current); err != nil {
		return err
	}
	if err := s.Backup(ctx, s.Path+".bak"); err != nil {
		return err
	}
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	if err := s.DB.Close(); err != nil {
		return err
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		_ = os.Remove(s.Path + suffix)
	}
	fresh, err := s.replaceAndOpen(ctx, data)
	if err != nil {
		// échec : on remet la base d'origine depuis la copie de sécurité
		if old, readErr := os.ReadFile(s.Path + ".bak"); readErr == nil {
			if back, reopenErr := s.replaceAndOpen(ctx, old); reopenErr == nil {
				s.DB = back.DB
			}
		}
		return fmt.Errorf("restauration : %w", err)
	}
	s.DB = fresh.DB
	return nil
}

func (s *Store) replaceAndOpen(ctx context.Context, data []byte) (*Store, error) {
	if err := os.WriteFile(s.Path, data, 0o644); err != nil {
		return nil, err
	}
	return Open(ctx, s.Path)
}

// Close ferme la connexion.
func (s *Store) Close() error {
	return s.DB.Close()
}
