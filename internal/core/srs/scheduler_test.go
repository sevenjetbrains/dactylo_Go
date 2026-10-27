package srs

import (
	"testing"
	"time"
)

var day0 = time.Date(2026, 10, 27, 18, 30, 0, 0, time.Local)

func TestMistypedWordIsDueTomorrowAndMasteredAfterThreeSuccesses(t *testing.T) {
	var s Scheduler = SM2{}
	c := s.New(day0)
	c = s.Review(c, Wrong, day0)
	if IsDue(c, day0) {
		t.Error("mot faux : ne doit pas être à réviser le jour même")
	}
	tomorrow := day0.AddDate(0, 0, 1)
	if !IsDue(c, tomorrow) || c.Due != "2026-10-28" {
		t.Fatalf("mot faux : à réviser le lendemain, carte = %+v", c)
	}

	when := tomorrow
	for i := 1; i <= MasteredAfter; i++ {
		if Mastered(c) {
			t.Fatalf("maîtrisé trop tôt, après %d réussites", i-1)
		}
		c = s.Review(c, Good, when)
		when = when.AddDate(0, 0, c.Interval)
	}
	if !Mastered(c) || Mastery(c) != 1 || IsDue(c, when.AddDate(1, 0, 0)) {
		t.Errorf("après 3 réussites : %+v", c)
	}
}

func TestIntervals(t *testing.T) {
	s := SM2{}
	c := s.New(day0)
	if !IsDue(c, day0) || Mastery(c) != 0 {
		t.Errorf("nouvelle carte = %+v", c)
	}
	var got []int
	for range 4 {
		c = s.Review(c, Good, day0)
		got = append(got, c.Interval)
	}
	if got[0] != 1 || got[1] != 6 || got[2] < 15 || got[3] <= got[2] {
		t.Errorf("intervalles = %v", got)
	}
}

func TestSlowKeepsStreakAndWrongResets(t *testing.T) {
	s := SM2{}
	c := s.Review(s.Review(s.New(day0), Good, day0), Good, day0)
	slow := s.Review(c, Slow, day0)
	if slow.Repetitions != 2 || slow.Interval != 1 {
		t.Errorf("lent = %+v", slow)
	}
	wrong := s.Review(c, Wrong, day0)
	if wrong.Repetitions != 0 || wrong.Ease >= c.Ease {
		t.Errorf("faux = %+v", wrong)
	}
	floor := Card{Ease: 1.3}
	if s.Review(floor, Wrong, day0).Ease != 1.3 {
		t.Error("facilité minimale 1,3")
	}
	if s.Review(Card{}, Good, day0).Ease == 0 {
		t.Error("carte sans facilité : valeur par défaut")
	}
}
