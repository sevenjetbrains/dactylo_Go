package text

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// NFC renvoie s en forme normale composée : « e + ◌̂ » devient « ê »,
// une seule unité à taper et à comparer.
func NFC(s string) string {
	return norm.NFC.String(s)
}

// IsSpaceLike indique les espaces qu'on ne sait pas taper au clavier
// (insécables, fines) et qui valent une espace ordinaire.
func IsSpaceLike(r rune) bool {
	switch r {
	case ' ', '\t', ' ', ' ', ' ', ' ':
		return true
	}
	return false
}

// FoldTypography ramène apostrophes, guillemets et tirets typographiques
// à leur équivalent tapable (option F-22 / section 2.3).
func FoldTypography(r rune) rune {
	switch r {
	case '’', '‘', '‛', 'ʼ', '´', '`':
		return '\''
	case '«', '»', '“', '”', '„', '‟':
		return '"'
	case '–', '—', '−':
		return '-'
	}
	return r
}

// CleanForPractice prépare un texte collé : NFC, fins de ligne Unix,
// tabulations en espaces, espaces de fin de ligne supprimées, au plus une
// ligne vide entre deux paragraphes.
func CleanForPractice(s string) string {
	s = NFC(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\t", " ")

	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	blank := 0
	for _, l := range lines {
		l = strings.TrimRight(l, "   ")
		if l == "" {
			blank++
			if blank > 1 {
				continue
			}
		} else {
			blank = 0
		}
		out = append(out, l)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
