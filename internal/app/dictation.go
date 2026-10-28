package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"dactylo/internal/core/dictation"
	"dactylo/internal/core/events"
	"dactylo/internal/core/stats"
	"dactylo/internal/core/typing"
	"dactylo/internal/store"
)

// ModeDictation est le mode des sessions de dictée.
const ModeDictation = "dictation"

// PrepareRequest décrit la lecture voulue (F-11, F-13).
type PrepareRequest struct {
	TextID           int64  `json:"textId"`
	Start            int    `json:"start"`
	End              int    `json:"end"` // 0 : jusqu'à la fin
	Chunking         string `json:"chunking"`
	GroupSize        int    `json:"groupSize"`
	SpeakPunctuation bool   `json:"speakPunctuation"`
}

// DictationPlan est le texte découpé, prêt à lire.
type DictationPlan struct {
	TextID int64             `json:"textId"`
	Title  string            `json:"title"`
	Start  int               `json:"start"`
	End    int               `json:"end"`
	Words  int               `json:"words"`
	Chunks []dictation.Chunk `json:"chunks"`
}

// FinishDictation est la copie rendue.
type FinishDictation struct {
	SessionID   int64  `json:"sessionId"`
	Typed       string `json:"typed"`
	DurationMs  int64  `json:"durationMs"`
	Repetitions int    `json:"repetitions"`
}

// DictationResult est l'écran de correction (F-16).
type DictationResult struct {
	SessionID   int64                   `json:"sessionId"`
	TextID      int64                   `json:"textId"`
	Report      dictation.Report        `json:"report"`
	WPM         float64                 `json:"wpm"`
	DurationMs  int64                   `json:"durationMs"`
	Repetitions int                     `json:"repetitions"`
	Mistakes    []typing.WordError      `json:"mistakes"` // mots ajoutés aux révisions
	Average     store.DictationAverages `json:"average"`
}

type activeDictation struct {
	textID   int64
	expected string
	settings map[string]any
}

// DictationService gère le mode Dictée : préparation, sessions, profils.
type DictationService struct {
	store   *store.Store
	bus     *events.Bus
	profile *ActiveProfile
	now     func() time.Time

	mu     sync.Mutex
	active map[int64]*activeDictation
}

// NewDictationService crée le service.
func NewDictationService(s *store.Store, bus *events.Bus, profile *ActiveProfile) *DictationService {
	return &DictationService{store: s, bus: bus, profile: profile, now: time.Now, active: map[int64]*activeDictation{}}
}

func (d *DictationService) excerpt(ctx context.Context, textID int64, start, end int) (store.Text, string, int, int, error) {
	t, err := d.store.GetText(ctx, textID)
	if err != nil {
		return t, "", 0, 0, fmt.Errorf("texte %d : %w", textID, err)
	}
	runes := []rune(t.Content)
	if end == 0 {
		end = len(runes)
	}
	if start < 0 || start >= end || end > len(runes) {
		return t, "", 0, 0, fmt.Errorf("extrait hors du texte : [%d, %d)", start, end)
	}
	return t, string(runes[start:end]), start, end, nil
}

// Prepare découpe l'extrait en segments de lecture.
func (d *DictationService) Prepare(req PrepareRequest) (DictationPlan, error) {
	t, content, start, end, err := d.excerpt(context.Background(), req.TextID, req.Start, req.End)
	if err != nil {
		return DictationPlan{}, err
	}
	chunks := dictation.Chunks(content, req.Chunking, req.GroupSize, req.SpeakPunctuation)
	if chunks == nil {
		chunks = []dictation.Chunk{}
	}
	return DictationPlan{
		TextID: t.ID, Title: t.Title, Start: start, End: end,
		Words: dictation.CountWords(content), Chunks: chunks,
	}, nil
}

// Start ouvre une session de dictée sur l'extrait ; settings est le JSON des
// réglages de lecture utilisés.
func (d *DictationService) Start(req StartRequest) (int64, error) {
	ctx := context.Background()
	_, content, start, end, err := d.excerpt(ctx, req.TextID, req.Start, req.End)
	if err != nil {
		return 0, err
	}
	settings := map[string]any{}
	if req.Settings != "" {
		if err := json.Unmarshal([]byte(req.Settings), &settings); err != nil {
			return 0, fmt.Errorf("réglages invalides : %w", err)
		}
	}
	settings["segment"] = map[string]int{"start": start, "end": end}
	raw, err := json.Marshal(settings)
	if err != nil {
		return 0, err
	}
	id, err := d.store.CreateSession(ctx, d.profile.ID(), req.TextID, ModeDictation, string(raw), d.now())
	if err != nil {
		return 0, err
	}
	d.mu.Lock()
	d.active[id] = &activeDictation{textID: req.TextID, expected: content, settings: settings}
	d.mu.Unlock()
	return id, nil
}

