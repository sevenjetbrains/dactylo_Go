package text

import (
	"reflect"
	"strings"
	"testing"
)

func page(n int, lines ...string) ImportPage { return ImportPage{Number: n, Lines: lines} }

func TestCleanImportedFullBook(t *testing.T) {
	pages := []ImportPage{
		page(3, "LE TOUR DU MONDE", "Il était une fois un exem-", "ple de texte qui continue", "sur plusieurs lignes du livre.", "", "3"),
		page(4, "LE TOUR DU MONDE", "Un second para\u00adgraphe commence ici et", "il se poursuit sans s’arrêter.", "- 4 -"),
		page(5, "LE TOUR DU MONDE", "Dernière page,   avec  des espaces.", "Page 5 / 12"),
	}
	report := CleanImported(pages, DefaultCleanOptions)
	got := report.Text
	want := "Il était une fois un exemple de texte qui continue sur plusieurs lignes du livre.\n\n" +
		"Un second para\u00adgraphe commence ici et il se poursuit sans s'arrêter.\n\n" +
		"Dernière page, avec des espaces."
	if got != want {
		t.Errorf("CleanImported =\n%q\nwant\n%q", got, want)
	}
	// 3 en-têtes répétés + 3 numéros de page ; un mot recollé
	if report.RemovedLines != 6 || len(report.Joined) != 1 || report.Joined[0] != "exemple" {
		t.Errorf("compte rendu = %d lignes retirées, recollés %q", report.RemovedLines, report.Joined)
	}
}

func TestJoinedWord(t *testing.T) {
	cases := [][3]string{
		{"un exem", "ple de", "exemple"},
		{"«porte", "monnaie», dit-il", "portemonnaie"},
		{"seul", "mot", "seulmot"},
	}
	for _, c := range cases {
		if got := joinedWord(c[0], c[1]); got != c[2] {
			t.Errorf("joinedWord(%q, %q) = %q, want %q", c[0], c[1], got, c[2])
		}
	}
}

func TestCleanImportedOptionsOff(t *testing.T) {
	pages := []ImportPage{
		page(1, "En-tête", "Mot coupé exem-", "ple.", "1"),
		page(2, "En-tête", "Suite l’ami.", "2"),
	}
	got := CleanImported(pages, CleanOptions{}).Text
	for _, keep := range []string{"En-tête", "exem- ple", "l’ami", "1", "2"} {
		if !strings.Contains(got, keep) {
			t.Errorf("%q absent de %q (nettoyage désactivé)", keep, got)
		}
	}
}

func TestCleanImportedKeepsLegitimateLines(t *testing.T) {
	// une page seule : pas de détection de lignes répétées ; un nombre en
	// milieu de page n'est pas un numéro de page ; un tiret suivi d'une
	// majuscule n'est pas une césure.
	pages := []ImportPage{page(1, "Titre", "Il avait", "1815", "ans, peut-", "Être pas.", "Fin")}
	report := CleanImported(pages, DefaultCleanOptions)
	got := report.Text
	if !strings.Contains(got, "1815") || !strings.Contains(got, "peut- Être") || !strings.HasPrefix(got, "Titre") {
		t.Errorf("lignes légitimes supprimées : %q", got)
	}
}

func TestCleanImportedEmpty(t *testing.T) {
	if got := CleanImported(nil, DefaultCleanOptions).Text; got != "" {
		t.Errorf("vide = %q", got)
	}
	if got := CleanImported([]ImportPage{page(1, "", "  ")}, DefaultCleanOptions).Text; got != "" {
		t.Errorf("pages blanches (PDF scanné) = %q", got)
	}
}

func TestPageNumberPattern(t *testing.T) {
	for _, s := range []string{"3", "- 12 -", "Page 5", "p. 7", "5 / 12", "12 sur 300"} {
		if !pageNumberRe.MatchString(s) {
			t.Errorf("%q devrait être un numéro de page", s)
		}
	}
	for _, s := range []string{"Chapitre 3", "1815 ans", "3 heures"} {
		if pageNumberRe.MatchString(s) {
			t.Errorf("%q n'est pas un numéro de page", s)
		}
	}
}

func TestSearch(t *testing.T) {
	if SearchKey("Élève À") != "eleve a" {
		t.Errorf("SearchKey = %q", SearchKey("Élève À"))
	}
	if !Contains("La Cigale et la Fourmi", " cigale ") || !Contains("Maître corbeau", "maitre") || Contains("chat", "chien") {
		t.Error("Contains")
	}
}

func TestSegments(t *testing.T) {
	poem := "Vers un\nVers deux\n\nVers trois\nVers quatre\n\nVers cinq"
	starts := func(segs []Segment, content ...string) []string {
		var out []string
		r := []rune(poem)
		if len(content) > 0 {
			r = []rune(content[0])
		}
		for _, s := range segs {
			out = append(out, string(r[s.Start:s.End]))
		}
		return out
	}
	if got := starts(Segments(poem, SplitParagraph, 1)); !reflect.DeepEqual(got,
		[]string{"Vers un\nVers deux", "Vers trois\nVers quatre", "Vers cinq"}) {
		t.Errorf("paragraphes = %q", got)
	}
	if got := starts(Segments(poem, SplitParagraph, 2)); len(got) != 2 || got[1] != "Vers cinq" {
		t.Errorf("2 paragraphes = %q", got)
	}
	if got := starts(Segments(poem, SplitLines, 2)); !reflect.DeepEqual(got,
		[]string{"Vers un\nVers deux", "Vers trois\nVers quatre", "Vers cinq"}) {
		t.Errorf("lignes = %q", got)
	}
	flat := "  Ligne a\nLigne b  "
	if got := starts(Segments(flat, SplitParagraph, 0), flat); !reflect.DeepEqual(got, []string{"Ligne a", "Ligne b"}) {
		t.Errorf("sans ligne vide = %q", got)
	}
	whole := Segments(poem, SplitNone, 1)
	if len(whole) != 1 || whole[0].Start != 0 || whole[0].End != len([]rune(poem)) {
		t.Errorf("texte entier = %+v", whole)
	}
	if Segments(" \n ", SplitLines, 1) != nil {
		t.Error("texte blanc : aucun exercice")
	}
	for i, s := range Segments(poem, SplitLines, 1) {
		if s.Index != i {
			t.Errorf("index %d = %d", i, s.Index)
		}
	}
}

func TestResumeSegment(t *testing.T) {
	segs := []Segment{{0, 0, 10}, {1, 12, 20}, {2, 22, 30}}
	cases := []struct{ pos, seg, start int }{
		{0, 0, 0}, {5, 0, 5}, {10, 1, 12}, {15, 1, 15}, {29, 2, 29}, {30, 0, 0}, {99, 0, 0},
	}
	for _, c := range cases {
		s, start := ResumeSegment(segs, c.pos)
		if s.Index != c.seg || start != c.start {
			t.Errorf("ResumeSegment(%d) = %d, %d ; want %d, %d", c.pos, s.Index, start, c.seg, c.start)
		}
	}
	if s, start := ResumeSegment(nil, 3); s != (Segment{}) || start != 0 {
		t.Error("sans exercice")
	}
}

func TestWordStart(t *testing.T) {
	r := []rune("le chat noir")
	for pos, want := range map[int]int{0: 0, 2: 0, 3: 3, 5: 3, 8: 8, 12: 8, 99: 8, -1: 0} {
		if got := WordStart(r, pos); got != want {
			t.Errorf("WordStart(%d) = %d, want %d", pos, got, want)
		}
	}
}
