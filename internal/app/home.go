// Service de l'accueil : révision du jour, copie à reprendre, indicateurs,
// sessions récentes et session interrompue (NF-06).

package app

import (
	"context"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"dactylo/internal/core/srs"
	"dactylo/internal/core/stats"
	"dactylo/internal/core/text"
	"dactylo/internal/store"
)

// Accueil (maquette « Accueil ») : révision du jour, reprise de la copie,
// indicateurs et sessions récentes, en un seul appel.

// SecondsPerReviewWord sert à estimer la durée de la révision du jour.
const SecondsPerReviewWord = 25

// InterruptedMaxAge : au-delà, une session interrompue n'est plus proposée.
const InterruptedMaxAge = 7 * 24 * time.Hour

// DueCard est la carte « Révision du jour » (F-43).
type DueCard struct {
	Count   int      `json:"count"`
	Words   []string `json:"words"` // les premiers mots, pour l'aperçu
	Minutes int      `json:"minutes"`
}

// ContinueCard est la carte « Reprendre la copie ».
type ContinueCard struct {
	TextID    int64    `json:"textId"`
	Title     string   `json:"title"`
	Tags      []string `json:"tags"`
	Exercise  int      `json:"exercise"`  // numéro (à partir de 1)
	Exercises int      `json:"exercises"` // nombre total
	Excerpt   string   `json:"excerpt"`   // texte à partir de la reprise
	Progress  float64  `json:"progress"`  // 0 à 1
}

// DictationCard est la carte « Dictée » : texte proposé.
type DictationCard struct {
	TextID int64    `json:"textId"`
	Title  string   `json:"title"`
	Tags   []string `json:"tags"`
	Words  int      `json:"words"`
}

// Trend est une moyenne sur les 10 dernières sessions et son écart avec les
// 10 précédentes (Delta = 0 sans historique suffisant).
type Trend struct {
	Value    float64 `json:"value"`
	Delta    float64 `json:"delta"`
	HasDelta bool    `json:"hasDelta"`
}

// Overview est le contenu de l'accueil.
type Overview struct {
	Profile   store.Profile  `json:"profile"`
	Due       DueCard        `json:"due"`
	Continue  *ContinueCard  `json:"continue"`
	Dictation *DictationCard `json:"dictation"`
	// Session de copie interrompue par une fermeture brutale (NF-06).
	Interrupted  *store.InterruptedSession `json:"interrupted"`
	Speed        Trend                     `json:"speed"`
	Accuracy     Trend                     `json:"accuracy"`
	Streak       int                       `json:"streak"`
	BestStreak   int                       `json:"bestStreak"`
	WeekMs       int64                     `json:"weekMs"`
	WeekSessions int                       `json:"weekSessions"`
	Recent       []store.SessionPoint      `json:"recent"` // de la plus récente à la plus ancienne
}

// HomeService alimente l'accueil.
type HomeService struct {
	store   *store.Store
	profile *ActiveProfile
	now     func() time.Time
}

// NewHomeService crée le service.
func NewHomeService(s *store.Store, profile *ActiveProfile) *HomeService {
	return &HomeService{store: s, profile: profile, now: time.Now}
}

// Overview rassemble l'accueil. split et size sont le découpage général
// (réglages), pour numéroter l'exercice à reprendre.
func (h *HomeService) Overview(split string, size int) (Overview, error) {
	ctx := context.Background()
	now := h.now()
	var o Overview
	var err error
	if o.Profile, err = h.store.GetProfile(ctx, h.profile.ID()); err != nil {
		return o, err
	}

	due, err := h.store.ListWords(ctx, h.profile.ID(), store.WordFilter{Status: "due", Today: srs.Day(now)})
	if err != nil {
		return o, err
	}
	o.Due = DueCard{Count: len(due), Words: []string{}}
	for i, w := range due {
		if i == 5 {
			break
		}
		o.Due.Words = append(o.Due.Words, w.Shown)
	}
	o.Due.Minutes = int(math.Ceil(float64(len(due)*SecondsPerReviewWord) / 60))

	if o.Continue, err = h.continueCard(ctx, split, size); err != nil {
		return o, err
	}
	if x, ok, err := h.store.LastInterrupted(ctx, h.profile.ID(), now.Add(-InterruptedMaxAge)); err != nil {
		return o, err
	} else if ok {
		o.Interrupted = &x
	}
	if t, ok, err := h.store.LastTextForMode(ctx, h.profile.ID(), ModeDictation); err != nil {
		return o, err
	} else if ok {
		o.Dictation = &DictationCard{TextID: t.ID, Title: t.Title, Tags: t.Tags, Words: len(strings.Fields(t.Content))}
	}

	points, err := h.store.SessionPoints(ctx, h.profile.ID(), 20)
	if err != nil {
		return o, err
	}
	o.Speed = trend(points, func(p store.SessionPoint) float64 { return p.WPMNet })
	o.Accuracy = trend(points, func(p store.SessionPoint) float64 { return p.Accuracy })
	o.Recent = []store.SessionPoint{}
	for i := len(points) - 1; i >= 0 && len(o.Recent) < 5; i-- {
		o.Recent = append(o.Recent, points[i])
	}

	days, err := h.store.SessionDays(ctx, h.profile.ID())
	if err != nil {
		return o, err
	}
	o.Streak = stats.Streak(days, now)
	o.BestStreak = stats.BestStreak(days, now.Location())

	all, err := h.store.SessionPoints(ctx, h.profile.ID(), 0)
	if err != nil {
		return o, err
	}
	monday := stats.WeekStart(now)
	for _, p := range all {
		if t, err := time.Parse(time.RFC3339Nano, p.EndedAt); err == nil && !t.Before(monday) {
			o.WeekMs += p.DurationMs
			o.WeekSessions++
		}
	}
	return o, nil
}

// continueCard prépare la carte « Reprendre la copie » du dernier texte
// commencé, avec le découpage choisi pour ce texte.
func (h *HomeService) continueCard(ctx context.Context, split string, size int) (*ContinueCard, error) {
	t, pos, ok, err := h.store.LastProgress(ctx, h.profile.ID())
	if err != nil || !ok {
		return nil, err
	}
	if t.Split != "" {
		split, size = t.Split, t.SplitSize
	}
	segs := text.Segments(t.Content, split, size)
	seg, start := text.ResumeSegment(segs, pos)
	runes := []rune(t.Content)
	excerpt := strings.Join(strings.Fields(string(runes[start:min(len(runes), start+140)])), " ")
	if start+140 < len(runes) {
		excerpt += "…"
	}
	return &ContinueCard{
		TextID: t.ID, Title: t.Title, Tags: t.Tags,
		Exercise: seg.Index + 1, Exercises: len(segs), Excerpt: excerpt,
		Progress: float64(pos) / float64(max(1, utf8.RuneCountInString(t.Content))),
	}, nil
}

// trend : moyenne des 10 dernières sessions, écart avec les 10 précédentes.
func trend(points []store.SessionPoint, value func(store.SessionPoint) float64) Trend {
	avg := func(ps []store.SessionPoint) float64 {
		sum := 0.0
		for _, p := range ps {
			sum += value(p)
		}
		return sum / float64(len(ps))
	}
	n := len(points)
	if n == 0 {
		return Trend{}
	}
	last := points[max(0, n-10):]
	t := Trend{Value: avg(last)}
	if prev := points[:max(0, n-10)]; len(prev) > 0 {
		t.Delta, t.HasDelta = t.Value-avg(prev), true
	}
	return t
}
