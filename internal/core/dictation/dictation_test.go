package dictation

import (
	"reflect"
	"strings"
	"testing"

	"dactylo/internal/core/typing"
)

func texts(cs []Chunk) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.Text)
	}
	return out
}

func TestChunks(t *testing.T) {
	src := "Le petit chat dort. Il rêve d'une souris grise, très grise.\n\nLe chien aboie."
	cases := []struct {
		mode string
		size int
		want []string
	}{
		{ChunkWord, 0, strings.Fields(src)},
		{ChunkGroup, 4, []string{"Le petit chat dort.", "Il rêve d'une souris", "grise, très grise.", "Le chien aboie."}},
		{ChunkGroup, 0, []string{"Le petit chat dort.", "Il rêve d'une souris", "grise, très grise.", "Le chien aboie."}},
		{ChunkSentence, 0, []string{"Le petit chat dort.", "Il rêve d'une souris grise, très grise.", "Le chien aboie."}},
		{ChunkParagraph, 0, []string{"Le petit chat dort. Il rêve d'une souris grise, très grise.", "Le chien aboie."}},
	}
	for _, c := range cases {
		got := texts(Chunks(src, c.mode, c.size, false))
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s/%d =\n%q\nwant\n%q", c.mode, c.size, got, c.want)
		}
	}
	cs := Chunks(src, ChunkGroup, 4, false)
	if cs[1].Words != 4 || cs[1].Index != 1 || cs[1].Spoken != cs[1].Text {
		t.Errorf("segment = %+v", cs[1])
	}
	if Chunks("  ", ChunkGroup, 4, false) != nil && len(Chunks("  ", ChunkGroup, 4, false)) != 0 {
		t.Error("texte vide")
	}
}

func TestAbbreviations(t *testing.T) {
	got := Chunks("En 1815, M. Myriel et Mme Magloire. J. Valjean arrive.", ChunkSentence, 0, true)
	if len(got) != 2 || got[0].Text != "En 1815, M. Myriel et Mme Magloire." {
		t.Fatalf("phrases = %q", texts(got))
	}
	if got[0].Spoken != "En 1815 virgule M. Myriel et Mme Magloire point" {
		t.Errorf("lu = %q", got[0].Spoken)
	}
	for w, want := range map[string]bool{"M.": true, "Dr.": true, "cf.": true, "J.": true, "fin.": false, "Paris.": false} {
		if isAbbreviation(w) != want {
			t.Errorf("isAbbreviation(%q) = %v", w, !want)
		}
	}
}

func TestSpokenPunctuation(t *testing.T) {
	src := "« Que faisiez-vous au temps chaud ?\nDit-elle, à cette emprunteuse... »\n\nFin (vraiment) ; non !"
	got := Chunks(src, ChunkParagraph, 0, true)
	want := []string{
		"ouvrez les guillemets Que faisiez-vous au temps chaud point d'interrogation à la ligne " +
			"Dit-elle virgule à cette emprunteuse points de suspension fermez les guillemets à la ligne",
		"Fin ouvrez la parenthèse vraiment fermez la parenthèse point-virgule non point d'exclamation",
	}
	if len(got) != 2 || got[0].Spoken != want[0] || got[1].Spoken != want[1] {
		t.Errorf("spoken =\n%q\n%q\nwant\n%q\n%q", got[0].Spoken, got[1].Spoken, want[0], want[1])
	}
	if CountWords(" un  deux\ntrois ") != 3 {
		t.Error("CountWords")
	}
}

