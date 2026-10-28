package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"dactylo/internal/core/events"
	"dactylo/internal/store"
)

func TestProfilesSwitchIsolatesData(t *testing.T) {
	f := newReviewFixture(t)
	prof := f.practice.profile
	ps := NewProfileService(f.store, prof)
	first := prof.ID()

	kid, err := ps.Create("  Léa ", "bepo")
	if err != nil || kid.Name != "Léa" || kid.KeyboardLayout != "bepo" {
		t.Fatalf("Create = %+v, %v", kid, err)
	}
	if _, err := ps.Create("x", "dvorak"); err == nil {
		t.Error("disposition inconnue acceptée")
	}
	if prof.ID() != first {
		t.Error("créer un profil ne doit pas en changer")
	}

	// un mot difficile pour le premier profil
	_ = f.review.AddWord("chêne")
	if s, _ := f.review.Summary(); s.Due != 1 {
		t.Fatalf("profil 1 : %+v", s)
	}
	if _, err := ps.Switch(kid.ID); err != nil {
		t.Fatal(err)
	}
	if prof.ID() != kid.ID {
		t.Fatal("bascule sans effet")
	}
	if s, _ := f.review.Summary(); s.Due != 0 {
		t.Errorf("profil 2 voit les mots du profil 1 : %+v", s)
	}
	if err := ps.SaveSettings(`{"theme":"dark"}`); err != nil {
		t.Fatal(err)
	}
	if err := ps.SaveSettings("{pas du json"); err == nil {
		t.Error("réglages invalides acceptés")
	}
	cur, _ := ps.Get()
	if cur.Settings != `{"theme":"dark"}` {
		t.Errorf("réglages = %s", cur.Settings)
	}
	// réglages par défaut d'un autre profil, sans en changer
	if err := ps.SaveSettingsFor(first, `{"theme":"light"}`); err != nil {
		t.Fatal(err)
	}
	if p, _ := f.store.GetProfile(context.Background(), first); p.Settings != `{"theme":"light"}` || prof.ID() != kid.ID {
		t.Errorf("SaveSettingsFor : %s, actif %d", p.Settings, prof.ID())
	}

	// le profil actif est retrouvé au démarrage suivant
	again, err := RestoreActiveProfile(context.Background(), f.store)
	if err != nil || again.ID() != kid.ID {
		t.Errorf("profil au redémarrage = %v, %v", again.ID(), err)
	}

	// mise à jour, suppression du profil actif → bascule sur l'autre
	if p, err := ps.Update(kid.ID, "Léa M.", "qwerty"); err != nil || p.Name != "Léa M." || p.KeyboardLayout != "qwerty" {
		t.Errorf("Update = %+v, %v", p, err)
	}
	if err := ps.Delete(kid.ID); err != nil {
		t.Fatal(err)
	}
	if prof.ID() != first {
		t.Errorf("après suppression du profil actif : %d", prof.ID())
	}
	if err := ps.Delete(first); err == nil {
		t.Error("suppression du dernier profil acceptée")
	}
	list, _ := ps.List()
	if len(list) != 1 {
		t.Errorf("profils restants = %d", len(list))
	}
	if _, err := ps.Switch(999); err == nil {
		t.Error("bascule vers un profil absent acceptée")
	}
}

