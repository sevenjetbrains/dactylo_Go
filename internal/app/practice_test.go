package app

import (
	"context"
	"strings"
	"testing"

	"dactylo/internal/core/events"
	"dactylo/internal/core/typing"
	"dactylo/internal/store"
)

type fixture struct {
	store    *store.Store
	bus      *events.Bus
	texts    *TextService
	practice *PracticeService
	picker   *fakePicker
	events   []events.Event
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	s, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	pid, err := s.EnsureDefaultProfile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{store: s, bus: events.NewBus()}
	f.bus.SubscribeAll(func(e events.Event) { f.events = append(f.events, e) })
	f.picker = &fakePicker{}
	f.texts = NewTextService(s, f.bus, pid, f.picker)
	f.practice = NewPracticeService(s, f.bus, pid)
	return f
}

func typeString(start, seq int, s string) []typing.Keystroke {
	var ks []typing.Keystroke
	for i, r := range []rune(s) {
		ks = append(ks, typing.Keystroke{Seq: seq + i, Position: start + i, Typed: string(r), TimeMs: int64(seq+i) * 200})
	}
	return ks
}

func TestCreateTextCleansAndPublishes(t *testing.T) {
	f := newFixture(t)
	id, err := f.texts.CreateText(TextInput{Content: "  Le petit chat dort sur le canapé rouge.\r\n", Tags: []string{" chats ", ""}})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := f.texts.GetText(id)
	if got.Content != "Le petit chat dort sur le canapé rouge." {
		t.Errorf("contenu = %q", got.Content)
	}
	if got.Title != "Le petit chat dort sur le…" {
		t.Errorf("titre par défaut = %q", got.Title)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "chats" {
		t.Errorf("étiquettes = %v", got.Tags)
	}
	if len(f.events) != 1 || f.events[0].Name != events.TextImported {
		t.Errorf("événements = %+v", f.events)
	}

	if _, err := f.texts.CreateText(TextInput{Content: " \n "}); err == nil {
		t.Error("texte vide accepté")
	}
	if _, err := f.texts.CreateText(TextInput{Content: strings.Repeat("a", MaxTextLength+1)}); err == nil {
		t.Error("texte trop long accepté")
	}
	if err := f.texts.DeleteText(id); err != nil {
		t.Error(err)
	}
}

func TestPracticeFullSession(t *testing.T) {
	f := newFixture(t)
	tid, _ := f.texts.CreateText(TextInput{Title: "t", Content: "Le chat être"})
	f.events = nil

	sid, err := f.practice.StartSession(StartRequest{TextID: tid, Settings: `{"foldTypography":true,"correction":"free"}`})
	if err != nil {
		t.Fatal(err)
	}
	// "Le chat " juste, puis "etre" (accent oublié), envoyé en deux lots.
	if err := f.practice.RecordKeystrokes(sid, typeString(0, 0, "Le chat ")); err != nil {
		t.Fatal(err)
	}
	if err := f.practice.RecordKeystrokes(sid, typeString(8, 8, "etre")); err != nil {
		t.Fatal(err)
	}

	res, err := f.practice.FinishSession(sid, 60000)
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.Keystrokes != 12 || res.Summary.UncorrectedErrors != 1 {
		t.Errorf("summary = %+v", res.Summary)
	}
	if len(res.Mistakes) != 1 || res.Mistakes[0].Expected != "être" || res.Mistakes[0].Type != typing.ErrAccent {
		t.Errorf("mistakes = %+v", res.Mistakes)
	}
	if res.Average.Sessions != 0 {
		t.Errorf("première session : moyenne = %+v", res.Average)
	}
	if len(f.events) != 2 || f.events[0].Name != events.SessionCompleted || f.events[1].Name != events.WordMistyped {
		t.Errorf("événements = %+v", f.events)
	}

	if err := f.practice.RecordKeystrokes(sid, typeString(0, 99, "x")); err == nil {
		t.Error("frappes acceptées après la fin")
	}
	if _, err := f.practice.FinishSession(sid, 1); err == nil {
		t.Error("session terminée deux fois")
	}

	// Seconde session : la moyenne porte sur la première.
	sid2, _ := f.practice.StartSession(StartRequest{TextID: tid, Settings: ""})
	_ = f.practice.RecordKeystrokes(sid2, typeString(0, 0, "Le chat être"))
	res2, err := f.practice.FinishSession(sid2, 60000)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Average.Sessions != 1 || res2.Average.WPMNet != res.Summary.WPMNet {
		t.Errorf("moyenne = %+v", res2.Average)
	}
	if len(res2.Mistakes) != 0 {
		t.Errorf("session parfaite : %+v", res2.Mistakes)
	}
}

func TestPracticeErrors(t *testing.T) {
	f := newFixture(t)
	if _, err := f.practice.StartSession(StartRequest{TextID: 42}); err == nil {
		t.Error("texte inexistant accepté")
	}
	tid, _ := f.texts.CreateText(TextInput{Content: "ab"})
	if _, err := f.practice.StartSession(StartRequest{TextID: tid, Settings: "{pas du json"}); err == nil {
		t.Error("réglages invalides acceptés")
	}
	sid, _ := f.practice.StartSession(StartRequest{TextID: tid, Settings: ""})
	if _, err := f.practice.FinishSession(sid, -1); err == nil {
		t.Error("durée négative acceptée")
	}
	if _, err := f.practice.FinishSession(999, 1); err == nil {
		t.Error("session inconnue acceptée")
	}
}
