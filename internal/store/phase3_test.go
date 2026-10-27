package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"dactylo/internal/core/srs"
	"dactylo/internal/core/stats"
	"dactylo/internal/core/typing"
)

func TestWordsRepository(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	pid, _ := s.EnsureDefaultProfile(ctx)

	if _, ok, err := s.GetWord(ctx, pid, "être"); ok || err != nil {
		t.Fatalf("mot absent : ok=%v err=%v", ok, err)
	}
	avg := 120.0
	w := Word{Word: "être", Shown: "Être", Encounters: 1, Errors: 1, AvgMsPerChr: &avg,
		Mastery: 0, NextReview: "2026-10-28", LastErrorAt: "2026-10-27T10:00:00.000Z",
		Card: srs.Card{Interval: 1, Ease: 2.3, Due: "2026-10-28"}}
	id, err := s.SaveWord(ctx, pid, w)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	for _, a := range []typing.WordAttempt{
		{Typed: "etre", ErrorType: typing.ErrAccent},
		{Typed: "etre", ErrorType: typing.ErrAccent},
		{Typed: "ètre", ErrorType: typing.ErrAccent},
		{Typed: "êtr", ErrorType: typing.ErrMissingLetter},
		{Typed: "être", Correct: true, Slow: true},
	} {
		if err := s.AddAttempt(ctx, id, 0, a, now); err != nil {
			t.Fatal(err)
		}
	}

	got, ok, err := s.GetWord(ctx, pid, "être")
	if !ok || err != nil || got.Shown != "Être" || got.Card.Ease != 2.3 || *got.AvgMsPerChr != 120 {
		t.Fatalf("GetWord = %+v, %v, %v", got, ok, err)
	}
	got.Errors, got.Mastery, got.NextReview = 3, 1, "2026-12-01"
	if _, err := s.SaveWord(ctx, pid, got); err != nil {
		t.Fatal(err)
	}

	_, _ = s.SaveWord(ctx, pid, Word{Word: "chat", Shown: "chat", NextReview: "2026-10-27", Card: srs.Card{Due: "2026-10-27"}})
	_, _ = s.SaveWord(ctx, pid, Word{Word: "chien", Shown: "chien", NextReview: "2026-11-30", Card: srs.Card{Due: "2026-11-30"}})

	list := func(status string) []string {
		ws, err := s.ListWords(ctx, pid, WordFilter{Status: status, Today: "2026-10-27"})
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, w := range ws {
			out = append(out, w.Word)
		}
		return out
	}
	if got := list("due"); len(got) != 1 || got[0] != "chat" {
		t.Errorf("à réviser = %v", got)
	}
	if got := list("learning"); len(got) != 2 {
		t.Errorf("en cours = %v", got)
	}
	if got := list("mastered"); len(got) != 1 || got[0] != "être" {
		t.Errorf("maîtrisés = %v", got)
	}
	all, _ := s.ListWords(ctx, pid, WordFilter{Status: "all"})
	if len(all) != 3 || all[0].Word != "être" { // le plus fautif d'abord
		t.Fatalf("tous = %+v", all)
	}
	if e := all[0]; e.DominantError != typing.ErrAccent || len(e.Forms) != 3 || e.Forms[0] != "etre" {
		t.Errorf("détail = forms %v, dominant %q", e.Forms, e.DominantError)
	}
	if n, _ := s.DueCount(ctx, pid, "2026-10-27"); n != 1 {
		t.Errorf("DueCount = %d", n)
	}

	if byID, err := s.GetWordByID(ctx, pid, id); err != nil || byID.Word != "être" {
		t.Errorf("GetWordByID = %+v, %v", byID, err)
	}
	if err := s.DeleteWord(ctx, pid, id); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetWordByID(ctx, pid, id); !errors.Is(err, ErrNotFound) {
		t.Errorf("après suppression : %v", err)
	}
	if err := s.DeleteWord(ctx, pid, id); !errors.Is(err, ErrNotFound) {
		t.Errorf("double suppression : %v", err)
	}
}

func TestStatsRepository(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	pid, _ := s.EnsureDefaultProfile(ctx)
	tid, _ := s.CreateText(ctx, Text{Title: "Fable", Content: "ab"})
	base := time.Date(2026, 10, 25, 9, 0, 0, 0, time.UTC)

	for i, wpm := range []float64{20, 35, 30} {
		textID := tid
		if i == 2 {
			textID = 0 // révision : pas de texte
		}
		sid, err := s.CreateSession(ctx, pid, textID, "copy", "", base)
		if err != nil {
			t.Fatal(err)
		}
		_ = s.SaveKeystrokes(ctx, sid, []typing.Scored{
			{Keystroke: typing.Keystroke{Seq: 0, Position: 0, Typed: "a", TimeMs: 100}, Expected: "a", Correct: true},
			{Keystroke: typing.Keystroke{Seq: 1, Position: 1, Typed: "x", TimeMs: 250}, Expected: "b"},
		})
		errs := []typing.WordError{{Expected: "ab", Typed: "ax", Type: typing.ErrSubstitution}}
		if err := s.FinishSession(ctx, sid, base.AddDate(0, 0, i), stats.Summary{WPMNet: wpm, Accuracy: 50, DurationMs: 60000}, errs); err != nil {
			t.Fatal(err)
		}
	}
	_, _ = s.CreateSession(ctx, pid, tid, "copy", "", base) // non terminée : ignorée

	pts, err := s.SessionPoints(ctx, pid, 2)
	if err != nil || len(pts) != 2 || pts[0].WPMNet != 35 || pts[1].WPMNet != 30 {
		t.Fatalf("SessionPoints = %+v, %v", pts, err)
	}
	if pts[0].TextTitle != "Fable" || pts[1].TextTitle != "" || pts[0].DurationMs != 60000 || pts[0].Errors != 1 {
		t.Errorf("points = %+v", pts)
	}
	if all, _ := s.SessionPoints(ctx, pid, 0); len(all) != 3 {
		t.Errorf("toutes = %d", len(all))
	}

	tot, err := s.SessionTotals(ctx, pid)
	if err != nil || tot.Sessions != 3 || tot.TotalMs != 180000 || tot.BestWPM != 35 {
		t.Errorf("totaux = %+v, %v", tot, err)
	}
	days, _ := s.SessionDays(ctx, pid)
	if len(days) != 3 || days[0].IsZero() {
		t.Errorf("jours = %v", days)
	}
	ks, err := s.RecentKeystrokes(ctx, pid, 2)
	if err != nil || len(ks) != 2 || len(ks[0]) != 2 || ks[0][1].Correct || !ks[0][0].Correct || ks[0][1].Expected != "b" {
		t.Errorf("frappes = %+v, %v", ks, err)
	}
}
