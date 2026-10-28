// Nettoyage d'un texte importé d'un PDF : en-têtes et pieds de page
// répétés, numéros de page, mots coupés, paragraphes (F-22).

package text

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ImportPage est une page extraite d'un document (PDF...), ligne par ligne.
type ImportPage struct {
	Number int      `json:"number"`
	Lines  []string `json:"lines"`
}

// CleanOptions règle le nettoyage d'un texte importé (F-22).
type CleanOptions struct {
	RemovePageNumbers    bool `json:"removePageNumbers"`
	RemoveRepeatedLines  bool `json:"removeRepeatedLines"` // en-têtes et pieds de page
	JoinHyphenated       bool `json:"joinHyphenated"`      // « exem-\nple » → « exemple »
	NormalizeApostrophes bool `json:"normalizeApostrophes"`
}

// DefaultCleanOptions active tous les nettoyages.
var DefaultCleanOptions = CleanOptions{
	RemovePageNumbers: true, RemoveRepeatedLines: true, JoinHyphenated: true, NormalizeApostrophes: true,
}

// edgeLines est le nombre de lignes, en haut et en bas de page, où l'on
// cherche numéros de page, en-têtes et pieds de page.
const edgeLines = 2

var (
	pageNumberRe = regexp.MustCompile(`(?i)^(page|p\.)?\s*[-–—]?\s*\d{1,4}\s*[-–—]?\s*((/|sur|of)\s*\d{1,4})?$`)
	spacesRe     = regexp.MustCompile(`[ \t\x{00A0}\x{202F}\x{2007}\x{2009}]+`)
	digitsRe     = regexp.MustCompile(`\d+`)
)

// CleanReport est le texte nettoyé et le compte rendu du nettoyage
// (affiché dans l'aperçu de l'import).
type CleanReport struct {
	Text         string   `json:"text"`
	Joined       []string `json:"joined"`       // mots recollés après une césure
	RemovedLines int      `json:"removedLines"` // numéros de page, en-têtes, pieds de page
}

// CleanImported transforme des pages extraites en texte à taper : lignes
// recollées en paragraphes (séparés par une ligne vide), césures
// supprimées, numéros de page et lignes répétées retirés.
func CleanImported(pages []ImportPage, opts CleanOptions) CleanReport {
	var report CleanReport
	cleaned := make([][]string, len(pages))
	for i, p := range pages {
		lines := make([]string, 0, len(p.Lines))
		for _, l := range p.Lines {
			l = strings.TrimSpace(spacesRe.ReplaceAllString(NFC(l), " "))
			if opts.NormalizeApostrophes {
				l = strings.NewReplacer("’", "'", "‘", "'", "ʼ", "'").Replace(l)
			}
			lines = append(lines, l)
		}
		cleaned[i] = lines
	}

	if opts.RemovePageNumbers {
		for i := range cleaned {
			forEachEdge(cleaned[i], func(j int) {
				if pageNumberRe.MatchString(cleaned[i][j]) {
					cleaned[i][j] = ""
					report.RemovedLines++
				}
			})
		}
	}
	if opts.RemoveRepeatedLines && len(cleaned) >= 2 {
		report.RemovedLines += removeRepeated(cleaned)
	}

	var all []string
	for _, lines := range cleaned {
		all = append(all, lines...)
	}
	report.Text, report.Joined = joinParagraphs(all, opts.JoinHyphenated)
	if report.Joined == nil {
		report.Joined = []string{}
	}
	return report
}

// forEachEdge appelle fn sur les premières et dernières lignes non vides.
func forEachEdge(lines []string, fn func(i int)) {
	seen := map[int]bool{}
	visit := func(order []int) {
		n := 0
		for _, i := range order {
			if lines[i] == "" {
				continue
			}
			if !seen[i] {
				seen[i] = true
				fn(i)
			}
			if n++; n == edgeLines {
				return
			}
		}
	}
	idx := make([]int, len(lines))
	rev := make([]int, len(lines))
	for i := range lines {
		idx[i] = i
		rev[i] = len(lines) - 1 - i
	}
	visit(idx)
	visit(rev)
}

