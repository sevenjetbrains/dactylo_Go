// Découpage en mots, mots mal tapés et classement des erreurs (accent,
// majuscule, lettre manquante...).

package typing

import (
	"strings"
	"unicode"

	"dactylo/internal/core/text"
)

// Span délimite un mot dans le texte, en runes : [Start, End).
type Span struct {
	Start int
	End   int
}

// Words découpe le texte en mots séparés par des blancs. La ponctuation
// collée fait partie du mot (« bonjour, » est un seul mot à taper).
func Words(text []rune) []Span {
	var spans []Span
	start := -1
	for i, r := range text {
		if unicode.IsSpace(r) {
			if start >= 0 {
				spans = append(spans, Span{start, i})
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		spans = append(spans, Span{start, len(text)})
	}
	return spans
}

// Types de faute (F-16). La copie produit surtout accent, majuscule,
// ponctuation, inversion et substitution ; la dictée utilisera aussi les
// lettres manquantes ou en trop et les mots oubliés.
const (
	ErrAccent        = "accent"
	ErrCase          = "case"
	ErrPunctuation   = "punctuation"
	ErrMissingLetter = "missing_letter"
	ErrExtraLetter   = "extra_letter"
	ErrTransposition = "transposition"
	ErrSubstitution  = "substitution"
	ErrOmittedWord   = "omitted_word"
	ErrExtraWord     = "extra_word" // dictée : mot tapé sans équivalent attendu
)

// WordError décrit un mot mal tapé.
type WordError struct {
	Expected string `json:"expected"`
	Typed    string `json:"typed"`
	Position int    `json:"position"`
	Type     string `json:"type"`
}

// MistypedWords liste les mots où au moins une frappe a été fausse, même
// corrigée ensuite. Typed est la forme tapée au premier essai, position par
// position. Seuls les mots entièrement atteints sont pris en compte.
func MistypedWords(text []rune, scored []Scored) []WordError {
	first := map[int]string{}
	wrong := map[int]bool{}
	reached := 0
	for _, s := range scored {
		if s.IsBackspace() {
			continue
		}
		if _, seen := first[s.Position]; !seen {
			first[s.Position] = s.Typed
		}
		if !s.Correct {
			wrong[s.Position] = true
		}
		if s.Position+1 > reached {
			reached = s.Position + 1
		}
	}

	var out []WordError
	for _, w := range Words(text) {
		if w.End > reached {
			break
		}
		bad := false
		var typed strings.Builder
		for p := w.Start; p < w.End; p++ {
			if wrong[p] {
				bad = true
			}
			typed.WriteString(first[p])
		}
		if !bad {
			continue
		}
		exp := string(text[w.Start:w.End])
		out = append(out, WordError{
			Expected: exp,
			Typed:    typed.String(),
			Position: w.Start,
			Type:     Classify(exp, typed.String()),
		})
	}
	return out
}

// CoreWord retire la ponctuation en début et fin de mot (« l'été, » → « l'été »).
func CoreWord(w string) string {
	return strings.TrimFunc(w, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r)
	})
}

// Classify donne le type dominant de la faute entre le mot attendu et le
// mot tapé.
func Classify(expected, typed string) string {
	switch {
	case typed == "":
		return ErrOmittedWord
	case expected == "":
		return ErrExtraWord
	case strings.EqualFold(expected, typed):
		return ErrCase
	case strings.EqualFold(text.StripAccents(expected), text.StripAccents(typed)):
		return ErrAccent
	case lettersOnly(expected) == lettersOnly(typed):
		// « chat, » / « chat. », « soixante-quinze » / « soixante quinze »
		return ErrPunctuation
	}
	e, t := []rune(expected), []rune(typed)
	switch {
	case len(t) == len(e)-1 && isOneDeletion(e, t):
		return ErrMissingLetter
	case len(t) == len(e)+1 && isOneDeletion(t, e):
		return ErrExtraLetter
	case len(t) == len(e) && isTransposition(e, t):
		return ErrTransposition
	}
	return ErrSubstitution
}

// lettersOnly garde lettres et chiffres (ponctuation, tirets et espaces retirés).
func lettersOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// isOneDeletion indique si short s'obtient en retirant une rune de long.
func isOneDeletion(long, short []rune) bool {
	i := 0
	for i < len(short) && long[i] == short[i] {
		i++
	}
	return string(long[i+1:]) == string(short[i:])
}

// isTransposition indique deux runes voisines inversées et rien d'autre.
func isTransposition(a, b []rune) bool {
	diff := []int{}
	for i := range a {
		if a[i] != b[i] {
			diff = append(diff, i)
		}
	}
	return len(diff) == 2 && diff[1] == diff[0]+1 &&
		a[diff[0]] == b[diff[1]] && a[diff[1]] == b[diff[0]]
}
