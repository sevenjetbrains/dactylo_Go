package typing

import (
	"reflect"
	"testing"
)

// typeAll simule la frappe d'une chaîne, une rune par frappe, à partir de pos.
func typeAll(start, seq int, s string) []Keystroke {
	var ks []Keystroke
	for i, r := range []rune(s) {
		ks = append(ks, Keystroke{Seq: seq + i, Position: start + i, Typed: string(r), TimeMs: int64(100 * (seq + i))})
	}
	return ks
}

func TestMatch(t *testing.T) {
	strict := MatchOptions{}
	fold := MatchOptions{FoldTypography: true}
	cases := []struct {
		exp, typed rune
		opts       MatchOptions
		want       bool
	}{
		{'ê', 'ê', strict, true},
		{'ê', 'e', strict, false},
		{'\n', '\n', strict, true},
		{'\n', ' ', strict, true},
		{' ', ' ', strict, true},
		{'’', '\'', strict, false},
		{'’', '\'', fold, true},
		{'«', '"', fold, true},
		{'a', 'A', fold, false},
	}
	for _, c := range cases {
		if got := Match(c.exp, c.typed, c.opts); got != c.want {
			t.Errorf("Match(%q, %q, %+v) = %v, want %v", c.exp, c.typed, c.opts, got, c.want)
		}
	}
}

func TestScoreRecomputesCorrectness(t *testing.T) {
	text := []rune("être")
	ks := []Keystroke{
		{Seq: 2, Position: 1, Typed: "t"},
		{Seq: 1, Position: 0, Typed: "ê"}, // touche morte : une seule frappe
		{Seq: 3, Position: 2, Typed: "e"},
		{Seq: 4, Position: 9, Typed: "x"}, // hors texte
		{Seq: 5, Position: 2, Typed: Backspace},
	}
	got := Score(text, ks, MatchOptions{})
	if got[0].Seq != 1 || !got[0].Correct || got[0].Expected != "ê" {
		t.Errorf("première frappe = %+v", got[0])
	}
	if got[2].Correct {
		t.Errorf("'e' à la place de 'r' compté juste")
	}
	if got[3].Correct || got[3].Expected != "" {
		t.Errorf("frappe hors texte = %+v", got[3])
	}
	if !got[4].IsBackspace() || got[4].Correct {
		t.Errorf("retour arrière = %+v", got[4])
	}
}

func TestReplayFreeModeWithCorrection(t *testing.T) {
	text := []rune("chat")
	ks := typeAll(0, 0, "chot") // erreur en position 2
	ks = append(ks, Keystroke{Seq: 4, Position: 3, Typed: Backspace},
		Keystroke{Seq: 5, Position: 2, Typed: Backspace})
	ks = append(ks, typeAll(2, 6, "at")...)
	f := Replay(Score(text, ks, MatchOptions{}))
	if f.UncorrectedErrors() != 0 || f.CorrectChars() != 4 || f.Reached != 4 {
		t.Errorf("final = %+v", f)
	}
}

func TestReplayBlockingMode(t *testing.T) {
	text := []rune("ab")
	ks := []Keystroke{
		{Seq: 0, Position: 0, Typed: "x"}, // refusée : on reste en 0
		{Seq: 1, Position: 0, Typed: "a"},
		{Seq: 2, Position: 1, Typed: "b"},
	}
	f := Replay(Score(text, ks, MatchOptions{}))
	if f.UncorrectedErrors() != 0 || f.CorrectChars() != 2 {
		t.Errorf("final = %+v", f)
	}
}

func TestReplayUncorrected(t *testing.T) {
	f := Replay(Score([]rune("chat"), typeAll(0, 0, "chot"), MatchOptions{}))
	if f.UncorrectedErrors() != 1 || f.CorrectChars() != 3 {
		t.Errorf("final = %+v", f)
	}
}

func TestWords(t *testing.T) {
	got := Words([]rune("  Le chat,\nnoir. "))
	want := []Span{{2, 4}, {5, 10}, {11, 16}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Words = %v, want %v", got, want)
	}
	if Words([]rune("   ")) != nil {
		t.Error("texte blanc : aucun mot attendu")
	}
}

func TestMistypedWords(t *testing.T) {
	text := []rune("Le chat être là")
	// "Le" juste, "chat" tapé "chot" puis corrigé, "être" tapé "etre",
	// "là" pas encore atteint en entier.
	ks := typeAll(0, 0, "Le cho")
	ks = append(ks, Keystroke{Seq: 6, Position: 5, Typed: Backspace})
	ks = append(ks, typeAll(5, 7, "at etre l")...)
	got := MistypedWords(text, Score(text, ks, MatchOptions{}))
	want := []WordError{
		{Expected: "chat", Typed: "chot", Position: 3, Type: ErrSubstitution},
		{Expected: "être", Typed: "etre", Position: 8, Type: ErrAccent},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MistypedWords =\n%+v\nwant\n%+v", got, want)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ exp, typed, want string }{
		{"Paris", "paris", ErrCase},
		{"élève", "eleve", ErrAccent},
		{"Élève", "eleve", ErrAccent},
		{"chat,", "chat.", ErrPunctuation},
		{"chat,", "chat", ErrPunctuation},
		{"chat", "cht", ErrMissingLetter},
		{"chat", "chaat", ErrExtraLetter},
		{"chat", "caht", ErrTransposition},
		{"chat", "chien", ErrSubstitution},
		{"chat", "chot", ErrSubstitution},
		{"chat", "", ErrOmittedWord},
	}
	for _, c := range cases {
		if got := Classify(c.exp, c.typed); got != c.want {
			t.Errorf("Classify(%q, %q) = %q, want %q", c.exp, c.typed, got, c.want)
		}
	}
}

func TestCoreWord(t *testing.T) {
	cases := map[string]string{"l'été,": "l'été", "«bonjour»": "bonjour", "...": "", "chat": "chat"}
	for in, want := range cases {
		if got := CoreWord(in); got != want {
			t.Errorf("CoreWord(%q) = %q, want %q", in, got, want)
		}
	}
}
