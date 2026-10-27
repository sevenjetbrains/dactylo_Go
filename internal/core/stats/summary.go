// Package stats calcule les mesures d'une session (glossaire, section 1).
package stats

import "dactylo/internal/core/typing"

// CharsPerWord est la convention standard : 1 mot = 5 frappes.
const CharsPerWord = 5

// Summary résume une session.
type Summary struct {
	DurationMs        int64   `json:"durationMs"`
	Keystrokes        int     `json:"keystrokes"`        // frappes de caractères (hors Retour arrière)
	CorrectKeystrokes int     `json:"correctKeystrokes"` // frappes justes du premier coup ou non
	Backspaces        int     `json:"backspaces"`
	CorrectChars      int     `json:"correctChars"`      // positions justes en fin de session
	UncorrectedErrors int     `json:"uncorrectedErrors"` // positions restées fausses
	WPMGross          float64 `json:"wpmGross"`          // frappes / 5 / minute
	WPMNet            float64 `json:"wpmNet"`            // brut − erreurs non corrigées / minute
	CPM               float64 `json:"cpm"`               // caractères justes / minute
	Accuracy          float64 `json:"accuracy"`          // frappes justes / frappes, en %
}

// Summarize calcule les mesures à partir des frappes corrigées et de la
// durée effective (pauses exclues).
func Summarize(scored []typing.Scored, durationMs int64) Summary {
	s := Summary{DurationMs: durationMs}
	for _, k := range scored {
		if k.IsBackspace() {
			s.Backspaces++
			continue
		}
		s.Keystrokes++
		if k.Correct {
			s.CorrectKeystrokes++
		}
	}
	final := typing.Replay(scored)
	s.CorrectChars = final.CorrectChars()
	s.UncorrectedErrors = final.UncorrectedErrors()

	if s.Keystrokes > 0 {
		s.Accuracy = 100 * float64(s.CorrectKeystrokes) / float64(s.Keystrokes)
	}
	if durationMs > 0 {
		minutes := float64(durationMs) / 60000
		s.WPMGross = float64(s.Keystrokes) / CharsPerWord / minutes
		s.WPMNet = max(0, s.WPMGross-float64(s.UncorrectedErrors)/minutes)
		s.CPM = float64(s.CorrectChars) / minutes
	}
	return s
}
