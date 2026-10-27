package exercise

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func rng() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func TestShuffled(t *testing.T) {
	var g ExerciseGenerator = Generator{}
	ex := g.Generate(Request{Kind: KindShuffled, Words: []string{"maître", "être", "Être", " ", "œil"}}, rng())
	words := strings.Fields(ex.Content)
	if ex.Kind != KindShuffled || len(words) < MinShuffledWords {
		t.Fatalf("exercice = %+v", ex)
	}
	count := map[string]int{}
	for _, w := range words {
		count[w]++
	}
	if len(count) != 3 || count["Être"] != 0 { // doublon « Être » retiré
		t.Errorf("mots = %v", count)
	}
	if empty := g.Generate(Request{Kind: KindShuffled}, rng()); empty.Content != "" {
		t.Errorf("sans mot = %+v", empty)
	}
}

func TestRepeated(t *testing.T) {
	ex := Generator{}.Generate(Request{Kind: KindRepeated, Words: []string{"chat", "chien"}, Repeat: 3}, rng())
	lines := strings.Split(ex.Content, "\n")
	if len(lines) != 2 {
		t.Fatalf("lignes = %q", lines)
	}
	for _, l := range lines {
		f := strings.Fields(l)
		if len(f) != 3 || f[0] != f[1] || f[1] != f[2] {
			t.Errorf("ligne = %q", l)
		}
	}
	if ex := (Generator{}).Generate(Request{Kind: KindRepeated, Words: []string{"a"}, Repeat: 0}, rng()); ex.Content != "a" {
		t.Errorf("répétition minimale = %q", ex.Content)
	}
}

func TestSentences(t *testing.T) {
	lib := []string{
		"Maître Corbeau, sur un arbre perché,\nTenait en son bec un fromage. Le Renard s'en saisit !",
		"Un maitre-nageur surveille. Ceci est court.",
		"Il était une fois un être étrange qui vivait là.",
	}
	ex := Generator{}.Generate(Request{Kind: KindSentences, Words: []string{"maître", "fromage", "être"}, Library: lib}, rng())
	if ex.Kind != KindSentences {
		t.Fatalf("exercice = %+v", ex)
	}
	got := strings.Split(ex.Content, "\n")
	if len(got) != 2 { // la 1re phrase contient « maître » et « fromage » : utilisée une fois
		t.Errorf("phrases = %q", got)
	}
	joined := ex.Content
	if !strings.Contains(joined, "Maître Corbeau, sur un arbre perché, Tenait en son bec un fromage.") ||
		!strings.Contains(joined, "un être étrange") || strings.Contains(joined, "nageur") {
		t.Errorf("phrases = %q", got)
	}

	fallback := Generator{}.Generate(Request{Kind: KindSentences, Words: []string{"zèbre"}, Library: lib}, rng())
	if fallback.Kind != KindShuffled {
		t.Errorf("sans phrase : liste mélangée attendue, %+v", fallback)
	}
}

func TestContainsWord(t *testing.T) {
	cases := []struct {
		s, w string
		want bool
	}{
		{"Le Maître dort.", "maitre", true},
		{"Le maître-nageur.", "maître", false},
		{"Un maître-nageur.", "maître-nageur", true},
		{"Pour l'être aimé.", "être", true},
		{"Pour l'élève.", "élève", true},
		{"Les maîtresses.", "maître", false},
		{"maître", "maître", true},
	}
	for _, c := range cases {
		if got := containsWord(c.s, c.w); got != c.want {
			t.Errorf("containsWord(%q, %q) = %v", c.s, c.w, got)
		}
	}
}
