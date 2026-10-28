package dictation

import (
	"strings"
	"unicode/utf8"

	"dactylo/internal/core/text"
	"dactylo/internal/core/typing"
)

// Op est un mot de la correction : juste, faux, oublié ou en trop.
type Op struct {
	Expected string `json:"expected"` // "" pour un mot en trop
	Typed    string `json:"typed"`    // "" pour un mot oublié
	Correct  bool   `json:"correct"`
	Type     string `json:"type"`     // type de faute (vide si juste)
	Position int    `json:"position"` // indice du mot attendu (-1 pour un mot en trop)
}

// Coûts de l'alignement : un mot proche mal écrit est une substitution, un
// mot sans rapport coûte presque autant qu'un oubli suivi d'un ajout.
const (
	costOmit      = 1.0
	costExtra     = 1.0
	costNear      = 0.4 // accent, majuscule, ponctuation
	costSimilar   = 0.8 // quelques lettres de différence
	costUnrelated = 1.6
	costMerge     = 0.4 // « soixante-quinze » tapé en deux mots
)

// Align aligne mot à mot le texte tapé sur le texte attendu (distance
// d'édition sur les mots) et classe chaque faute (F-16).
func Align(expected, typed string) []Op {
	e := strings.Fields(text.NFC(expected))
	t := strings.Fields(text.NFC(typed))
	n, m := len(e), len(t)

	// cost[i][j] : coût minimal pour aligner e[i:] et t[j:]
	cost := make([][]float64, n+1)
	for i := range cost {
		cost[i] = make([]float64, m+1)
	}
	for i := n; i >= 0; i-- {
		for j := m; j >= 0; j-- {
			switch {
			case i == n && j == m:
				cost[i][j] = 0
			case i == n:
				cost[i][j] = costExtra + cost[i][j+1]
			case j == m:
				cost[i][j] = costOmit + cost[i+1][j]
			default:
				best := subCost(e[i], t[j]) + cost[i+1][j+1]
				best = min(best, costOmit+cost[i+1][j], costExtra+cost[i][j+1])
				if j+1 < m && mergeable(e[i], t[j], t[j+1]) {
					best = min(best, costMerge+cost[i+1][j+2])
				}
				cost[i][j] = best
			}
		}
	}

	var ops []Op
	i, j := 0, 0
	const eps = 1e-9
	for i < n || j < m {
		switch {
		case i < n && j < m && near(cost[i][j], subCost(e[i], t[j])+cost[i+1][j+1], eps):
			ops = append(ops, op(e[i], t[j], i))
			i, j = i+1, j+1
		case i < n && j+1 < m && mergeable(e[i], t[j], t[j+1]) && near(cost[i][j], costMerge+cost[i+1][j+2], eps):
			ops = append(ops, op(e[i], t[j]+" "+t[j+1], i))
			i, j = i+1, j+2
		case i < n && near(cost[i][j], costOmit+cost[i+1][j], eps):
			ops = append(ops, Op{Expected: e[i], Type: typing.ErrOmittedWord, Position: i})
			i++
		default:
			ops = append(ops, Op{Typed: t[j], Type: typing.ErrExtraWord, Position: -1})
			j++
		}
	}
	return ops
}

func op(expected, typed string, pos int) Op {
	if expected == typed {
		return Op{Expected: expected, Typed: typed, Correct: true, Position: pos}
	}
	return Op{Expected: expected, Typed: typed, Type: typing.Classify(expected, typed), Position: pos}
}

func near(a, b, eps float64) bool { return a-b < eps && b-a < eps }

func subCost(e, t string) float64 {
	if e == t {
		return 0
	}
	switch typing.Classify(e, t) {
	case typing.ErrCase, typing.ErrAccent, typing.ErrPunctuation:
		return costNear
	case typing.ErrMissingLetter, typing.ErrExtraLetter, typing.ErrTransposition:
		return costSimilar
	}
	// substitution : proche si moins de la moitié des lettres diffère
	d := levenshtein(text.SearchKey(e), text.SearchKey(t))
	if 2*d <= max(utf8.RuneCountInString(e), utf8.RuneCountInString(t)) {
		return costSimilar
	}
	return costUnrelated
}

// mergeable : un mot attendu tapé en deux (« soixante-quinze » →
// « soixante quinze », « aujourd'hui » → « aujourd hui »).
func mergeable(e, t1, t2 string) bool {
	return lettersKey(e) == lettersKey(t1+t2) && lettersKey(t1) != "" && lettersKey(t2) != ""
}

func lettersKey(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '-' || r == '\'' || r == '’' || r == ' ' {
			return -1
		}
		return r
	}, text.SearchKey(s))
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			c := 1
			if ra[i-1] == rb[j-1] {
				c = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+c)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}

// Report résume une correction.
type Report struct {
	Ops      []Op           `json:"ops"`
	Words    int            `json:"words"`    // mots attendus
	Faults   int            `json:"faults"`   // mots faux, oubliés ou en trop
	Accuracy float64        `json:"accuracy"` // % de mots justes
	ByType   map[string]int `json:"byType"`
}

// Correct aligne et compte les fautes par type.
func Correct(expected, typed string) Report {
	ops := Align(expected, typed)
	r := Report{Ops: ops, ByType: map[string]int{}}
	for _, o := range ops {
		if o.Expected != "" {
			r.Words++
		}
		if !o.Correct {
			r.Faults++
			r.ByType[o.Type]++
		}
	}
	if r.Words > 0 {
		r.Accuracy = max(0, 100*float64(r.Words-r.Faults)/float64(r.Words))
	}
	return r
}

// Attempts convertit la correction en tentatives de mots pour les
// révisions (les mots en trop n'ont pas de mot attendu : ignorés).
func Attempts(ops []Op) []typing.WordAttempt {
	var out []typing.WordAttempt
	for _, o := range ops {
		key := typing.NormalizeWord(o.Expected)
		if key == "" {
			continue
		}
		out = append(out, typing.WordAttempt{
			Word: key, Shown: typing.StripElision(typing.CoreWord(o.Expected)), Typed: o.Typed,
			Correct: o.Correct, ErrorType: o.Type, Position: o.Position,
		})
	}
	return out
}
