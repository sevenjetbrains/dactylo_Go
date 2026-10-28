// Clés de recherche insensibles aux accents et à la casse.

package text

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// StripAccents retire les diacritiques : « élève » → « eleve ».
func StripAccents(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SearchKey prépare une chaîne pour une recherche insensible à la casse et
// aux accents.
func SearchKey(s string) string {
	return strings.ToLower(StripAccents(s))
}

// Contains indique si haystack contient needle, sans tenir compte de la
// casse ni des accents.
func Contains(haystack, needle string) bool {
	return strings.Contains(SearchKey(haystack), SearchKey(strings.TrimSpace(needle)))
}
