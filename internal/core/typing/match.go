// Package typing compare la frappe au texte attendu : correction des
// frappes, rejeu de la session et repérage des mots fautifs.
package typing

import "dactylo/internal/core/text"

// MatchOptions règle la tolérance de la comparaison.
type MatchOptions struct {
	// FoldTypography assimile ’/' , «»/" et –/- (fréquents dans les PDF).
	FoldTypography bool `json:"foldTypography"`
}

// Match indique si le caractère tapé correspond au caractère attendu.
// Les deux sont supposés en NFC. Une fin de ligne attendue accepte Entrée
// ou Espace ; une espace insécable attendue accepte l'espace ordinaire.
func Match(expected, typed rune, opts MatchOptions) bool {
	if expected == typed {
		return true
	}
	if expected == '\n' {
		return typed == ' '
	}
	if text.IsSpaceLike(expected) {
		return text.IsSpaceLike(typed)
	}
	if opts.FoldTypography {
		return text.FoldTypography(expected) == text.FoldTypography(typed)
	}
	return false
}