// Critère d'acceptation transverse (section 7.3) : une dictée de 50 mots
// avec 5 fautes connues produit exactement ces 5 fautes, bien classées.
func TestFiftyWordsFiveKnownFaults(t *testing.T) {
	expected := "En 1815, M. Charles-François-Bienvenu Myriel était évêque de Digne. " +
		"C'était un vieillard d'environ soixante-quinze ans ; il occupait le siège de Digne depuis 1806. " +
		"Quoique ce détail ne touche en aucune manière au fond même de ce que nous avons à raconter, " +
		"il n'est peut-être pas inutile d'indiquer les bruits."
	if n := len(strings.Fields(expected)); n != 50 {
		t.Fatalf("texte de %d mots, 50 attendus", n)
	}
	typed := strings.NewReplacer(
		"était évêque", "était évèque", // accent
		"de Digne.", "de digne.", // majuscule
		"vieillard", "viellard", // lettre manquante
		"d'environ", "d'environs", // lettre en trop
		"soixante-quinze", "soixante quinze", // ponctuation (mot coupé en deux)
	).Replace(expected)

	r := Correct(expected, typed)
	want := map[string]int{
		typing.ErrAccent: 1, typing.ErrCase: 1, typing.ErrMissingLetter: 1,
		typing.ErrExtraLetter: 1, typing.ErrPunctuation: 1,
	}
	if r.Words != 50 || r.Faults != 5 || !reflect.DeepEqual(r.ByType, want) {
		t.Fatalf("rapport = %d mots, %d fautes, %v", r.Words, r.Faults, r.ByType)
	}
	if r.Accuracy != 90 {
		t.Errorf("précision = %v", r.Accuracy)
	}
	var faults []string
	for _, o := range r.Ops {
		if !o.Correct {
			faults = append(faults, o.Expected+"→"+o.Typed)
		}
	}
	wantFaults := []string{"évêque→évèque", "Digne.→digne.", "vieillard→viellard", "d'environ→d'environs", "soixante-quinze→soixante quinze"}
	if !reflect.DeepEqual(faults, wantFaults) {
		t.Errorf("fautes = %q", faults)
	}
}

func TestOmittedExtraAndInversion(t *testing.T) {
	r := Correct("le petit chat noir dort", "le chat nior dort vite")
	var got []string
	for _, o := range r.Ops {
		got = append(got, o.Expected+"|"+o.Typed+"|"+o.Type)
	}
	want := []string{
		"le|le|", "petit||omitted_word", "chat|chat|", "noir|nior|transposition", "dort|dort|", "|vite|extra_word",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ops = %q", got)
	}
	if r.Faults != 3 || r.Words != 5 || r.Accuracy != 40 {
		t.Errorf("rapport = %+v", r)
	}

	// mot sans rapport : substitution plutôt qu'oubli + ajout
	r = Correct("le chat dort", "le zèbre dort")
	if len(r.Ops) != 3 || r.Ops[1].Type != typing.ErrSubstitution {
		t.Errorf("sans rapport = %+v", r.Ops)
	}
	// rien tapé
	r = Correct("un deux", "")
	if r.Faults != 2 || r.Accuracy != 0 || r.ByType[typing.ErrOmittedWord] != 2 {
		t.Errorf("rien tapé = %+v", r)
	}
	if r := Correct("", ""); r.Words != 0 || r.Accuracy != 0 {
		t.Errorf("vide = %+v", r)
	}
}

func TestAttempts(t *testing.T) {
	ops := Align("Le maître, l'élève 12", "Le maitre, l'élève 12 bis")
	got := Attempts(ops)
	if len(got) != 3 { // « 12 » n'est pas un mot, « bis » est en trop
		t.Fatalf("tentatives = %+v", got)
	}
	if got[1].Word != "maître" || got[1].Correct || got[1].ErrorType != typing.ErrAccent || got[1].Shown != "maître" {
		t.Errorf("maître = %+v", got[1])
	}
	if got[2].Word != "élève" || !got[2].Correct || got[2].Shown != "élève" {
		t.Errorf("élève = %+v", got[2])
	}
}

func TestLevenshtein(t *testing.T) {
	for _, c := range []struct {
		a, b string
		d    int
	}{{"", "", 0}, {"chat", "chat", 0}, {"chat", "chot", 1}, {"chat", "", 4}, {"kitten", "sitting", 3}} {
		if got := levenshtein(c.a, c.b); got != c.d {
			t.Errorf("levenshtein(%q, %q) = %d", c.a, c.b, got)
		}
	}
}
