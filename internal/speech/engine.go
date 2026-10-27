// Package speech définit l'interface de synthèse vocale. En V1 la voix est
// produite par la Web Speech API côté interface ; une implémentation Go
// (espeak-ng, Piper) pourra satisfaire la même interface plus tard.
package speech

import "context"

// Voice décrit une voix disponible.
type Voice struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Language string `json:"language"` // ex. "fr-FR"
}

// Options règle une lecture (F-12).
type Options struct {
	VoiceID string  `json:"voiceId"`
	Rate    float64 `json:"rate"`   // 0,5 à 2
	Pitch   float64 `json:"pitch"`  // 0 à 2
	Volume  float64 `json:"volume"` // 0 à 1
}

// SpeechEngine lit un texte à voix haute.
type SpeechEngine interface {
	Voices(ctx context.Context) ([]Voice, error)
	Speak(ctx context.Context, text string, opts Options) error
	Stop() error
}
