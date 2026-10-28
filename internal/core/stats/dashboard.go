// Calculs du tableau de bord : séries de jours, semaine, statistiques par
// touche et par paire de lettres.

package stats

import (
	"sort"
	"strings"
	"time"
	"unicode"

	"dactylo/internal/core/typing"
)

// Streak compte les jours consécutifs d'entraînement jusqu'à aujourd'hui
// (ou hier : la série n'est pas perdue avant la fin de la journée).
func Streak(days []time.Time, today time.Time) int {
	set := map[string]bool{}
	for _, d := range days {
		set[d.In(today.Location()).Format(time.DateOnly)] = true
	}
	day := today
	if !set[day.Format(time.DateOnly)] {
		day = day.AddDate(0, 0, -1)
	}
	n := 0
	for set[day.Format(time.DateOnly)] {
		n++
		day = day.AddDate(0, 0, -1)
	}
	return n
}

// BestStreak renvoie la plus longue série de jours consécutifs.
func BestStreak(days []time.Time, loc *time.Location) int {
	set := map[string]bool{}
	for _, d := range days {
		set[d.In(loc).Format(time.DateOnly)] = true
	}
	best := 0
	for day := range set {
		t, _ := time.ParseInLocation(time.DateOnly, day, loc)
		if set[t.AddDate(0, 0, -1).Format(time.DateOnly)] {
			continue // pas le début d'une série
		}
		n := 0
		for set[t.Format(time.DateOnly)] {
			n++
			t = t.AddDate(0, 0, 1)
		}
		best = max(best, n)
	}
	return best
}

// WeekStart renvoie lundi 0 h (heure locale) de la semaine de t.
func WeekStart(t time.Time) time.Time {
	offset := (int(t.Weekday()) + 6) % 7 // lundi = 0
	y, m, d := t.AddDate(0, 0, -offset).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// KeyStat mesure une touche (un caractère attendu, en minuscule).
type KeyStat struct {
	Key       string  `json:"key"`
	Count     int     `json:"count"`
	Errors    int     `json:"errors"`
	ErrorRate float64 `json:"errorRate"` // en %
	AvgMs     float64 `json:"avgMs"`     // délai moyen depuis la frappe précédente
}

// BigramStat mesure l'enchaînement de deux lettres.
type BigramStat struct {
	Bigram string  `json:"bigram"`
	Count  int     `json:"count"`
	AvgMs  float64 `json:"avgMs"`
}

// MinBigramCount : un bigramme n'est classé qu'à partir de ce nombre
// d'occurrences, pour éviter les hasards.
const MinBigramCount = 3

// KeyStats calcule, sur plusieurs sessions, le taux d'erreur et le délai
// moyen par touche, et les bigrammes les plus lents (F-34).
func KeyStats(sessions [][]typing.Scored, topBigrams int) ([]KeyStat, []BigramStat) {
	type acc struct {
		count, errors, timed int
		totalMs              int64
	}
	keys := map[string]*acc{}
	bigrams := map[string]*acc{}

	for _, ks := range sessions {
		var prev *typing.Scored
		for i := range ks {
			k := &ks[i]
			if k.IsBackspace() || k.Expected == "" {
				prev = nil
				continue
			}
			key := keyName(k.Expected)
			a := keys[key]
			if a == nil {
				a = &acc{}
				keys[key] = a
			}
			a.count++
			if !k.Correct {
				a.errors++
			}
			if prev != nil {
				dt := k.TimeMs - prev.TimeMs
				a.timed++
				a.totalMs += dt
				if prev.Correct && k.Correct && k.Position == prev.Position+1 {
					if bg, ok := letterBigram(prev.Expected, k.Expected); ok {
						b := bigrams[bg]
						if b == nil {
							b = &acc{}
							bigrams[bg] = b
						}
						b.count++
						b.totalMs += dt
					}
				}
			}
			prev = k
		}
	}

	outKeys := make([]KeyStat, 0, len(keys))
	for k, a := range keys {
		s := KeyStat{Key: k, Count: a.count, Errors: a.errors, ErrorRate: 100 * float64(a.errors) / float64(a.count)}
		if a.timed > 0 {
			s.AvgMs = float64(a.totalMs) / float64(a.timed)
		}
		outKeys = append(outKeys, s)
	}
	sort.Slice(outKeys, func(i, j int) bool { return outKeys[i].Key < outKeys[j].Key })

	var outBigrams []BigramStat
	for bg, a := range bigrams {
		if a.count >= MinBigramCount {
			outBigrams = append(outBigrams, BigramStat{Bigram: bg, Count: a.count, AvgMs: float64(a.totalMs) / float64(a.count)})
		}
	}
	sort.Slice(outBigrams, func(i, j int) bool {
		if outBigrams[i].AvgMs != outBigrams[j].AvgMs {
			return outBigrams[i].AvgMs > outBigrams[j].AvgMs
		}
		return outBigrams[i].Bigram < outBigrams[j].Bigram
	})
	if len(outBigrams) > topBigrams {
		outBigrams = outBigrams[:topBigrams]
	}
	if outBigrams == nil {
		outBigrams = []BigramStat{}
	}
	return outKeys, outBigrams
}

// keyName rattache un caractère à sa touche : minuscule, espace, entrée.
func keyName(expected string) string {
	switch expected {
	case "\n":
		return "↵"
	case " ", " ", " ":
		return " "
	}
	return strings.ToLower(expected)
}

// letterBigram renvoie la paire de lettres, si les deux sont des lettres.
func letterBigram(a, b string) (string, bool) {
	ra, rb := []rune(a), []rune(b)
	if len(ra) != 1 || len(rb) != 1 || !unicode.IsLetter(ra[0]) || !unicode.IsLetter(rb[0]) {
		return "", false
	}
	return strings.ToLower(a + b), true
}
