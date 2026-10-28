// Package dictation contient la logique du mode Dictée (F-10 à F-17) :
// découpage de la lecture, ponctuation lue, alignement et correction.
package dictation

import (
	"regexp"
	"strings"
	"unicode"
)

// Découpages de la lecture (F-11).
const (
	ChunkWord      = "word"
	ChunkGroup     = "group" // groupe de N mots (4 par défaut)
	ChunkSentence  = "sentence"
	ChunkParagraph = "paragraph"
)

// DefaultGroupSize est la taille par défaut d'un groupe de mots.
const DefaultGroupSize = 4

// Chunk est un segment lu d'un seul tenant.
type Chunk struct {
	Index  int    `json:"index"`
	Text   string `json:"text"`   // texte du segment
	Spoken string `json:"spoken"` // ce que la voix lit (ponctuation dite, F-13)
	Words  int    `json:"words"`  // nombre de mots (avancement adaptatif, F-15)
}

// word est un mot du texte et la coupure qui le suit.
type word struct {
	text      string
	breakLine bool // suivi d'un saut de ligne
	paragraph bool // suivi d'une ligne vide
}

// Chunks découpe le texte en segments de lecture.
func Chunks(content, mode string, groupSize int, speakPunctuation bool) []Chunk {
	words := splitWords(content)
	if groupSize < 1 {
		groupSize = DefaultGroupSize
	}
	var groups [][]word
	var cur []word
	flush := func() {
		if len(cur) > 0 {
			groups = append(groups, cur)
			cur = nil
		}
	}
	for _, w := range words {
		cur = append(cur, w)
		switch mode {
		case ChunkWord:
			flush()
		case ChunkSentence:
			if endsSentence(w.text) || w.paragraph {
				flush()
			}
		case ChunkParagraph:
			if w.paragraph {
				flush()
			}
		default: // groupe : on ne franchit pas une fin de phrase
			if len(cur) == groupSize || endsSentence(w.text) || w.paragraph {
				flush()
			}
		}
	}
	flush()

	out := make([]Chunk, len(groups))
	for i, g := range groups {
		parts := make([]string, len(g))
		for j, w := range g {
			parts[j] = w.text
		}
		c := Chunk{Index: i, Text: strings.Join(parts, " "), Words: len(g)}
		c.Spoken = c.Text
		if speakPunctuation {
			c.Spoken = Spoken(g)
		}
		out[i] = c
	}
	return out
}

// CountWords compte les mots d'un texte tapé (avancement adaptatif).
func CountWords(s string) int { return len(strings.Fields(s)) }

// splitWords découpe le texte en mots en notant fins de ligne et paragraphes.
func splitWords(content string) []word {
	var out []word
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for i, line := range lines {
		fields := strings.Fields(line)
		for _, f := range fields {
			out = append(out, word{text: f})
		}
		if len(out) == 0 || len(fields) == 0 {
			if len(out) > 0 && i > 0 && strings.TrimSpace(line) == "" {
				out[len(out)-1].paragraph = true
			}
			continue
		}
		if i < len(lines)-1 {
			out[len(out)-1].breakLine = true
		}
	}
	return out
}

// abbreviationRe : abréviations courantes et initiales (« M. », « Mme. »,
// « J. ») qui ne terminent pas une phrase et se lisent telles quelles.
var abbreviationRe = regexp.MustCompile(`^(\p{Lu}|MM|Mme|Mmes|Mlle|Mlles|Mgr|Me|Dr|Pr|St|Ste|cf|ex|p|pp|av|apr|env|vol|chap|fig|n°)\.$`)

func isAbbreviation(w string) bool { return abbreviationRe.MatchString(w) }

// endsSentence indique si le mot termine une phrase (hors abréviations).
func endsSentence(w string) bool {
	if isAbbreviation(w) {
		return false
	}
	w = strings.TrimRightFunc(w, func(r rune) bool { return r == '»' || r == '"' || r == ')' })
	if w == "" {
		return false
	}
	r := []rune(w)
	return strings.ContainsRune(".!?…", r[len(r)-1])
}

// Ponctuation dite à voix haute, comme une dictée scolaire (F-13).
var spokenPunct = map[rune]string{
	',': "virgule",
	'.': "point",
	';': "point-virgule",
	':': "deux-points",
	'?': "point d'interrogation",
	'!': "point d'exclamation",
	'…': "points de suspension",
	'«': "ouvrez les guillemets",
	'»': "fermez les guillemets",
	'(': "ouvrez la parenthèse",
	')': "fermez la parenthèse",
	'—': "tiret",
	'–': "tiret",
}

// Spoken renvoie le texte à lire avec la ponctuation dite.
func Spoken(words []word) string {
	var parts []string
	for _, w := range words {
		if isAbbreviation(w.text) { // « M. » se lit « monsieur », pas « M point »
			parts = append(parts, w.text)
			if w.paragraph || w.breakLine {
				parts = append(parts, "à la ligne")
			}
			continue
		}
		runes := []rune(w.text)
		// ponctuation ouvrante en tête de mot
		start := 0
		for start < len(runes) && isSpokenPunct(runes[start]) && (runes[start] == '«' || runes[start] == '(' || runes[start] == '—' || runes[start] == '–') {
			parts = append(parts, spokenPunct[runes[start]])
			start++
		}
		end := len(runes)
		for end > start && isSpokenPunct(runes[end-1]) {
			end--
		}
		if core := strings.TrimFunc(string(runes[start:end]), unicode.IsSpace); core != "" {
			parts = append(parts, core)
		}
		// « ... » se dit une seule fois
		trail := strings.Replace(string(runes[end:]), "...", "…", 1)
		for _, r := range trail {
			parts = append(parts, spokenPunct[r])
		}
		if w.paragraph || w.breakLine {
			parts = append(parts, "à la ligne")
		}
	}
	return strings.Join(parts, " ")
}

func isSpokenPunct(r rune) bool {
	_, ok := spokenPunct[r]
	return ok
}
