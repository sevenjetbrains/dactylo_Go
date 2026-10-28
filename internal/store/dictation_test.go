package store

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"dactylo/internal/core/stats"
	"dactylo/internal/core/typing"
)

func TestDictationProfiles(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	list, err := s.ListDictationProfiles(ctx)
	if err != nil || len(list) != 2 || list[0].Name != "Débutant lent" || list[1].Advance != "auto" || list[1].MaxRepeats != 2 {
		t.Fatalf("profils intégrés = %+v, %v", list, err)
	}
	if !list[0].SpeakPunctuation || list[0].GroupSize != 4 || list[0].Rate != 0.9 {
		t.Errorf("Débutant lent = %+v", list[0])
	}

	id, err := s.SaveDictationProfile(ctx, DictationProfile{Name: "Rapide", Voice: "Denise", Rate: 1.4, Pitch: 1, Volume: 1,
		Chunking: "sentence", GroupSize: 4, PauseMs: 500, Advance: "auto"})
	if err != nil {
		t.Fatal(err)
	}
	p := DictationProfile{ID: id, Name: "Rapide 2", Rate: 1.5, Pitch: 1, Volume: 0.8, Chunking: "word", GroupSize: 1, Advance: "adaptive", MaxRepeats: 3}
	if _, err := s.SaveDictationProfile(ctx, p); err != nil {
		t.Fatal(err)
	}
	list, _ = s.ListDictationProfiles(ctx)
	if len(list) != 3 || list[2].Name != "Rapide 2" || list[2].Voice != "" || list[2].MaxRepeats != 3 {
		t.Errorf("après modification = %+v", list)
	}
	if _, err := s.SaveDictationProfile(ctx, DictationProfile{ID: 999, Name: "x", Chunking: "word", Advance: "auto"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("profil absent : %v", err)
	}
	if _, err := s.SaveDictationProfile(ctx, DictationProfile{Name: "x", Chunking: "pages", Advance: "auto"}); err == nil {
		t.Error("découpage invalide accepté")
	}
	if err := s.DeleteDictationProfile(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteDictationProfile(ctx, id); !errors.Is(err, ErrNotFound) {
		t.Errorf("double suppression : %v", err)
	}
}

func TestDictationSessionsAndAverages(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	pid, _ := s.EnsureDefaultProfile(ctx)
	tid, _ := s.CreateText(ctx, Text{Title: "t", Content: "un deux"})
	now := time.Now()

	var last int64
	for i, faults := range []int{2, 4} {
		sid, err := s.CreateSession(ctx, pid, tid, "dictation", "{}", now)
		if err != nil {
			t.Fatal(err)
		}
		errs := make([]typing.WordError, faults)
		for k := range errs {
			errs[k] = typing.WordError{Expected: "un", Typed: "um", Type: typing.ErrSubstitution}
		}
		if err := s.FinishSession(ctx, sid, now.Add(time.Duration(i)*time.Minute), stats.Summary{WPMNet: float64(20 + 10*i), Accuracy: float64(90 - i*10)}, errs); err != nil {
			t.Fatal(err)
		}
		last = sid
	}
	if err := s.SetSessionSettings(ctx, last, `{"typed":"um deux"}`); err != nil {
		t.Fatal(err)
	}
	if se, _ := s.GetSession(ctx, last); !strings.Contains(se.Settings, "um deux") {
		t.Errorf("réglages = %s", se.Settings)
	}
	a, err := s.RecentDictationAverages(ctx, pid, 0, 10)
	if err != nil || a.Sessions != 2 || a.Faults != 3 || a.WPM != 25 || a.Accuracy != 85 {
		t.Errorf("moyennes = %+v, %v", a, err)
	}
	if a, _ := s.RecentDictationAverages(ctx, pid, last, 10); a.Sessions != 1 || a.Faults != 2 {
		t.Errorf("hors session = %+v", a)
	}
}
