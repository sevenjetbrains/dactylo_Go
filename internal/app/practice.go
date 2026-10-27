package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"dactylo/internal/core/events"
	"dactylo/internal/core/stats"
	"dactylo/internal/core/typing"
	"dactylo/internal/store"
)

// AverageWindow est le nombre de sessions de la moyenne de comparaison (F-32).
const AverageWindow = 10

// SessionResult est affiché sur l'écran de fin de session (F-32).
type SessionResult struct {
	SessionID int64              `json:"sessionId"`
	TextID    int64              `json:"textId"`
	Summary   stats.Summary      `json:"summary"`
	Mistakes  []typing.WordError `json:"mistakes"`
	Average   store.Averages     `json:"average"`
}

type activeSession struct {
	profileID int64
	textID    int64
	text      []rune
	opts      typing.MatchOptions
}

// PracticeService gère les sessions du mode Copie : création, frappes
// envoyées par lots, calcul final.
type PracticeService struct {
	store     *store.Store
	bus       *events.Bus
	profileID int64
	now       func() time.Time

	mu     sync.Mutex
	active map[int64]*activeSession
}

// NewPracticeService crée le service pour le profil donné.
func NewPracticeService(s *store.Store, bus *events.Bus, profileID int64) *PracticeService {
	return &PracticeService{store: s, bus: bus, profileID: profileID, now: time.Now, active: map[int64]*activeSession{}}
}

// StartSession ouvre une session sur un texte. settings est le JSON des
// réglages utilisés, conservé tel quel ; seul foldTypography est lu ici.
func (p *PracticeService) StartSession(textID int64, settings string) (int64, error) {
	ctx := context.Background()
	t, err := p.store.GetText(ctx, textID)
	if err != nil {
		return 0, fmt.Errorf("texte %d : %w", textID, err)
	}
	var opts typing.MatchOptions
	if settings != "" {
		if err := json.Unmarshal([]byte(settings), &opts); err != nil {
			return 0, fmt.Errorf("réglages invalides : %w", err)
		}
	}
	id, err := p.store.CreateSession(ctx, p.profileID, textID, "copy", settings, p.now())
	if err != nil {
		return 0, err
	}
	p.mu.Lock()
	p.active[id] = &activeSession{profileID: p.profileID, textID: textID, text: []rune(t.Content), opts: opts}
	p.mu.Unlock()
	return id, nil
}

// RecordKeystrokes corrige et enregistre un lot de frappes (toutes les 2 s).
func (p *PracticeService) RecordKeystrokes(sessionID int64, ks []typing.Keystroke) error {
	a, err := p.session(sessionID)
	if err != nil {
		return err
	}
	return p.store.SaveKeystrokes(context.Background(), sessionID, typing.Score(a.text, ks, a.opts))
}

// FinishSession calcule les mesures à partir des frappes enregistrées,
// clôt la session et émet SessionCompleted puis un WordMistyped par mot.
// durationMs est le temps effectif, pauses exclues.
func (p *PracticeService) FinishSession(sessionID int64, durationMs int64) (SessionResult, error) {
	ctx := context.Background()
	a, err := p.session(sessionID)
	if err != nil {
		return SessionResult{}, err
	}
	if durationMs < 0 {
		return SessionResult{}, errors.New("durée négative")
	}
	raw, err := p.store.Keystrokes(ctx, sessionID)
	if err != nil {
		return SessionResult{}, err
	}
	scored := typing.Score(a.text, raw, a.opts)
	sum := stats.Summarize(scored, durationMs)
	mistakes := typing.MistypedWords(a.text, scored)
	if mistakes == nil {
		mistakes = []typing.WordError{}
	}

	if err := p.store.FinishSession(ctx, sessionID, p.now(), sum, mistakes); err != nil {
		return SessionResult{}, err
	}
	p.mu.Lock()
	delete(p.active, sessionID)
	p.mu.Unlock()

	avg, err := p.store.RecentAverages(ctx, a.profileID, "copy", sessionID, AverageWindow)
	if err != nil {
		return SessionResult{}, err
	}

	p.bus.Publish(events.Event{Name: events.SessionCompleted, Payload: events.SessionCompletedPayload{
		SessionID: sessionID, ProfileID: a.profileID, TextID: a.textID, Mode: "copy", Summary: sum,
	}})
	for _, m := range mistakes {
		p.bus.Publish(events.Event{Name: events.WordMistyped, Payload: events.WordMistypedPayload{
			SessionID: sessionID, ProfileID: a.profileID, Error: m,
		}})
	}

	return SessionResult{SessionID: sessionID, TextID: a.textID, Summary: sum, Mistakes: mistakes, Average: avg}, nil
}

func (p *PracticeService) session(id int64) (*activeSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	a, ok := p.active[id]
	if !ok {
		return nil, fmt.Errorf("session %d inconnue ou terminée", id)
	}
	return a, nil
}
