package typing

import (
	"sort"
	"unicode/utf8"
)

// Backspace est la valeur de Typed pour une frappe Retour arrière.
const Backspace = "\b"

// Keystroke est une frappe brute envoyée par l'interface (F-31).
// Position est l'indice, en runes, du caractère visé dans le texte ;
// pour un Retour arrière, c'est l'indice du caractère effacé.
type Keystroke struct {
	Seq      int    `json:"seq"`
	Position int    `json:"position"`
	Typed    string `json:"typed"`
	TimeMs   int64  `json:"timeMs"` // depuis le début de la session
}

// Scored est une frappe corrigée par le cœur à partir du texte attendu :
// la correction de l'interface n'est jamais reprise telle quelle.
type Scored struct {
	Keystroke
	Expected string `json:"expected"`
	Correct  bool   `json:"correct"`
}

// IsBackspace indique une frappe Retour arrière.
func (s Scored) IsBackspace() bool { return s.Typed == Backspace }

// Score trie les frappes par numéro de séquence et les corrige.
// Une frappe hors du texte est comptée fausse.
func Score(text []rune, ks []Keystroke, opts MatchOptions) []Scored {
	sorted := make([]Keystroke, len(ks))
	copy(sorted, ks)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Seq < sorted[j].Seq })

	out := make([]Scored, len(sorted))
	for i, k := range sorted {
		s := Scored{Keystroke: k}
		inText := k.Position >= 0 && k.Position < len(text)
		if inText {
			s.Expected = string(text[k.Position])
		}
		if k.Typed != Backspace && inText {
			r, _ := utf8.DecodeRuneInString(k.Typed)
			s.Correct = utf8.RuneCountInString(k.Typed) == 1 && Match(text[k.Position], r, opts)
		}
		out[i] = s
	}
	return out
}

// Final est l'état du texte après rejeu de toutes les frappes.
type Final struct {
	// Typed associe chaque position encore remplie à sa correction.
	Typed map[int]bool
	// Reached est la position la plus avancée atteinte (exclue).
	Reached int
}

// Replay rejoue les frappes : une frappe remplit sa position, un Retour
// arrière la vide. Fonctionne pour les modes bloquant et libre.
func Replay(scored []Scored) Final {
	f := Final{Typed: map[int]bool{}}
	for _, s := range scored {
		if s.IsBackspace() {
			delete(f.Typed, s.Position)
			continue
		}
		f.Typed[s.Position] = s.Correct
		if s.Position+1 > f.Reached {
			f.Reached = s.Position + 1
		}
	}
	return f
}

// UncorrectedErrors compte les positions restées fausses en fin de session.
func (f Final) UncorrectedErrors() int {
	n := 0
	for _, ok := range f.Typed {
		if !ok {
			n++
		}
	}
	return n
}

// CorrectChars compte les positions justes en fin de session.
func (f Final) CorrectChars() int {
	n := 0
	for _, ok := range f.Typed {
		if ok {
			n++
		}
	}
	return n
}
