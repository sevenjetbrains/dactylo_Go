package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"dactylo/internal/core/stats"
	"dactylo/internal/core/typing"
)

func openMem(t *testing.T) *Store {
	t.Helper()
	s, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestEnsureDefaultProfileIsIdempotent(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	id1, err := s.EnsureDefaultProfile(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.EnsureDefaultProfile(ctx)
	if err != nil || id1 != id2 {
		t.Fatalf("ids %d / %d, err %v", id1, id2, err)
	}
}

func TestTextsCRUD(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	id, err := s.CreateText(ctx, Text{Title: "Fable", Content: "Maître corbeau", Tags: []string{"poésie"}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetText(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Fable" || got.Content != "Maître corbeau" || got.Language != "fr" ||
		got.Source != "paste" || len(got.Tags) != 1 || got.Tags[0] != "poésie" {
		t.Errorf("GetText = %+v", got)
	}

	pid, _ := s.EnsureDefaultProfile(ctx)
	list, err := s.ListTexts(ctx, pid)
	if err != nil || len(list) == 0 || list[0].ID != id { // le plus récent d'abord
		t.Fatalf("ListTexts : %v", err)
	}
	if l := list[0]; l.Length != 14 || l.Preview != "Maître corbeau" || l.BestWPM != nil ||
		l.Position != 0 || l.Content != "Maître corbeau" {
		t.Errorf("résumé = %+v", l)
	}

	got.Title, got.Content, got.Tags = "Fable 2", "Maître renard", nil
	if err := s.UpdateText(ctx, got); err != nil {
		t.Fatal(err)
	}
	if u, _ := s.GetText(ctx, id); u.Title != "Fable 2" || u.Content != "Maître renard" || len(u.Tags) != 0 {
		t.Errorf("après modification = %+v", u)
	}
	if err := s.UpdateText(ctx, Text{ID: 999, Title: "x", Content: "y"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("modification d'un texte absent : %v", err)
	}

	pdfID, err := s.CreateText(ctx, Text{Title: "Livre", Content: "abc", Source: "pdf", SourceFile: "livre.pdf", Pages: "3-5"})
	if err != nil {
		t.Fatal(err)
	}
	if p, _ := s.GetText(ctx, pdfID); p.Source != "pdf" || p.SourceFile != "livre.pdf" || p.Pages != "3-5" {
		t.Errorf("texte PDF = %+v", p)
	}

	if err := s.DeleteText(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetText(ctx, id); !errors.Is(err, ErrNotFound) {
		t.Errorf("après suppression : %v", err)
	}
	if err := s.DeleteText(ctx, id); !errors.Is(err, ErrNotFound) {
		t.Errorf("double suppression : %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	pid, _ := s.EnsureDefaultProfile(ctx)
	tid, _ := s.CreateText(ctx, Text{Title: "t", Content: "ab"})
	now := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)

	sid, err := s.CreateSession(ctx, pid, tid, "copy", "", now)
	if err != nil {
		t.Fatal(err)
	}
	batch := []typing.Scored{
		{Keystroke: typing.Keystroke{Seq: 0, Position: 0, Typed: "a", TimeMs: 10}, Expected: "a", Correct: true},
		{Keystroke: typing.Keystroke{Seq: 1, Position: 1, Typed: "x", TimeMs: 20}, Expected: "b"},
	}
	if err := s.SaveKeystrokes(ctx, sid, batch); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveKeystrokes(ctx, sid, batch); err != nil { // renvoi du même lot
		t.Fatal(err)
	}
	ks, err := s.Keystrokes(ctx, sid)
	if err != nil || len(ks) != 2 || ks[1].Typed != "x" || ks[1].Position != 1 {
		t.Fatalf("Keystrokes = %+v, %v", ks, err)
	}

	sum := stats.Summary{WPMGross: 40, WPMNet: 30, CPM: 150, Accuracy: 90}
	errs := []typing.WordError{{Expected: "ab", Typed: "ax", Position: 0, Type: typing.ErrSubstitution}}
	if err := s.FinishSession(ctx, sid, now.Add(time.Minute), sum, errs); err != nil {
		t.Fatal(err)
	}
	if err := s.FinishSession(ctx, sid, now, sum, nil); err == nil {
		t.Error("une session terminée ne doit pas l'être deux fois")
	}
	se, err := s.GetSession(ctx, sid)
	if err != nil || !se.EndedAt.Valid || se.Mode != "copy" {
		t.Errorf("GetSession = %+v, %v", se, err)
	}
	var n int
	_ = s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM session_errors WHERE session_id = ?", sid).Scan(&n)
	if n != 1 {
		t.Errorf("session_errors = %d", n)
	}
	if _, err := s.GetSession(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("session absente : %v", err)
	}
}

func TestProgressAndBestScore(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	pid, _ := s.EnsureDefaultProfile(ctx)
	tid, _ := s.CreateText(ctx, Text{Title: "t", Content: "abcdef"})

	if pos, err := s.GetProgress(ctx, pid, tid); err != nil || pos != 0 {
		t.Fatalf("progression initiale = %d, %v", pos, err)
	}
	if err := s.SetProgress(ctx, pid, tid, 3); err != nil {
		t.Fatal(err)
	}
	if err := s.SetProgress(ctx, pid, tid, 5); err != nil { // mise à jour
		t.Fatal(err)
	}
	if pos, _ := s.GetProgress(ctx, pid, tid); pos != 5 {
		t.Errorf("progression = %d, want 5", pos)
	}

	now := time.Now()
	for _, wpm := range []float64{20, 35} {
		sid, _ := s.CreateSession(ctx, pid, tid, "copy", "", now)
		_ = s.FinishSession(ctx, sid, now, stats.Summary{WPMNet: wpm}, nil)
	}
	list, _ := s.ListTexts(ctx, pid)
	if list[0].BestWPM == nil || *list[0].BestWPM != 35 || list[0].Position != 5 {
		t.Errorf("résumé = %+v", list[0])
	}
}

func TestRecentAverages(t *testing.T) {
	s, ctx := openMem(t), context.Background()
	pid, _ := s.EnsureDefaultProfile(ctx)
	tid, _ := s.CreateText(ctx, Text{Title: "t", Content: "ab"})
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	var last int64
	for i := 0; i < 12; i++ {
		sid, _ := s.CreateSession(ctx, pid, tid, "copy", "", base)
		wpm := float64(i * 10) // 0, 10, ..., 110
		if err := s.FinishSession(ctx, sid, base.Add(time.Duration(i)*time.Hour), stats.Summary{WPMNet: wpm, Accuracy: 90}, nil); err != nil {
			t.Fatal(err)
		}
		last = sid
	}
	open, _ := s.CreateSession(ctx, pid, tid, "copy", "", base) // non terminée : ignorée

	a, err := s.RecentAverages(ctx, pid, "copy", last, 10)
	if err != nil {
		t.Fatal(err)
	}
	// sessions 1..10 hors la dernière (110) : 10..100 → moyenne 55
	if a.Sessions != 10 || a.WPMNet != 55 || a.Accuracy != 90 {
		t.Errorf("RecentAverages = %+v", a)
	}

	empty, err := s.RecentAverages(ctx, pid, "dictation", open, 10)
	if err != nil || empty.Sessions != 0 || empty.WPMNet != 0 {
		t.Errorf("sans historique = %+v, %v", empty, err)
	}
}
