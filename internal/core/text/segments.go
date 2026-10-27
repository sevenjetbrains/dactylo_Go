package text

import "unicode"

// Modes de découpage d'un texte en exercices (F-24).
const (
	SplitNone      = "none"      // le texte entier
	SplitParagraph = "paragraph" // N paragraphes par exercice
	SplitLines     = "lines"     // N lignes par exercice
)

// Segment est un exercice : runes [Start, End) du texte, sans les sauts de
// ligne qui l'entourent.
type Segment struct {
	Index int `json:"index"`
	Start int `json:"start"`
	End   int `json:"end"`
}

type span struct{ start, end int }

// Segments découpe le texte. Un paragraphe est un bloc séparé par une ligne
// vide ; si le texte n'en contient aucune, chaque ligne est un paragraphe.
func Segments(content string, mode string, size int) []Segment {
	runes := []rune(content)
	lines := nonEmptyLines(runes)
	if len(lines) == 0 {
		return nil
	}
	size = max(1, size)

	var blocks []span
	switch mode {
	case SplitLines:
		blocks = lines
	case SplitParagraph:
		blocks = paragraphs(runes, lines)
	default:
		return []Segment{{Index: 0, Start: lines[0].start, End: lines[len(lines)-1].end}}
	}

	var out []Segment
	for i := 0; i < len(blocks); i += size {
		last := min(i+size, len(blocks)) - 1
		out = append(out, Segment{Index: len(out), Start: blocks[i].start, End: blocks[last].end})
	}
	return out
}

// ResumeSegment renvoie l'exercice à reprendre et la position de départ
// pour une progression donnée. Un texte terminé reprend au début.
func ResumeSegment(segs []Segment, position int) (Segment, int) {
	for _, s := range segs {
		if position < s.End {
			return s, max(position, s.Start)
		}
	}
	if len(segs) == 0 {
		return Segment{}, 0
	}
	return segs[0], segs[0].Start
}

// WordStart recule pos jusqu'au début du mot qui le contient.
func WordStart(runes []rune, pos int) int {
	pos = min(max(pos, 0), len(runes))
	for pos > 0 && !unicode.IsSpace(runes[pos-1]) {
		pos--
	}
	return pos
}

// nonEmptyLines renvoie les lignes contenant autre chose que des blancs,
// bornées à leur premier et dernier caractère visible.
func nonEmptyLines(runes []rune) []span {
	var out []span
	start := 0
	for i := 0; i <= len(runes); i++ {
		if i < len(runes) && runes[i] != '\n' {
			continue
		}
		s, e := start, i
		for s < e && unicode.IsSpace(runes[s]) {
			s++
		}
		for e > s && unicode.IsSpace(runes[e-1]) {
			e--
		}
		if s < e {
			out = append(out, span{s, e})
		}
		start = i + 1
	}
	return out
}

func paragraphs(runes []rune, lines []span) []span {
	blank := func(a, b span) bool { // ligne vide entre deux lignes ?
		n := 0
		for _, r := range runes[a.end:b.start] {
			if r == '\n' {
				n++
			}
		}
		return n >= 2
	}
	hasBlank := false
	for i := 1; i < len(lines); i++ {
		if blank(lines[i-1], lines[i]) {
			hasBlank = true
			break
		}
	}
	if !hasBlank {
		return lines
	}
	out := []span{lines[0]}
	for i := 1; i < len(lines); i++ {
		if blank(lines[i-1], lines[i]) {
			out = append(out, lines[i])
		} else {
			out[len(out)-1].end = lines[i].end
		}
	}
	return out
}
