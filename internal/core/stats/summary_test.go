package stats

import (
	"math"
	"testing"

	"dactylo/internal/core/typing"
)

func keys(start int, s string) []typing.Keystroke {
	var ks []typing.Keystroke
	for i, r := range []rune(s) {
		ks = append(ks, typing.Keystroke{Seq: start + i, Position: start + i, Typed: string(r)})
	}
	return ks
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestSummarizePerfectMinute(t *testing.T) {
	text := []rune("aaaaaaaaaa") // 10 caractères = 2 mots
	s := Summarize(typing.Score(text, keys(0, "aaaaaaaaaa"), typing.MatchOptions{}), 60000)
	if s.Keystrokes != 10 || s.CorrectChars != 10 || s.UncorrectedErrors != 0 {
		t.Errorf("compteurs = %+v", s)
	}
	if !near(s.WPMGross, 2) || !near(s.WPMNet, 2) || !near(s.CPM, 10) || !near(s.Accuracy, 100) {
		t.Errorf("mesures = %+v", s)
	}
}

func TestSummarizeWithErrors(t *testing.T) {
	text := []rune("abcdefghij")
	// 10 frappes dont 2 fausses non corrigées, en 30 s.
	s := Summarize(typing.Score(text, keys(0, "abXdefYhij"), typing.MatchOptions{}), 30000)
	if !near(s.Accuracy, 80) {
		t.Errorf("précision = %v, want 80", s.Accuracy)
	}
	if !near(s.WPMGross, 4) { // 10/5 = 2 mots en 0,5 min
		t.Errorf("MPM brut = %v, want 4", s.WPMGross)
	}
	if !near(s.WPMNet, 0) { // 4 − 2/0,5 = 0
		t.Errorf("MPM net = %v, want 0", s.WPMNet)
	}
	if !near(s.CPM, 16) {
		t.Errorf("CPM = %v, want 16", s.CPM)
	}
}

func TestSummarizeCorrectedErrorKeepsAccuracyPenalty(t *testing.T) {
	text := []rune("ab")
	ks := []typing.Keystroke{
		{Seq: 0, Position: 0, Typed: "x"},
		{Seq: 1, Position: 0, Typed: typing.Backspace},
		{Seq: 2, Position: 0, Typed: "a"},
		{Seq: 3, Position: 1, Typed: "b"},
	}
	s := Summarize(typing.Score(text, ks, typing.MatchOptions{}), 60000)
	if s.Backspaces != 1 || s.Keystrokes != 3 || s.UncorrectedErrors != 0 {
		t.Errorf("compteurs = %+v", s)
	}
	if !near(s.Accuracy, 200.0/3) {
		t.Errorf("précision = %v", s.Accuracy)
	}
}

func TestSummarizeEmptyOrZeroDuration(t *testing.T) {
	s := Summarize(nil, 0)
	if s.Accuracy != 0 || s.WPMGross != 0 || s.CPM != 0 {
		t.Errorf("session vide = %+v", s)
	}
	s = Summarize(typing.Score([]rune("a"), keys(0, "a"), typing.MatchOptions{}), 0)
	if s.WPMGross != 0 || s.Accuracy != 100 {
		t.Errorf("durée nulle = %+v", s)
	}
}
