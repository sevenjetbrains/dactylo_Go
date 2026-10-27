// Package srs contient la répétition espacée des mots difficiles (F-42).
// L'algorithme (variante de SM-2 aujourd'hui, FSRS demain) est choisi
// derrière l'interface Scheduler.
package srs

import (
	"math"
	"time"
)

// DateLayout est le format des dates de révision (jour local).
const DateLayout = "2006-01-02"

// MasteredAfter est le nombre de réussites consécutives qui rendent un mot
// « maîtrisé » (F-42).
const MasteredAfter = 3

// Card est l'état de révision d'un mot.
type Card struct {
	Repetitions int     `json:"repetitions"` // réussites consécutives
	Interval    int     `json:"interval"`    // jours
	Ease        float64 `json:"ease"`
	Due         string  `json:"due"` // AAAA-MM-JJ
}

// Outcome est le résultat d'un mot dans une session.
type Outcome int

// Résultats possibles, du pire au meilleur.
const (
	Wrong Outcome = iota // mal écrit
	Slow                 // juste mais nettement trop lent
	Good                 // juste, à vitesse normale
)

// Scheduler calcule le prochain état d'une carte après une révision.
type Scheduler interface {
	New(today time.Time) Card
	Review(c Card, o Outcome, today time.Time) Card
}

// Mastered indique un mot maîtrisé.
func Mastered(c Card) bool { return c.Repetitions >= MasteredAfter }

// Mastery est la maîtrise entre 0 et 1.
func Mastery(c Card) float64 {
	return math.Min(1, float64(c.Repetitions)/MasteredAfter)
}

// IsDue indique un mot à réviser ce jour-là (non maîtrisé, échéance passée).
func IsDue(c Card, today time.Time) bool {
	return !Mastered(c) && c.Due <= Day(today)
}

// Day renvoie la date locale d'un instant (AAAA-MM-JJ).
func Day(t time.Time) string { return t.Format(DateLayout) }

// SM2 est une variante de SM-2 :
//   - mot faux : on repart de zéro, révision le lendemain ;
//   - mot juste mais lent : la série ne progresse pas, révision le lendemain ;
//   - mot juste : 1, puis 6 jours, puis intervalle × facilité.
type SM2 struct{}

// New crée la carte d'un mot qui vient d'entrer dans la liste.
func (SM2) New(today time.Time) Card {
	return Card{Ease: 2.5, Interval: 0, Due: Day(today)}
}

// Review applique le résultat d'une session.
func (SM2) Review(c Card, o Outcome, today time.Time) Card {
	if c.Ease == 0 {
		c.Ease = 2.5
	}
	switch o {
	case Wrong:
		c.Repetitions = 0
		c.Interval = 1
		c.Ease = math.Max(1.3, c.Ease-0.2)
	case Slow:
		c.Interval = 1
	case Good:
		c.Repetitions++
		switch c.Repetitions {
		case 1:
			c.Interval = 1
		case 2:
			c.Interval = 6
		default:
			c.Interval = int(math.Round(float64(max(c.Interval, 1)) * c.Ease))
		}
		c.Ease = math.Min(3.0, c.Ease+0.1)
	}
	c.Due = Day(today.AddDate(0, 0, c.Interval))
	return c
}