func TestInterruptedSessionResume(t *testing.T) {
	f := newReviewFixture(t)
	home := NewHomeService(f.store, f.practice.profile)
	home.now = func() time.Time { return f.clock }
	id, _ := f.texts.CreateText(TextInput{Title: "Long", Content: "un deux trois quatre"})

	// session abandonnée (Ctrl+R) : pas de reprise
	s1, _ := f.practice.StartSession(StartRequest{TextID: id, Settings: `{"foldTypography":true}`})
	_ = f.practice.RecordKeystrokes(s1, timedKeys("un", 100))
	if err := f.practice.Abandon(s1); err != nil {
		t.Fatal(err)
	}
	if o, _ := home.Overview("none", 1); o.Interrupted != nil {
		t.Errorf("session abandonnée proposée : %+v", o.Interrupted)
	}

	// fermeture brutale : frappes enregistrées, session ni finie ni abandonnée
	s2, _ := f.practice.StartSession(StartRequest{TextID: id, Start: 3, End: 20})
	_ = f.practice.RecordKeystrokes(s2, timedKeys("deux tr", 150))
	f.practice = NewPracticeService(f.store, f.bus, f.practice.profile) // « redémarrage »

	o, _ := home.Overview("none", 1)
	if o.Interrupted == nil || o.Interrupted.ID != s2 || o.Interrupted.TextTitle != "Long" || o.Interrupted.Keystrokes != 7 {
		t.Fatalf("session interrompue = %+v", o.Interrupted)
	}
	r, err := f.practice.Resume(s2)
	if err != nil {
		t.Fatal(err)
	}
	if r.Content != "deux trois quatre" || r.Start != 3 || r.End != 20 || len(r.Keystrokes) != 7 || r.ElapsedMs != 1050 {
		t.Errorf("reprise = %+v", r)
	}
	// la frappe continue puis la session se termine normalement
	more := timedKeys("deux trois quatre", 150)[7:]
	for i := range more {
		more[i].Seq = 7 + i
	}
	if err := f.practice.RecordKeystrokes(s2, more); err != nil {
		t.Fatal(err)
	}
	res, err := f.practice.FinishSession(s2, 3000)
	if err != nil || res.Summary.Keystrokes != 17 || res.Summary.Accuracy != 100 {
		t.Errorf("fin après reprise = %+v, %v", res.Summary, err)
	}
	if o, _ := home.Overview("none", 1); o.Interrupted != nil {
		t.Error("session terminée encore proposée")
	}
	if _, err := f.practice.Resume(s2); err == nil {
		t.Error("reprise d'une session terminée acceptée")
	}
	if _, err := f.practice.Resume(999); err == nil {
		t.Error("session absente")
	}
}

func TestBackupService(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := store.Open(ctx, filepath.Join(dir, "dactylo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	prof, _ := RestoreActiveProfile(ctx, s)
	picker := &fakePicker{}
	b := NewBackupService(s, prof, picker)
	b.now = func() time.Time { return time.Date(2026, 10, 28, 9, 0, 0, 0, time.Local) }
	ps := NewProfileService(s, prof)
	texts := NewTextService(s, events.NewBus(), prof, picker)

	// annulation
	if r, err := b.Export("Sauvegarder"); err != nil || r.Path != "" {
		t.Errorf("export annulé = %+v, %v", r, err)
	}
	picker.path = filepath.Join(dir, "sauvegarde.db")
	if r, err := b.Export("Sauvegarder"); err != nil || r.Path != picker.path {
		t.Fatalf("export = %+v, %v", r, err)
	}
	if fi, err := os.Stat(picker.path); err != nil || fi.Size() == 0 {
		t.Fatalf("fichier de sauvegarde : %v", err)
	}

	// après la sauvegarde : un nouveau profil actif et un texte
	kid, _ := ps.Create("Léa", "azerty")
	_, _ = ps.Switch(kid.ID)
	_, _ = texts.CreateText(TextInput{Content: "après la sauvegarde"})

	r, err := b.Restore("Restaurer")
	if err != nil || r.Path != picker.path {
		t.Fatalf("restauration = %+v, %v", r, err)
	}
	if prof.ID() == kid.ID {
		t.Error("le profil actif n'existe plus dans la sauvegarde : il faut en changer")
	}
	if list, _ := ps.List(); len(list) != 1 {
		t.Errorf("profils restaurés = %d", len(list))
	}
	if l, _ := texts.ListTexts(TextQuery{Search: "après la sauvegarde"}); len(l.Texts) != 0 {
		t.Error("texte postérieur à la sauvegarde encore présent")
	}
	picker.path = ""
	if r, err := b.Restore("Restaurer"); err != nil || r.Path != "" {
		t.Errorf("restauration annulée = %+v, %v", r, err)
	}
}
