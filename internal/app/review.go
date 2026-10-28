package app

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"time"

	"dactylo/internal/core/events"
	"dactylo/internal/core/exercise"
	"dactylo/internal/core/srs"
	"dactylo/internal/core/typing"
	"dactylo/internal/store"
)

// ReviewSummary alimente l'accueil (F-43) et l'écran Révisions.
type ReviewSummary struct {
	Due      int `json:"due"`
	Learning int `json:"learning"`
	Mastered int `json:"mastered"`
}

// ReviewExercise est un exercice de révision prêt à taper (F-44).
type ReviewExercise struct {
	Kind    string   `json:"kind"`
	Content string   `json:"content"`
	Words   []string `json:"words"`
}

// ReviewService tient la liste des mots difficiles (F-40 à F-45). Il ne
// connaît pas le mode Copie : il écoute l'événement SessionCompleted.
type ReviewService struct {
	store     *store.Store
	profileID int64
	scheduler srs.Scheduler
	generator exercise.ExerciseGenerator
	now       func() time.Time
	rng       *rand.Rand
}

// NewReviewService crée le service et l'abonne aux fins de session.
func NewReviewService(s *store.Store, bus *events.Bus, profileID int64) *ReviewService {
	r := &ReviewService{
		store: s, profileID: profileID, scheduler: srs.SM2{}, generator: exercise.Generator{},
		now: time.Now, rng: rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 7)),
	}
	bus.Subscribe(events.SessionCompleted, func(e events.Event) {
		p, ok := e.Payload.(events.SessionCompletedPayload)
		if !ok || p.ProfileID != r.profileID {
			return
		}
		if err := r.recordSession(p); err != nil {
			log.Printf("révisions : session %d non prise en compte : %v", p.SessionID, err)
		}
	})
	return r
}

// recordSession met à jour les mots de la session : un mot faux ou lent
// entre dans la liste ; un mot déjà suivi avance ou recule dans ses
// révisions. Un mot compte une seule fois par session (son pire résultat).
func (r *ReviewService) recordSession(p events.SessionCompletedPayload) error {
	ctx := context.Background()
	now := r.now()
	groups := map[string][]typing.WordAttempt{}
	var order []string
	for _, a := range p.Words {
		if _, seen := groups[a.Word]; !seen {
			order = append(order, a.Word)
		}
		groups[a.Word] = append(groups[a.Word], a)
	}
	for _, key := range order {
		attempts := groups[key]
		outcome := srs.Good
		for _, a := range attempts {
			switch {
			case !a.Correct:
				outcome = srs.Wrong
			case a.Slow && outcome == srs.Good:
				outcome = srs.Slow
			}
		}
		w, exists, err := r.store.GetWord(ctx, r.profileID, key)
		if err != nil {
			return err
		}
		if !exists {
			if outcome == srs.Good {
				continue // mot facile : pas suivi
			}
			w = store.Word{Word: key, Card: r.scheduler.New(now)}
		}
		for _, a := range attempts {
			w.Encounters++
			if !a.Correct {
				w.Errors++
				w.LastErrorAt = now.UTC().Format(time.RFC3339)
			}
			if a.MsPerChar > 0 {
				avg := a.MsPerChar
				if w.AvgMsPerChr != nil { // moyenne glissante sur les rencontres
					avg = *w.AvgMsPerChr + (a.MsPerChar-*w.AvgMsPerChr)/float64(w.Encounters)
				}
				w.AvgMsPerChr = &avg
			}
		}
		w.Shown = attempts[len(attempts)-1].Shown
		w.Card = r.scheduler.Review(w.Card, outcome, now)
		w.Mastery, w.NextReview = srs.Mastery(w.Card), w.Card.Due
		id, err := r.store.SaveWord(ctx, r.profileID, w)
		if err != nil {
			return err
		}
		for _, a := range attempts {
			if err := r.store.AddAttempt(ctx, id, p.SessionID, a, now); err != nil {
				return err
			}
		}
	}
	return nil
}

// Summary compte les mots à réviser aujourd'hui, en cours et maîtrisés.
func (r *ReviewService) Summary() (ReviewSummary, error) {
	ctx := context.Background()
	due, err := r.store.DueCount(ctx, r.profileID, srs.Day(r.now()))
	if err != nil {
		return ReviewSummary{}, err
	}
	all, err := r.store.ListWords(ctx, r.profileID, store.WordFilter{Status: "all"})
	if err != nil {
		return ReviewSummary{}, err
	}
	s := ReviewSummary{Due: due}
	for _, w := range all {
		if w.Mastery >= 1 {
			s.Mastered++
		} else {
			s.Learning++
		}
	}
	return s, nil
}