// removeRepeated retire les lignes de bord présentes sur au moins la moitié
// des pages (titre courant, nom d'auteur, « Chapitre 3 – p. 12 »...).
func removeRepeated(pages [][]string) (removed int) {
	key := func(s string) string {
		return strings.TrimSpace(digitsRe.ReplaceAllString(SearchKey(s), "#"))
	}
	count := map[string]int{}
	for _, lines := range pages {
		onPage := map[string]bool{}
		forEachEdge(lines, func(j int) {
			if k := key(lines[j]); utf8.RuneCountInString(k) >= 3 {
				onPage[k] = true
			}
		})
		for k := range onPage {
			count[k]++
		}
	}
	threshold := max(2, (len(pages)+1)/2)
	for _, lines := range pages {
		forEachEdge(lines, func(j int) {
			if count[key(lines[j])] >= threshold {
				lines[j] = ""
				removed++
			}
		})
	}
	return removed
}

// joinParagraphs recolle les lignes. Une ligne vide, ou une ligne courte
// terminée par une ponctuation forte, termine le paragraphe. Renvoie aussi
// les mots recollés après une césure.
func joinParagraphs(lines []string, joinHyphenated bool) (string, []string) {
	var joined []string
	short := shortLineThreshold(lines)
	var paragraphs []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			paragraphs = append(paragraphs, cur.String())
			cur.Reset()
		}
	}
	for _, l := range lines {
		if l == "" {
			// une ligne vide ne coupe que si le paragraphe semble terminé :
			// les numéros de page retirés laissent des vides en milieu de phrase
			if endsSentence(cur.String()) {
				flush()
			}
			continue
		}
		if cur.Len() > 0 {
			prev := cur.String()
			if joinHyphenated && endsWithHyphen(prev) && startsLower(l) {
				s := strings.TrimSuffix(strings.TrimSuffix(prev, "-"), "\u00ad")
				cur.Reset()
				cur.WriteString(s)
				joined = append(joined, joinedWord(s, l))
			} else {
				cur.WriteByte(' ')
			}
		}
		cur.WriteString(l)
		if endsSentence(l) && utf8.RuneCountInString(l) < short {
			flush()
		}
	}
	flush()
	return strings.Join(paragraphs, "\n\n"), joined
}

// joinedWord reconstitue le mot recollé : fin de la ligne + début de la suivante.
func joinedWord(before, after string) string {
	head := before[strings.LastIndexByte(before, ' ')+1:]
	tail, _, _ := strings.Cut(after, " ")
	return strings.TrimFunc(head+tail, func(r rune) bool { return !unicode.IsLetter(r) && r != '-' && r != '\'' })
}

// shortLineThreshold : une ligne plus courte que 75 % de la longueur
// habituelle (80e centile) est une fin de paragraphe probable.
func shortLineThreshold(lines []string) int {
	var lens []int
	for _, l := range lines {
		if l != "" {
			lens = append(lens, utf8.RuneCountInString(l))
		}
	}
	if len(lens) < 3 {
		return 0 // trop peu de lignes : seules les lignes vides coupent
	}
	sort.Ints(lens)
	return lens[len(lens)*8/10] * 3 / 4
}

// endsSentence indique si une ligne se termine par une ponctuation forte.
func endsSentence(s string) bool {
	s = strings.TrimRight(s, " ")
	if s == "" {
		return true
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return strings.ContainsRune(".!?:…»\"", r)
}

// endsWithHyphen détecte un mot coupé en fin de ligne.
func endsWithHyphen(s string) bool {
	if strings.HasSuffix(s, "\u00ad") { // trait d'union conditionnel
		return true
	}
	if !strings.HasSuffix(s, "-") || len(s) < 2 {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(s[:len(s)-1])
	return unicode.IsLetter(r)
}

func startsLower(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsLower(r)
}
