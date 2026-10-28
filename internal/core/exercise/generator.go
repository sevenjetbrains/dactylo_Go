// Package exercise génère des exercices de révision ciblés sur les mots
// difficiles (F-44).
package exercise

import (
	"math/rand/v2"
	"regexp"
	"strings"
	"unicode"

	"dactylo/internal/core/text"
)

// Types d'exercice.
const (
	KindShuffled  = "shuffled"  // liste de mots mélangés
	KindRepeated  = "repeated"  // chaque mot répété N fois
	KindSentences = "sentences" // phrases de la bibliothèque contenant les mots
)

// MinShuffledWords : une liste trop courte est répétée jusqu'à cette taille.
const MinShuffledWords = 15

// MaxSentences borne la longueur d'un exercice de phrases.
const MaxSentences = 12

// Request décrit l'exercice voulu.
type Request struct {
	Kind    string
	Words   []string // formes à taper (ex. « Maître »)
	Repeat  int      // KindRepeated : répétitions par mot
	Library []string // KindSentences : contenus des textes
}

// Exercise est un texte prêt à taper.
type Exercise struct {
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

// ExerciseGenerator produit un exercice ; l'aléa est injecté pour les tests.
type ExerciseGenerator interface {
	Generate(req Request, rng *rand.Rand) Exercise
}

// Generator est l'implémentation par défaut.
type Generator struct{}

// Generate produit l'exercice demandé. Un exercice de phrases sans phrase
// trouvée se rabat sur la liste mélangée.
func (Generator) Generate(req Request, rng *rand.Rand) Exercise {
	words := dedupe(req.Words)
	if len(words) == 0 {
		return Exercise{Kind: req.Kind}
	}
	switch req.Kind {
	case KindRepeated:
		n := min(max(req.Repeat, 1), 20)
		lines := make([]string, len(words))
		for i, w := range shuffle(words, rng) {
			lines[i] = strings.TrimSpace(strings.Repeat(w+" ", n))
		}
		return Exercise{Kind: KindRepeated, Content: strings.Join(lines, "\n")}
	case KindSentences:
		if s := sentencesWith(words, req.Library, rng); len(s) > 0 {
			return Exercise{Kind: KindSentences, Content: strings.Join(s, "\n")}
		}
	}
	list := shuffle(words, rng)
	for len(list) < MinShuffledWords {
		list = append(list, shuffle(words, rng)...)
	}
	return Exercise{Kind: KindShuffled, Content: strings.Join(list, " ")}
}

// dedupe supprime les doublons (sans accents ni casse) et les mots vides.
func dedupe(words []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, w := range words {
		w = strings.TrimSpace(text.NFC(w))
		if k := text.SearchKey(w); w != "" && !seen[k] {
			seen[k] = true
			out = append(out, w)
		}
	}
	return out
}

func shuffle(words []string, rng *rand.Rand) []string {
	out := append([]string(nil), words...)
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

var sentenceRe = regexp.MustCompile(`[^.!?…]+[.!?…]+["»]?`)

// sentencesWith choisit, pour chaque mot, une phrase de la bibliothèque qui
// le contient (mot entier, sans tenir compte des accents ni de la casse).
func sentencesWith(words, library []string, rng *rand.Rand) []string {
	var all []string
	for _, content := range library {
		flat := strings.Join(strings.Fields(content), " ")
		for _, s := range sentenceRe.FindAllString(flat, -1) {
			if s = strings.TrimSpace(s); len([]rune(s)) >= 8 && len([]rune(s)) <= 220 {
				all = append(all, s)
			}
		}
	}
	used := map[string]bool{}
	var out []string
	for _, w := range shuffle(words, rng) {
		var candidates []string
		for _, s := range all {
			if !used[s] && containsWord(s, w) {
				candidates = append(candidates, s)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		s := candidates[rng.IntN(len(candidates))]
		used[s] = true
		out = append(out, s)
		if len(out) == MaxSentences {
			break
		}
	}
	return out
}

// containsWord cherche w comme mot entier dans s. Le trait d'union lie les
// mots composés (« maître-nageur » ne contient pas « maître ») ; l'apostrophe
// les sépare (« l'être » contient « être »).
func containsWord(s, w string) bool {
	hay, needle := []rune(text.SearchKey(s)), []rune(text.SearchKey(w))
	isWordRune := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' }
	for i := 0; i+len(needle) <= len(hay); i++ {
		if string(hay[i:i+len(needle)]) != string(needle) {
			continue
		}
		before := i == 0 || !isWordRune(hay[i-1])
		after := i+len(needle) == len(hay) || !isWordRune(hay[i+len(needle)])
		if before && after {
			return true
		}
	}
	return false
}