// ListWords renvoie les mots suivis : due, learning, mastered ou all.
func (r *ReviewService) ListWords(status string) ([]store.Word, error) {
	return r.store.ListWords(context.Background(), r.profileID, store.WordFilter{Status: status, Today: srs.Day(r.now())})
}

// AddWord ajoute un mot à réviser dès aujourd'hui (F-45).
func (r *ReviewService) AddWord(word string) error {
	ctx := context.Background()
	key := typing.NormalizeWord(word)
	if key == "" {
		return errors.New("ce n'est pas un mot")
	}
	w, exists, err := r.store.GetWord(ctx, r.profileID, key)
	if err != nil {
		return err
	}
	if !exists {
		w = store.Word{Word: key, Shown: typing.StripElision(typing.CoreWord(word))}
	}
	w.Card = r.scheduler.New(r.now()) // à réviser aujourd'hui, série remise à zéro
	w.Mastery, w.NextReview = 0, w.Card.Due
	_, err = r.store.SaveWord(ctx, r.profileID, w)
	return err
}

// RemoveWord retire un mot de la liste (F-45).
func (r *ReviewService) RemoveWord(id int64) error {
	return r.store.DeleteWord(context.Background(), r.profileID, id)
}

// RemoveWordByKey retire un mot désigné par son texte (écran de correction
// de la dictée : « Retirer »). Un mot absent n'est pas une erreur.
func (r *ReviewService) RemoveWordByKey(word string) error {
	ctx := context.Background()
	w, ok, err := r.store.GetWord(ctx, r.profileID, typing.NormalizeWord(word))
	if err != nil || !ok {
		return err
	}
	return r.store.DeleteWord(ctx, r.profileID, w.ID)
}

// BuildExerciseFor génère un exercice sur des mots choisis (« S'entraîner
// sur ces 5 mots »), même s'ils ne sont pas encore à réviser.
func (r *ReviewService) BuildExerciseFor(words []string, kind string, repeat int) (ReviewExercise, error) {
	shown := make([]string, 0, len(words))
	for _, w := range words {
		if s := typing.StripElision(typing.CoreWord(w)); typing.NormalizeWord(s) != "" {
			shown = append(shown, s)
		}
	}
	if len(shown) == 0 {
		return ReviewExercise{}, errors.New("aucun mot à travailler")
	}
	req := exercise.Request{Kind: kind, Words: shown, Repeat: repeat}
	if kind == exercise.KindSentences {
		texts, err := r.store.ListTexts(context.Background(), r.profileID)
		if err != nil {
			return ReviewExercise{}, err
		}
		for _, t := range texts {
			req.Library = append(req.Library, t.Content)
		}
	}
	ex := r.generator.Generate(req, r.rng)
	return ReviewExercise{Kind: ex.Kind, Content: ex.Content, Words: shown}, nil
}

// MarkMastered déclare un mot maîtrisé (F-45) ; une nouvelle faute le
// remettra en révision.
func (r *ReviewService) MarkMastered(id int64) error {
	ctx := context.Background()
	w, err := r.store.GetWordByID(ctx, r.profileID, id)
	if err != nil {
		return err
	}
	w.Card.Repetitions = srs.MasteredAfter
	w.Card.Interval = 30
	w.Card.Due = srs.Day(r.now().AddDate(0, 0, w.Card.Interval))
	w.Mastery, w.NextReview = 1, w.Card.Due
	_, err = r.store.SaveWord(ctx, r.profileID, w)
	return err
}

// BuildExercise génère un exercice sur les mots à réviser aujourd'hui, ou à
// défaut sur tous les mots en cours (F-44). kind : shuffled, repeated,
// sentences.
func (r *ReviewService) BuildExercise(kind string, repeat int) (ReviewExercise, error) {
	ctx := context.Background()
	words, err := r.ListWords("due")
	if err != nil {
		return ReviewExercise{}, err
	}
	if len(words) == 0 {
		if words, err = r.ListWords("learning"); err != nil {
			return ReviewExercise{}, err
		}
	}
	if len(words) == 0 {
		return ReviewExercise{}, errors.New("aucun mot à réviser")
	}
	shown := make([]string, len(words))
	for i, w := range words {
		shown[i] = w.Shown
	}
	req := exercise.Request{Kind: kind, Words: shown, Repeat: repeat}
	if kind == exercise.KindSentences {
		texts, err := r.store.ListTexts(ctx, r.profileID)
		if err != nil {
			return ReviewExercise{}, err
		}
		for _, t := range texts {
			req.Library = append(req.Library, t.Content)
		}
	}
	ex := r.generator.Generate(req, r.rng)
	return ReviewExercise{Kind: ex.Kind, Content: ex.Content, Words: shown}, nil
}
