package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"dactylo/internal/core/typing"
)

func TestProfilesCRUDAndSettings(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	first, _ := s.EnsureDefaultProfile(ctx)
	kid, err := s.CreateProfile(ctx, "Léa", "bepo")
	if err != nil {
		t.Fatal(err)
	}
	list, _ := s.ListProfiles(ctx)
	if len(list) != 2 || list[0].ID != first || list[1].Name != "Léa" || list[1].KeyboardLayout != "bepo" || list[1].Settings != "{}" {
		t.Fatalf("profils = %+v", list)
	}
	if err := s.UpdateProfile(ctx, kid, "Léa M.", "qwerty"); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveProfileSettings(ctx, kid, `{"theme":"dark"}`); err != nil {
		t.Fatal(err)
	}
	p, _ := s.GetProfile(ctx, kid)
	if p.Name != "Léa M." || p.KeyboardLayout != "qwerty" || p.Settings != `{"theme":"dark"}` {
		t.Errorf("profil = %+v", p)
	}
	if err := s.UpdateProfile(ctx, 999, "x", "azerty"); !errors.Is(err, ErrNotFound) {
		t.Errorf("profil absent : %v", err)
	}

	// suppression en cascade : sessions et mots du profil
	tid, _ := s.CreateText(ctx, Text{Title: "t", Content: "ab"})
	sid, _ := s.CreateSession(ctx, kid, tid, "copy", "", time.Now())
	_, _ = s.SaveWord(ctx, kid, Word{Word: "chat", Shown: "chat"})
	if err := s.DeleteProfile(ctx, kid); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetSession(ctx, sid); !errors.Is(err, ErrNotFound) {
		t.Errorf("session du profil supprimé : %v", err)
	}
	if ws, _ := s.ListWords(ctx, kid, WordFilter{Status: "all"}); len(ws) != 0 {
		t.Errorf("mots du profil supprimé : %v", ws)
	}
	if err := s.DeleteProfile(ctx, kid); !errors.Is(err, ErrNotFound) {
		t.Errorf("double suppression : %v", err)
	}

	// réglages de l'application
	if _, ok, _ := s.GetSetting(ctx, "active_profile"); ok {
		t.Error("réglage absent attendu")
	}
	_ = s.SetSetting(ctx, "active_profile", "1")
	_ = s.SetSetting(ctx, "active_profile", "2")
	if v, ok, _ := s.GetSetting(ctx, "active_profile"); !ok || v != "2" {
		t.Errorf("réglage = %q, %v", v, ok)
	}
}

func TestInterruptedSessions(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	pid, _ := s.EnsureDefaultProfile(ctx)
	tid, _ := s.CreateText(ctx, Text{Title: "Fable", Content: "abc"})
	now := time.Now()
	key := []typing.Scored{{Keystroke: typing.Keystroke{Seq: 0, Position: 0, Typed: "a"}, Expected: "a", Correct: true}}

	empty, _ := s.CreateSession(ctx, pid, tid, "copy", "", now)     // sans frappe : ignorée
	abandoned, _ := s.CreateSession(ctx, pid, tid, "copy", "", now) // abandonnée : ignorée
	crashed, _ := s.CreateSession(ctx, pid, tid, "copy", `{"x":1}`, now.Add(-time.Second))
	_ = s.SaveKeystrokes(ctx, abandoned, key)
	_ = s.SaveKeystrokes(ctx, crashed, key)
	_ = s.AbandonSession(ctx, abandoned, now)
	_ = empty

	x, ok, err := s.LastInterrupted(ctx, pid, now.Add(-time.Hour))
	if err != nil || !ok || x.ID != crashed || x.TextTitle != "Fable" || x.Keystrokes != 1 || x.Settings != `{"x":1}` {
		t.Fatalf("interrompue = %+v, %v, %v", x, ok, err)
	}
	if _, ok, _ := s.LastInterrupted(ctx, pid, now.Add(time.Hour)); ok {
		t.Error("trop ancienne : ignorée")
	}
	_ = s.AbandonSession(ctx, crashed, now)
	if _, ok, _ := s.LastInterrupted(ctx, pid, now.Add(-time.Hour)); ok {
		t.Error("abandonnée ensuite : ignorée")
	}
}

func TestBackupAndRestore(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(ctx, filepath.Join(dir, "dactylo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	pid, _ := s.EnsureDefaultProfile(ctx)
	_ = s.RenameProfile(ctx, pid, "Avant")

	backup := filepath.Join(dir, "sauvegarde.db")
	if err := s.Backup(ctx, backup); err != nil {
		t.Fatal(err)
	}
	if err := s.Backup(ctx, backup); err != nil { // fichier existant : remplacé
		t.Fatal(err)
	}
	_ = s.RenameProfile(ctx, pid, "Après")
	if _, err := s.CreateText(ctx, Text{Title: "nouveau", Content: "x"}); err != nil {
		t.Fatal(err)
	}

	if err := s.Restore(ctx, backup); err != nil {
		t.Fatal(err)
	}
	if p, _ := s.GetProfile(ctx, pid); p.Name != "Avant" {
		t.Errorf("après restauration : %q", p.Name)
	}
	if list, _ := s.ListTexts(ctx, pid); len(list) != 8 { // pack intégré seulement
		t.Errorf("textes restaurés = %d", len(list))
	}
	if _, err := os.Stat(filepath.Join(dir, "dactylo.db.bak")); err != nil {
		t.Errorf("copie de sécurité absente : %v", err)
	}

	// fichiers refusés
	bad := filepath.Join(dir, "faux.db")
	_ = os.WriteFile(bad, []byte("pas une base"), 0o600)
	if err := s.Restore(ctx, bad); err == nil {
		t.Error("fichier invalide accepté")
	}
	if p, err := s.GetProfile(ctx, pid); err != nil || p.Name != "Avant" {
		t.Errorf("base intacte après un refus : %+v, %v", p, err)
	}
	if _, err := CheckBackup(ctx, backup, 1); err == nil {
		t.Error("sauvegarde plus récente que l'application acceptée")
	}
	mem := openMem(t)
	if err := mem.Restore(ctx, backup); err == nil {
		t.Error("restauration en mémoire acceptée")
	}
}
