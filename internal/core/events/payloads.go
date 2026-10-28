// Contenu des événements de domaine échangés entre modules.

package events

import (
	"dactylo/internal/core/stats"
	"dactylo/internal/core/typing"
)

// SessionCompletedPayload accompagne SessionCompleted.
type SessionCompletedPayload struct {
	SessionID int64         `json:"sessionId"`
	ProfileID int64         `json:"profileId"`
	TextID    int64         `json:"textId"`
	Mode      string        `json:"mode"`
	Summary   stats.Summary `json:"summary"`
	// Mots entièrement tapés, avec leur résultat (F-40) : le module
	// Révision s'en sert pour tenir la liste des mots difficiles.
	Words []typing.WordAttempt `json:"words"`
}

// WordMistypedPayload accompagne WordMistyped (un événement par mot fautif).
type WordMistypedPayload struct {
	SessionID int64            `json:"sessionId"`
	ProfileID int64            `json:"profileId"`
	Error     typing.WordError `json:"error"`
}

// TextImportedPayload accompagne TextImported.
type TextImportedPayload struct {
	TextID int64  `json:"textId"`
	Source string `json:"source"`
}
