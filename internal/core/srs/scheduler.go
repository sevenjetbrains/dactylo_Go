// Package srs contient la répétition espacée des mots difficiles (F-42).
// L'algorithme (SM-2, FSRS) est choisi derrière l'interface Scheduler.
package srs

import "time"

// Card est l'état de révision d'un mot.
type Card struct {
	Repetitions int       `json:"repetitions"`
	Interval    int       `json:"interval"` // jours
	Ease        float64   `json:"ease"`
	Due         time.Time `json:"due"`
}

// Grade note une révision : 0 = échec ... 5 = parfait.
type Grade int

// Scheduler calcule le prochain état d'une carte après une révision.
type Scheduler interface {
	Review(c Card, g Grade, now time.Time) Card
}
