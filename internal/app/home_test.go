package app

import (
	"strings"
	"testing"
	"time"

	"dactylo/internal/core/text"
)

func TestHomeOverview(t *testing.T) {
	f := newReviewFixture(t)
	home := NewHomeService(f.store, f.practice.profileID)
	home.now = func() time.Time { return f.clock }

	// accueil vide
	o, err := home.Overview(text.SplitParagraph, 1)
	if err != nil {
		t.Fatal(err)
	}
	if o.Profile.Name != "Moi" || o.Due.Count != 0 || o.Continue != nil || o.Speed.Value != 0 || len(o.Recent) != 0 {
		t.Fatalf("accueil vide = %+v", o)
	}

	// un texte de 3 paragraphes, le 1er tapé avec une faute, puis un mot à réviser
	id, _ := f.texts.CreateText(TextInput{Title: "Trois", Content: "Un chat noir.\n\nDeux chiens.\n\nTrois oiseaux.", Tags: []string{"Verne"}})
	f.typeSession(t, StartRequest{TextID: id, Start: 0, End: 13}, "Un chot noir.")
	_ = f.review.AddWord("oiseau")

	o, err = home.Overview(text.SplitParagraph, 1)
	if err != nil {
		t.Fatal(err)
	}
	if o.Due.Count != 1 || o.Due.Words[0] != "oiseau" || o.Due.Minutes != 1 {
		t.Errorf("révision du jour = %+v", o.Due)
	}
	c := o.Continue
	if c == nil || c.TextID != id || c.Exercise != 2 || c.Exercises != 3 || !strings.HasPrefix(c.Excerpt, "Deux chiens.") ||
		c.Progress <= 0 || c.Progress >= 1 || c.Tags[0] != "Verne" {
		t.Errorf("reprise = %+v", c)
	}
	if len(o.Recent) != 1 || o.Recent[0].TextTitle != "Trois" || o.Speed.HasDelta || o.Speed.Value <= 0 {
		t.Errorf("sessions = %+v / %+v", o.Recent, o.Speed)
	}
	if o.Streak != 1 || o.BestStreak != 1 || o.WeekSessions != 1 || o.WeekMs != 1300 {
		t.Errorf("série / semaine = %d %d %d %d", o.Streak, o.BestStreak, o.WeekSessions, o.WeekMs)
	}

	// découpage propre au texte : prioritaire sur le réglage général
	if err := f.texts.UpdateText(id, TextInput{Title: "Trois", Content: "Un chat noir.\n\nDeux chiens.\n\nTrois oiseaux.", Split: text.SplitNone}); err != nil {
		t.Fatal(err)
	}
	if o, _ = home.Overview(text.SplitParagraph, 1); o.Continue.Exercises != 1 {
		t.Errorf("découpage du texte ignoré : %+v", o.Continue)
	}
}

func TestTrend(t *testing.T) {
	if tr := trend(nil, nil); tr.HasDelta || tr.Value != 0 {
		t.Errorf("vide = %+v", tr)
	}
}

func TestProfileRename(t *testing.T) {
	f := newFixture(t)
	p := NewProfileService(f.store, f.practice.profileID)
	got, err := p.Rename("  Boualem ")
	if err != nil || got.Name != "Boualem" || got.KeyboardLayout != "azerty" {
		t.Fatalf("Rename = %+v, %v", got, err)
	}
	if got, _ = p.Rename(""); got.Name != "Moi" {
		t.Errorf("nom vide = %q", got.Name)
	}
	if got, _ = p.Rename(strings.Repeat("é", 60)); len([]rune(got.Name)) != 40 {
		t.Errorf("nom tronqué = %d", len([]rune(got.Name)))
	}
}

func TestListTextsExercisesAndSplit(t *testing.T) {
	f := newFixture(t)
	content := "A un.\n\nB deux.\n\nC trois."
	id, _ := f.texts.CreateText(TextInput{Title: "Découpé", Content: content, Split: text.SplitLines, SplitSize: 2})
	if _, err := f.texts.CreateText(TextInput{Content: "x", Split: "pages"}); err == nil {
		t.Error("découpage inconnu accepté")
	}
	list, _ := f.texts.ListTexts(TextQuery{Search: "Découpé", DefaultSplit: text.SplitParagraph, DefaultSize: 1})
	if len(list.Texts) != 1 || list.Texts[0].Exercises != 2 || list.Texts[0].Split != text.SplitLines {
		t.Errorf("liste = %+v", list.Texts)
	}
	pdf, _ := f.texts.CreateText(TextInput{Title: "Importé", Content: "x y", Source: "pdf"})
	if l, _ := f.texts.ListTexts(TextQuery{Source: "pdf"}); len(l.Texts) != 1 || l.Texts[0].ID != pdf {
		t.Errorf("filtre par provenance = %+v", l.Texts)
	}
	plan, _ := f.texts.Exercises(id, text.SplitParagraph, 1)
	if len(plan.Segments) != 2 {
		t.Errorf("Exercises suit le découpage du texte : %d", len(plan.Segments))
	}
	if n := f.texts.CountExercises(content, text.SplitParagraph, 1); n != 3 {
		t.Errorf("CountExercises = %d", n)
	}
	dup, _ := f.texts.DuplicateText(id)
	if d, _ := f.texts.GetText(dup); d.Split != text.SplitLines || d.SplitSize != 2 {
		t.Errorf("copie = %+v", d)
	}
}
