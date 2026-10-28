// Tentatives par mot d'une session : mots justes, faux ou lents, pour
// les mots difficiles (F-40).

package typing

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"dactylo/internal/core/text"
)

// DefaultSlowThreshold : un mot est lent s'il est tapé plus de 1,5 fois plus
// lentement que la moyenne de la session (F-40, seuil réglable).
const DefaultSlowThreshold = 1.5

// MinSlowGapMs : un mot n'est lent que s'il dépasse aussi la moyenne d'au
// moins 80 ms par caractère. Sans ce plancher, une hésitation de quelques
// millisecondes chez un dactylo rapide suffirait à marquer un mot « lent ».
const MinSlowGapMs = 80

// WordAttempt est un mot entièrement tapé pendant une session.
type WordAttempt struct {
	Word       string  `json:"word"`  // forme normalisée (minuscules, sans ponctuation)
	Shown      string  `json:"shown"` // forme du texte, sans ponctuation autour
	Typed      string  `json:"typed"` // forme tapée au premier essai
	Correct    bool    `json:"correct"`
	ErrorType  string  `json:"errorType"`
	Position   int     `json:"position"`
	DurationMs int64   `json:"durationMs"`
	MsPerChar  float64 `json:"msPerChar"`
	Slow       bool    `json:"slow"`
}

// elisionRe repère l'article ou le pronom élidé en tête de mot
// (« l'élève », « qu'il », « jusqu'à ») : il ne fait pas partie du mot suivi.
var elisionRe = regexp.MustCompile(`(?i)^(?:jusqu|lorsqu|puisqu|quoiqu|qu|[cdjlmnst])['’](\pL)`)

// StripElision retire l'élision en tête : « l'élève » → « élève ».
func StripElision(w string) string {
	return elisionRe.ReplaceAllString(w, "$1")
}

// NormalizeWord donne la clé d'un mot : NFC, minuscules, sans ponctuation
// autour ni élision. Renvoie "" pour ce qui n'est pas un mot (nombre...).
func NormalizeWord(w string) string {
	w = strings.ToLower(StripElision(CoreWord(text.NFC(w))))
	letters := 0
	for _, r := range w {
		if unicode.IsLetter(r) {
			letters++
		}
	}
	if letters < 2 {
		return ""
	}
	return w
}

// WordAttempts liste les mots entièrement atteints, avec leur durée et le
// repérage des mots lents. Le temps d'un mot va de la dernière frappe qui
// le précède à la dernière frappe sur le mot.
func WordAttempts(txt []rune, scored []Scored, slowThreshold float64) []WordAttempt {
	type keyInfo struct {
		first    string
		wrong    bool
		lastTime int64
	}
	keys := map[int]*keyInfo{}
	var times []int64 // horodatage de chaque frappe, dans l'ordre
	var positions []int
	reached := 0
	for _, s := range scored {
		times = append(times, s.TimeMs)
		positions = append(positions, s.Position)
		if s.IsBackspace() {
			continue
		}
		k := keys[s.Position]
		if k == nil {
			k = &keyInfo{first: s.Typed}
			keys[s.Position] = k
		}
		k.wrong = k.wrong || !s.Correct
		k.lastTime = s.TimeMs
		reached = max(reached, s.Position+1)
	}

	// début du mot : horodatage de la frappe qui précède la première frappe
	// dans le mot (ou la première frappe du mot pour le tout premier mot)
	startOf := func(span Span) (int64, int) {
		for i, p := range positions {
			if p >= span.Start && p < span.End {
				if i == 0 {
					return times[0], span.End - span.Start - 1
				}
				return times[i-1], span.End - span.Start
			}
		}
		return 0, 0
	}

	var out []WordAttempt
	for _, w := range Words(txt) {
		if w.End > reached {
			break
		}
		raw := string(txt[w.Start:w.End])
		key := NormalizeWord(raw)
		if key == "" {
			continue
		}
		a := WordAttempt{Word: key, Shown: StripElision(CoreWord(raw)), Position: w.Start, Correct: true}
		var typed strings.Builder
		var last int64
		for p := w.Start; p < w.End; p++ {
			if k := keys[p]; k != nil {
				typed.WriteString(k.first)
				a.Correct = a.Correct && !k.wrong
				last = max(last, k.lastTime)
			} else {
				a.Correct = false // sauté (mode libre puis retour arrière...)
			}
		}
		a.Typed = typed.String()
		if !a.Correct {
			a.ErrorType = Classify(raw, a.Typed)
		}
		start, counted := startOf(w)
		a.DurationMs = max(0, last-start)
		if counted > 0 {
			a.MsPerChar = float64(a.DurationMs) / float64(counted)
		}
		out = append(out, a)
	}
	markSlow(out, slowThreshold)
	return out
}

// markSlow compare chaque mot juste à la moyenne par caractère de la session.
func markSlow(attempts []WordAttempt, threshold float64) {
	if threshold <= 1 {
		threshold = DefaultSlowThreshold
	}
	var sum float64
	n := 0
	for _, a := range attempts {
		if a.MsPerChar > 0 {
			sum += a.MsPerChar * float64(utf8.RuneCountInString(a.Shown))
			n += utf8.RuneCountInString(a.Shown)
		}
	}
	if n == 0 || len(attempts) < 3 { // trop peu de mots pour une moyenne fiable
		return
	}
	avg := sum / float64(n)
	for i := range attempts {
		a := attempts[i]
		attempts[i].Slow = a.Correct && a.MsPerChar > threshold*avg && a.MsPerChar-avg >= MinSlowGapMs
	}
}
