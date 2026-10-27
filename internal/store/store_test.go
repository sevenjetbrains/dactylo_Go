package store

import (
	"context"
	"path/filepath"
	"testing"

	"dactylo/internal/core/text"
)

var expectedTables = []string{
	"profiles", "texts", "text_progress", "sessions", "keystrokes",
	"session_errors", "words", "word_attempts", "dictation_profiles",
}

func TestOpenAppliesMigrations(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	for _, table := range expectedTables {
		var name string
		err := s.DB.QueryRowContext(ctx,
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q absente : %v", table, err)
		}
	}

	v, err := s.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if v < 1 {
		t.Errorf("version de schéma = %d, attendu >= 1", v)
	}
}

func TestBuiltinTextsAreNFCAndClean(t *testing.T) {
	s := openMem(t)
	rows, err := s.DB.Query("SELECT title, content FROM texts WHERE source = 'builtin'")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	n := 0
	for rows.Next() {
		var title, content string
		if err := rows.Scan(&title, &content); err != nil {
			t.Fatal(err)
		}
		n++
		if content != text.NFC(content) || content != text.CleanForPractice(content) {
			t.Errorf("%q : contenu non normalisé", title)
		}
	}
	if n < 8 {
		t.Errorf("textes intégrés = %d, attendu >= 8", n)
	}
}

func TestForeignKeysEnabled(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.DB.ExecContext(ctx,
		"INSERT INTO sessions (profile_id, mode, started_at) VALUES (999, 'copy', '2026-01-01')")
	if err == nil {
		t.Fatal("insertion avec profile_id inexistant acceptée : clés étrangères désactivées")
	}
}

func TestReopenIsIdempotent(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "sub", "dactylo.db")

	s1, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("premier Open: %v", err)
	}
	if _, err := s1.DB.ExecContext(ctx, "INSERT INTO profiles (name) VALUES ('Moi')"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	v1, _ := s1.SchemaVersion(ctx)
	if err := s1.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s2, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	t.Cleanup(func() { _ = s2.Close() })
	v2, _ := s2.SchemaVersion(ctx)
	if v1 != v2 {
		t.Errorf("version changée à la réouverture : %d -> %d", v1, v2)
	}
	var n int
	if err := s2.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM profiles").Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Errorf("profils après réouverture = %d, attendu 1 (perte de données)", n)
	}
}
