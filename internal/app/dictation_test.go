package app

import (
	"strings"
	"testing"
	"time"

	"dactylo/internal/core/dictation"
	"dactylo/internal/core/events"
	"dactylo/internal/core/exercise"
	"dactylo/internal/core/typing"
	"dactylo/internal/store"
)

func TestDictationFullSession(t *testing.T) {
	f := newReviewFixture(t)
	d := NewDictationService(f.store, f.bus, f.practice.profileID)
	d.now = func() time.Time { return f.clock }
	content := "En 1815, M. Myriel était évêque de Digne.\n\nC'était un vieillard d'environ soixante-quinze ans."
	id, _ := f.texts.CreateText(TextInput{Title: "Les Misérables", Content: content})

	plan, err := d.Prepare(PrepareRequest{TextID: id, Chunking: dictation.ChunkGroup, GroupSize: 4, SpeakPunctuation: true})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Words != 14 || plan.Title != "Les Misérables" || len(plan.Chunks) < 4 ||
		!strings.Contains(plan.Chunks[0].Spoken, "virgule") {
		t.Fatalf("plan = %+v", plan)
	}

	f.events = nil
	sid, err := d.Start(StartRequest{TextID: id, Settings: `{"rate":0.9}`})
	if err != nil {
		t.Fatal(err)
	}
	typed := "En 1815, M. Myriel était évèque de digne. C'était un viellard d'environs soixante quinze ans."
	res, err := d.Finish(FinishDictation{SessionID: sid, Typed: typed, DurationMs: 60000, Repetitions: 2})
	if err != nil {
		t.Fatal(err)
	}
	r := res.Report
	if r.Words != 14 || r.Faults != 5 || r.ByType[typing.ErrAccent] != 1 || r.ByType[typing.ErrPunctuation] != 1 {
		t.Errorf("correction = %d mots, %d fautes, %v", r.Words, r.Faults, r.ByType)
	}
	if len(res.Mistakes) != 5 || res.WPM <= 0 || res.Repetitions != 2 || res.Average.Sessions != 0 {
		t.Errorf("résultat = %+v", res)
	}

	// session enregistrée en mode dictée, copie conservée
	se, _ := f.store.GetSession(t.Context(), sid)
	if se.Mode != ModeDictation || !strings.Contains(se.Settings, `"typed":"En 1815`) || !strings.Contains(se.Settings, `"repetitions":2`) {
		t.Errorf("session = %+v", se)
	}
	// événements : SessionCompleted puis un WordMistyped par faute
	if len(f.events) != 6 || f.events[0].Name != events.SessionCompleted {
		t.Fatalf("événements = %d", len(f.events))
	}
	// les fautes rejoignent les révisions (« Révision prévue demain »)
	all, _ := f.review.ListWords("all")
	keys := map[string]bool{}
	for _, w := range all {
		keys[w.Word] = true
	}
	for _, k := range []string{"évêque", "digne", "vieillard", "environ", "soixante-quinze"} {
		if !keys[k] {
			t.Errorf("%q absent des révisions : %v", k, keys)
		}
	}
	f.clock = f.clock.AddDate(0, 0, 1)
	if due, _ := f.review.ListWords("due"); len(due) != 5 {
		t.Errorf("à réviser le lendemain = %d", len(due))
	}

	// « Retirer » et « S'entraîner sur ces 5 mots »
	if err := f.review.RemoveWordByKey("Digne."); err != nil {
		t.Fatal(err)
	}
	if err := f.review.RemoveWordByKey("absent"); err != nil {
		t.Errorf("mot absent : %v", err)
	}
	ex, err := f.review.BuildExerciseFor([]string{"évêque", "vieillard", "12"}, exercise.KindRepeated, 2)
	if err != nil || ex.Content == "" || len(ex.Words) != 2 {
		t.Errorf("exercice ciblé = %+v, %v", ex, err)
	}
	if _, err := f.review.BuildExerciseFor([]string{"1"}, exercise.KindShuffled, 0); err == nil {
		t.Error("exercice sans mot accepté")
	}

	// seconde dictée : moyenne sur la première
	sid2, _ := d.Start(StartRequest{TextID: id})
	res2, err := d.Finish(FinishDictation{SessionID: sid2, Typed: content, DurationMs: 50000})
	if err != nil {
		t.Fatal(err)
	}
	if res2.Report.Faults != 0 || res2.Average.Sessions != 1 || res2.Average.Faults != 5 {
		t.Errorf("seconde dictée = %+v", res2)
	}
	// carte Dictée de l'accueil : dernier texte dicté
	home := NewHomeService(f.store, f.practice.profileID)
	if o, _ := home.Overview("paragraph", 1); o.Dictation == nil || o.Dictation.TextID != id || o.Dictation.Words != 14 {
		t.Errorf("carte dictée = %+v", o.Dictation)
	}
}

func TestDictationErrors(t *testing.T) {
	f := newFixture(t)
	d := NewDictationService(f.store, f.bus, f.practice.profileID)
	id, _ := f.texts.CreateText(TextInput{Content: "un deux"})
	if _, err := d.Prepare(PrepareRequest{TextID: 999}); err == nil {
		t.Error("texte absent")
	}
	if _, err := d.Prepare(PrepareRequest{TextID: id, Start: 5, End: 2}); err == nil {
		t.Error("extrait invalide")
	}
	if _, err := d.Start(StartRequest{TextID: id, Settings: "{x"}); err == nil {
		t.Error("réglages invalides")
	}
	if _, err := d.Finish(FinishDictation{SessionID: 42}); err == nil {
		t.Error("session inconnue")
	}
	sid, _ := d.Start(StartRequest{TextID: id})
	if _, err := d.Finish(FinishDictation{SessionID: sid, DurationMs: -1}); err == nil {
		t.Error("durée négative")
	}
	// copie vide : tous les mots oubliés, vitesse nulle
	res, err := d.Finish(FinishDictation{SessionID: sid})
	if err != nil || res.Report.Faults != 2 || res.WPM != 0 {
		t.Errorf("copie vide = %+v, %v", res, err)
	}
}

func TestDictationProfilesService(t *testing.T) {
	f := newFixture(t)
	d := NewDictationService(f.store, f.bus, f.practice.profileID)
	list, _ := d.ListProfiles()
	if len(list) != 2 {
		t.Fatalf("profils intégrés = %d", len(list))
	}
	id, err := d.SaveProfile(store.DictationProfile{Name: "  Très lent ", Rate: 0.1, Volume: 3, Pitch: -1, PauseMs: 20000,
		GroupSize: 50, MaxRepeats: 99, Chunking: dictation.ChunkWord, Advance: "auto"})
	if err != nil {
		t.Fatal(err)
	}
	list, _ = d.ListProfiles()
	var p store.DictationProfile
	for _, x := range list {
		if x.ID == id {
			p = x
		}
	}
	if p.Name != "Très lent" || p.Rate != 0.5 || p.Volume != 1 || p.Pitch != 0 || p.PauseMs != 10000 ||
		p.GroupSize != 20 || p.MaxRepeats != 10 {
		t.Errorf("profil borné = %+v", p)
	}
	for _, bad := range []store.DictationProfile{
		{Name: "", Chunking: "word", Advance: "auto"},
		{Name: "x", Chunking: "pages", Advance: "auto"},
		{Name: "x", Chunking: "word", Advance: "vite"},
	} {
		if _, err := d.SaveProfile(bad); err == nil {
			t.Errorf("profil invalide accepté : %+v", bad)
		}
	}
	if err := d.DeleteProfile(id); err != nil {
		t.Fatal(err)
	}
}
