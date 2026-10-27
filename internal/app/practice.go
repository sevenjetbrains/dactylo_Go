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
	"dactylo/internal/core/text"
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

// StartRequest ouvre une session sur un exercice : les runes [Start, End)
// du texte (End = 0 : jusqu'à la fin).
type StartRequest struct {
	TextID   int64  `json:"textId"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Settings string `json:"settings"` // JSON des réglages utilisés
}

type activeSession struct {
	profileID int64
	textID    int64
	full      []rune // texte complet
	start     int    // début de l'exercice dans le texte
	text      []rune // exercice tapé ; les positions des frappes y sont relatives
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

// StartSession ouvre une session sur un exercice. Les réglages sont
// conservés avec les bornes de l'exercice ; seul foldTypography est lu ici.
func (p *PracticeService) StartSession(req StartRequest) (int64, error) {
	ctx := context.Background()
	t, err := p.store.GetText(ctx, req.TextID)
	if err != nil {
		return 0, fmt.Errorf("texte %d : %w", req.TextID, err)
	}
	full := []rune(t.Content)
	end := req.End
	if end == 0 {
		end = len(full)
	}
	if req.Start < 0 || req.Start >= end || end > len(full) {
		return 0, fmt.Errorf("exercice hors du texte : [%d, %d)", req.Start, end)
	}

	settings := map[string]any{}
	if req.Settings != "" {
		if err := json.Unmarshal([]byte(req.Settings), &settings); err != nil {
			return 0, fmt.Errorf("réglages invalides : %w", err)
		}
	}
	fold, _ := settings["foldTypography"].(bool)
	settings["segment"] = map[string]int{"start": req.Start, "end": end}
	raw, err := json.Marshal(settings)
	if err != nil {
		return 0, err
	}

	id, err := p.store.CreateSession(ctx, p.profileID, req.TextID, "copy", string(raw), p.now())
	if err != nil {
		return 0, err
	}
	p.mu.Lock()
	p.active[id] = &activeSession{
		profileID: p.profileID, textID: req.TextID, full: full, start: req.Start,
		text: full[req.Start:end], opts: typing.MatchOptions{FoldTypography: fold},
	}
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

	if err := p.saveProgress(ctx, a, scored); err != nil {
		return SessionResult{}, err
	}

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

// saveProgress mémorise où reprendre (F-24) : fin de l'exercice s'il a été
// tapé en entier, sinon début du dernier mot atteint.
func (p *PracticeService) saveProgress(ctx context.Context, a *activeSession, scored []typing.Scored) error {
	reached := typing.Replay(scored).Reached
	if reached == 0 {
		return nil
	}
	pos := a.start + len(a.text)
	if reached < len(a.text) {
		pos = text.WordStart(a.full, a.start+reached)
	}
	return p.store.SetProgress(ctx, a.profileID, a.textID, pos)
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
