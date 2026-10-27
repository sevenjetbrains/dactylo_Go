package typing

import "testing"

// timed simule la frappe de s à partir de start, une frappe toutes les `step` ms.
func timed(start, seq int, t0, step int64, s string) []Keystroke {
	var ks []Keystroke
	for i, r := range []rune(s) {
		ks = append(ks, Keystroke{Seq: seq + i, Position: start + i, Typed: string(r), TimeMs: t0 + int64(i)*step})
	}
	return ks
}

func TestNormalizeWord(t *testing.T) {
	cases := map[string]string{"Maître,": "maître", "«Être»": "être", "l'été.": "été", "qu'il": "il", "Jusqu’à": "", "aujourd'hui": "aujourd'hui", "d'Artagnan": "artagnan", "1872": "", "à": "", "être": "être"}
	for in, want := range cases {
		if got := NormalizeWord(in); got != want {
			t.Errorf("NormalizeWord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWordAttemptsCorrectWrongAndSlow(t *testing.T) {
	txt := []rune("le chat dort, le chien aboie 12 fois")
	// tout à 100 ms par frappe, sauf « chien » tapé à 400 ms par frappe,
	// et « dort » tapé « dirt ».
	var ks []Keystroke
	seq, t0 := 0, int64(0)
	add := func(start int, s string, step int64) {
		ks = append(ks, timed(start, seq, t0+step, step, s)...)
		seq += len([]rune(s))
		t0 += step * int64(len([]rune(s)))
	}
	add(0, "le chat ", 100)
	add(8, "dirt, ", 100)
	add(14, "le ", 100)
	add(17, "chien", 400)
	add(22, " aboie 12 fois", 100)

	got := WordAttempts(txt, Score(txt, ks, MatchOptions{}), DefaultSlowThreshold)
	byWord := map[string]WordAttempt{}
	for _, a := range got {
		byWord[a.Word] = a
	}
	if _, ok := byWord["12"]; ok {
		t.Error("un nombre n'est pas un mot")
	}
	if a := byWord["dort"]; a.Correct || a.Typed != "dirt," || a.ErrorType != ErrSubstitution || a.Shown != "dort" {
		t.Errorf("dort = %+v", a)
	}
	if a := byWord["chien"]; !a.Correct || !a.Slow || a.DurationMs != 2000 || a.MsPerChar != 400 {
		t.Errorf("chien = %+v", a)
	}
	if a := byWord["chat"]; !a.Correct || a.Slow || a.MsPerChar != 100 {
		t.Errorf("chat = %+v", a)
	}
	if a := byWord["le"]; a.Position != 14 { // dernier « le » rencontré
		t.Errorf("le = %+v", a)
	}
	if len(got) != 7 { // le chat dort le chien aboie fois
		t.Errorf("%d mots, want 7 : %+v", len(got), got)
	}
}

func TestWordAttemptsPartialAndBackspace(t *testing.T) {
	txt := []rune("un deux trois")
	ks := timed(0, 0, 0, 100, "un dx")
	ks = append(ks, Keystroke{Seq: 5, Position: 4, Typed: Backspace, TimeMs: 600})
	ks = append(ks, timed(4, 6, 700, 100, "eux t")...)
	got := WordAttempts(txt, Score(txt, ks, MatchOptions{}), 0)
	if len(got) != 2 { // « trois » n'est pas atteint en entier
		t.Fatalf("mots = %+v", got)
	}
	if got[1].Word != "deux" || got[1].Correct || got[1].Typed != "dxux" {
		t.Errorf("deux = %+v", got[1])
	}
	if got[0].Slow || got[1].Slow {
		t.Error("moins de 3 mots : pas de repérage des lents")
	}
	if WordAttempts(txt, nil, 1.5) != nil {
		t.Error("sans frappe : aucun mot")
	}
}
