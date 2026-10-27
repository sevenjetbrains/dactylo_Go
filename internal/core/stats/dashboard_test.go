package stats

import (
	"testing"
	"time"

	"dactylo/internal/core/typing"
)

func TestStreak(t *testing.T) {
	loc := time.Local
	today := time.Date(2026, 10, 27, 20, 0, 0, 0, loc)
	d := func(offset int) time.Time { return today.AddDate(0, 0, -offset) }
	cases := []struct {
		name string
		days []time.Time
		want int
	}{
		{"aucun", nil, 0},
		{"aujourd'hui seulement", []time.Time{d(0)}, 1},
		{"trois jours avec doublon", []time.Time{d(0), d(0), d(1), d(2), d(4)}, 3},
		{"pas encore aujourd'hui", []time.Time{d(1), d(2)}, 2},
		{"série cassée", []time.Time{d(2), d(3)}, 0},
	}
	for _, c := range cases {
		if got := Streak(c.days, today); got != c.want {
			t.Errorf("%s : %d, want %d", c.name, got, c.want)
		}
	}
}

func TestKeyStatsAndBigrams(t *testing.T) {
	text := []rune("quand quoi Que\nqu")
	var ks []typing.Keystroke
	ts := int64(0)
	for i, r := range text {
		step := int64(100)
		if r == 'u' { // « qu » lent : 300 ms
			step = 300
		}
		ts += step
		typed := string(r)
		if i == 2 { // « a » de quand tapé « z »
			typed = "z"
		}
		ks = append(ks, typing.Keystroke{Seq: i, Position: i, Typed: typed, TimeMs: ts})
	}
	ks = append(ks, typing.Keystroke{Seq: 99, Position: 3, Typed: typing.Backspace, TimeMs: ts + 10})
	scored := typing.Score(text, ks, typing.MatchOptions{})

	keys, bigrams := KeyStats([][]typing.Scored{scored, nil}, 3)
	byKey := map[string]KeyStat{}
	for _, k := range keys {
		byKey[k.Key] = k
	}
	if q := byKey["q"]; q.Count != 4 { // Q majuscule compté sur la touche q
		t.Errorf("q = %+v", q)
	}
	if a := byKey["a"]; a.Count != 1 || a.Errors != 1 || a.ErrorRate != 100 {
		t.Errorf("a = %+v", a)
	}
	if u := byKey["u"]; u.AvgMs != 300 {
		t.Errorf("u = %+v", u)
	}
	if _, ok := byKey["↵"]; !ok {
		t.Error("fin de ligne absente")
	}
	if byKey[" "].Count != 2 {
		t.Errorf("espace = %+v", byKey[" "])
	}
	if len(bigrams) == 0 || bigrams[0].Bigram != "qu" || bigrams[0].Count != 4 || bigrams[0].AvgMs != 300 {
		t.Errorf("bigrammes = %+v", bigrams)
	}
	for _, b := range bigrams {
		if b.Count < MinBigramCount {
			t.Errorf("bigramme trop rare classé : %+v", b)
		}
	}

	if k, b := KeyStats(nil, 5); len(k) != 0 || b == nil || len(b) != 0 {
		t.Errorf("vide = %v, %v", k, b)
	}
}