// Finish corrige la copie, enregistre la session et émet SessionCompleted
// (les fautes rejoignent les révisions) puis un WordMistyped par faute.
func (d *DictationService) Finish(req FinishDictation) (DictationResult, error) {
	ctx := context.Background()
	d.mu.Lock()
	a, ok := d.active[req.SessionID]
	d.mu.Unlock()
	if !ok {
		return DictationResult{}, fmt.Errorf("dictée %d inconnue ou terminée", req.SessionID)
	}
	if req.DurationMs < 0 || req.Repetitions < 0 {
		return DictationResult{}, errors.New("durée ou répétitions négatives")
	}

	report := dictation.Correct(a.expected, req.Typed)
	wpm := 0.0
	if req.DurationMs > 0 {
		wpm = float64(utf8.RuneCountInString(strings.TrimSpace(req.Typed))) / stats.CharsPerWord / (float64(req.DurationMs) / 60000)
	}
	sum := stats.Summary{DurationMs: req.DurationMs, WPMGross: wpm, WPMNet: wpm, Accuracy: report.Accuracy,
		CPM: wpm * stats.CharsPerWord}
	mistakes := []typing.WordError{}
	for _, o := range report.Ops {
		if !o.Correct {
			mistakes = append(mistakes, typing.WordError{Expected: o.Expected, Typed: o.Typed, Position: o.Position, Type: o.Type})
		}
	}

	if err := d.store.FinishSession(ctx, req.SessionID, d.now(), sum, mistakes); err != nil {
		return DictationResult{}, err
	}
	d.mu.Lock()
	delete(d.active, req.SessionID)
	d.mu.Unlock()

	a.settings["typed"] = req.Typed // données brutes : la copie reste relisible
	a.settings["repetitions"] = req.Repetitions
	if raw, err := json.Marshal(a.settings); err == nil {
		if err := d.store.SetSessionSettings(ctx, req.SessionID, string(raw)); err != nil {
			return DictationResult{}, err
		}
	}

	avg, err := d.store.RecentDictationAverages(ctx, d.profile.ID(), req.SessionID, AverageWindow)
	if err != nil {
		return DictationResult{}, err
	}

	d.bus.Publish(events.Event{Name: events.SessionCompleted, Payload: events.SessionCompletedPayload{
		SessionID: req.SessionID, ProfileID: d.profile.ID(), TextID: a.textID, Mode: ModeDictation, Summary: sum,
		Words: dictation.Attempts(report.Ops),
	}})
	for _, m := range mistakes {
		d.bus.Publish(events.Event{Name: events.WordMistyped, Payload: events.WordMistypedPayload{
			SessionID: req.SessionID, ProfileID: d.profile.ID(), Error: m,
		}})
	}

	return DictationResult{
		SessionID: req.SessionID, TextID: a.textID, Report: report, WPM: wpm,
		DurationMs: req.DurationMs, Repetitions: req.Repetitions, Mistakes: mistakes, Average: avg,
	}, nil
}

// ListProfiles renvoie les profils de dictée (F-17).
func (d *DictationService) ListProfiles() ([]store.DictationProfile, error) {
	return d.store.ListDictationProfiles(context.Background())
}

// SaveProfile enregistre un profil (ID = 0 : nouveau) et renvoie son id.
func (d *DictationService) SaveProfile(p store.DictationProfile) (int64, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return 0, errors.New("le profil doit avoir un nom")
	}
	switch p.Chunking {
	case dictation.ChunkWord, dictation.ChunkGroup, dictation.ChunkSentence, dictation.ChunkParagraph:
	default:
		return 0, fmt.Errorf("découpage inconnu : %q", p.Chunking)
	}
	if p.Advance != "adaptive" && p.Advance != "auto" {
		return 0, fmt.Errorf("avancement inconnu : %q", p.Advance)
	}
	clamp := func(v, lo, hi float64) float64 { return math.Min(hi, math.Max(lo, v)) }
	p.Rate = clamp(p.Rate, 0.5, 2)
	p.Pitch = clamp(p.Pitch, 0, 2)
	p.Volume = clamp(p.Volume, 0, 1)
	p.PauseMs = int(clamp(float64(p.PauseMs), 0, 10000))
	p.GroupSize = int(clamp(float64(p.GroupSize), 1, 20))
	p.MaxRepeats = int(clamp(float64(p.MaxRepeats), 0, 10))
	return d.store.SaveDictationProfile(context.Background(), p)
}

// DeleteProfile supprime un profil.
func (d *DictationService) DeleteProfile(id int64) error {
	return d.store.DeleteDictationProfile(context.Background(), id)
}
