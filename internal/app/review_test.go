package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dactylo/internal/core/events"
	"dactylo/internal/core/exercise"
	"dactylo/internal/core/typing"
)

type reviewFixture struct {
	*fixture
	review *ReviewService
	stats  *StatsService
	clock  time.Time
}

func newReviewFixture(t *testing.T) *reviewFixture {
	t.Helper()
	f := &reviewFixture{fixture: newFixture(t), clock: time.Date(2026, 10, 27, 18, 0, 0, 0, time.Local)}
	f.review = NewReviewService(f.store, f.bus, f.practice.profileID)
	f.review.now = func() time.Time { return f.clock }
	f.practice.now = func() time.Time { return f.clock } // même horloge partout : test indépendant de l'heure réelle
	f.stats = NewStatsService(f.store, f.practice.profileID, f.picker)
	f.stats.now = func() time.Time { return f.clock }
	return f
}

// typeSession tape `typed` sur le texte `content` (100 ms par frappe) et termine.
func (f *reviewFixture) typeSession(t *testing.T, req StartRequest, typed string) SessionResult {
	t.Helper()
	sid, err := f.practice.StartSession(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.practice.RecordKeystrokes(sid, timedKeys(typed, 100)); err != nil {
		t.Fatal(err)
	}
	res, err := f.practice.FinishSession(sid, int64(len([]rune(typed)))*100)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func timedKeys(s string, step int64) []typing.Keystroke {
	var ks []typing.Keystroke
	for i, r := range []rune(s) {
		ks = append(ks, typing.Keystroke{Seq: i, Position: i, Typed: string(r), TimeMs: int64(i+1) * step})
	}
	return ks
}

func dueWords(t *testing.T, f *reviewFixture) []string {
	t.Helper()
	ws, err := f.review.ListWords("due")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, w := range ws {
		out = append(out, w.Word)
	}
	return out
}

// Critère d'acceptation transverse (section 7.3).
func TestMistypedWordReviewedTomorrowAndGoneAfterThreeSuccesses(t *testing.T) {
	f := newReviewFixture(t)
	id, _ := f.texts.CreateText(TextInput{Content: "Le maître dort sous le chêne."})

	f.typeSession(t, StartRequest{TextID: id}, "Le maitre dort sous le chêne.")
	if got := dueWords(t, f); len(got) != 0 {
		t.Errorf("le jour même : %v", got)
	}
	f.clock = f.clock.AddDate(0, 0, 1)
	if got := dueWords(t, f); len(got) != 1 || got[0] != "maître" {
		t.Fatalf("le lendemain : %v", got)
	}
	if s, _ := f.review.Summary(); s.Due != 1 || s.Learning != 1 {
		t.Errorf("résumé = %+v", s)
	}

	for i := 1; i <= 3; i++ {
		ex, err := f.review.BuildExercise(exercise.KindShuffled, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(ex.Content, "maître") || ex.Words[0] != "maître" {
			t.Fatalf("exercice = %+v", ex)
		}
		// une seule occurrence suffit : un mot compte une fois par session
		res := f.typeSession(t, StartRequest{Content: "maître est là"}, "maître est là")
		if res.Summary.Accuracy != 100 {
			t.Fatalf("révision %d : %+v", i, res.Summary)
		}
		f.clock = f.clock.AddDate(0, 0, 30)
	}
	if got := dueWords(t, f); len(got) != 0 {
		t.Errorf("après 3 réussites : %v", got)
	}
	ws, _ := f.review.ListWords("mastered")
	if len(ws) != 1 || ws[0].Encounters != 4 || ws[0].Errors != 1 || ws[0].DominantError != typing.ErrAccent ||
		len(ws[0].Forms) != 1 || ws[0].Forms[0] != "maitre" {
		t.Errorf("mot maîtrisé = %+v", ws)
	}
}

func TestSlowWordEntersListAndManualActions(t *testing.T) {
	f := newReviewFixture(t)
	id, _ := f.texts.CreateText(TextInput{Content: "un deux trois quatre"})
	sid, _ := f.practice.StartSession(StartRequest{TextID: id, Settings: `{"slowThreshold":1.5}`})
	var ks []typing.Keystroke
	ts := int64(0)
	for i, r := range []rune("un deux trois quatre") {
		step := int64(100)
		if i >= 8 && i < 13 { // « trois » à 500 ms par frappe
			step = 500
		}
		ts += step
		ks = append(ks, typing.Keystroke{Seq: i, Position: i, Typed: string(r), TimeMs: ts})
	}
	_ = f.practice.RecordKeystrokes(sid, ks)
	if _, err := f.practice.FinishSession(sid, ts); err != nil {
		t.Fatal(err)
	}
	all, _ := f.review.ListWords("all")
	if len(all) != 1 || all[0].Word != "trois" || all[0].Errors != 0 {
		t.Fatalf("mots lents = %+v", all)
	}

	if err := f.review.AddWord("  «Œuvre», "); err != nil {
		t.Fatal(err)
	}
	if err := f.review.AddWord("12"); err == nil {
		t.Error("un nombre n'est pas un mot")
	}
	if got := dueWords(t, f); len(got) != 1 || got[0] != "œuvre" {
		t.Errorf("ajout manuel : à réviser aujourd'hui, %v", got)
	}
	all, _ = f.review.ListWords("all")
	for _, w := range all {
		if w.Word == "œuvre" {
			if w.Shown != "Œuvre" {
				t.Errorf("forme affichée = %q", w.Shown)
			}
			if err := f.review.MarkMastered(w.ID); err != nil {
				t.Fatal(err)
			}
		} else if err := f.review.RemoveWord(w.ID); err != nil {
			t.Fatal(err)
		}
	}
	if s, _ := f.review.Summary(); s.Mastered != 1 || s.Learning != 0 || s.Due != 0 {
		t.Errorf("résumé = %+v", s)
	}
	if err := f.review.MarkMastered(999); err == nil {
		t.Error("mot absent")
	}
	if _, err := f.review.BuildExercise(exercise.KindShuffled, 0); err == nil {
		t.Error("exercice sans mot à réviser")
	}
}

func TestBuildSentencesAndReviewSession(t *testing.T) {
	f := newReviewFixture(t)
	_ = f.review.AddWord("corbeau")
	ex, err := f.review.BuildExercise(exercise.KindSentences, 0)
	if err != nil || ex.Kind != exercise.KindSentences || !strings.Contains(ex.Content, "Corbeau") {
		t.Fatalf("phrases = %+v, %v", ex, err)
	}
	ex, _ = f.review.BuildExercise(exercise.KindRepeated, 3)
	if ex.Content != "corbeau corbeau corbeau" {
		t.Errorf("répété = %q", ex.Content)
	}

	// session de révision : sans texte, mode review
	res := f.typeSession(t, StartRequest{Content: ex.Content}, "corbeau corbeau corbeau")
	se, _ := f.store.GetSession(t.Context(), res.SessionID)
	if se.Mode != ModeReview || se.TextID.Valid || !strings.Contains(se.Settings, `"content":"corbeau corbeau corbeau"`) {
		t.Errorf("session de révision = %+v", se)
	}
	if _, err := f.practice.StartSession(StartRequest{Content: "  "}); err == nil {
		t.Error("exercice vide accepté")
	}
}

func TestDashboardAndCSV(t *testing.T) {
	f := newReviewFixture(t)
	id, _ := f.texts.CreateText(TextInput{Title: "Essai; avec point-virgule", Content: "quoi que qui"})
	f.typeSession(t, StartRequest{TextID: id}, "quoi que qui")
	f.typeSession(t, StartRequest{TextID: id}, "quoi qxe qui")

	d, err := f.stats.Dashboard()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Points) != 2 || d.Totals.Sessions != 2 || d.Totals.TotalMs != 2400 || d.Streak != 1 {
		t.Errorf("tableau de bord = %+v", d)
	}
	if len(d.Keys) == 0 || len(d.Bigrams) == 0 || d.Bigrams[0].Bigram == "" {
		t.Errorf("clavier = %+v / %+v", d.Keys, d.Bigrams)
	}

	csv, err := SessionsCSV(d.Points)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(csv)), "\n")
	if !bytes.HasPrefix(csv, []byte("\uFEFFdate;mode;texte")) || len(lines) != 3 ||
		!strings.Contains(lines[1], `"Essai; avec point-virgule"`) || !strings.Contains(lines[1], "100,0") {
		t.Errorf("CSV =\n%s", csv)
	}

	out := filepath.Join(t.TempDir(), "export.csv")
	f.picker.path = out
	res, err := f.stats.ExportCSV("Exporter")
	if err != nil || res.Path != out || res.Sessions != 2 {
		t.Fatalf("export = %+v, %v", res, err)
	}
	if data, _ := os.ReadFile(out); !bytes.Equal(data, csv) {
		t.Error("fichier exporté différent")
	}
	f.picker.path = ""
	if res, err := f.stats.ExportCSV("Exporter"); err != nil || res.Path != "" {
		t.Errorf("export annulé = %+v, %v", res, err)
	}
}

// Les tentatives de mots sont jointes à SessionCompleted pour les modules abonnés.
func TestSessionCompletedCarriesWords(t *testing.T) {
	f := newFixture(t)
	id, _ := f.texts.CreateText(TextInput{Content: "le chat"})
	f.events = nil
	sid, _ := f.practice.StartSession(StartRequest{TextID: id})
	_ = f.practice.RecordKeystrokes(sid, timedKeys("le chot", 100))
	_, _ = f.practice.FinishSession(sid, 700)
	p, ok := f.events[0].Payload.(events.SessionCompletedPayload)
	if !ok || p.Mode != ModeCopy {
		t.Fatalf("événement = %+v", f.events[0])
	}
	if len(p.Words) != 2 || p.Words[1].Word != "chat" || p.Words[1].Correct {
		t.Errorf("mots = %+v", p.Words)
	}
}
