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

	list, err := s.ListTexts(ctx)
	if err != nil || len(list) != 1 || list[0].Length != 14 || list[0].Preview != "Maître corbeau" {
		t.Errorf("ListTexts = %+v, %v", list, err)
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
