package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dactylo/internal/core/text"
	"dactylo/internal/core/typing"
)

type fakePicker struct {
	path string
	err  error
}

func (f *fakePicker) PickPDF(string) (string, error) { return f.path, f.err }

func userTexts(t *testing.T, f *fixture, q TextQuery) []string {
	t.Helper()
	list, err := f.texts.ListTexts(q)
	if err != nil {
		t.Fatal(err)
	}
	var titles []string
	for _, s := range list.Texts {
		if s.Source != "builtin" {
			titles = append(titles, s.Title)
		}
	}
	return titles
}

func TestListTextsSearchFilterSort(t *testing.T) {
	f := newFixture(t)
	a, _ := f.texts.CreateText(TextInput{Title: "Bravo", Content: "Un élève studieux.", Tags: []string{"École"}})
	_, _ = f.texts.CreateText(TextInput{Title: "alpha", Content: "Le chat dort longtemps sur le canapé.", Tags: []string{"animaux", "école"}})
	_, _ = f.texts.CreateText(TextInput{Title: "Charlie", Content: "Court."})

	if got := userTexts(t, f, TextQuery{}); strings.Join(got, ",") != "Charlie,alpha,Bravo" {
		t.Errorf("récents d'abord = %v", got)
	}
	if got := userTexts(t, f, TextQuery{Search: "ELEVE"}); strings.Join(got, ",") != "Bravo" {
		t.Errorf("recherche sans accents = %v", got)
	}
	if got := userTexts(t, f, TextQuery{Tag: "ecole", Sort: "title"}); strings.Join(got, ",") != "alpha,Bravo" {
		t.Errorf("filtre étiquette + tri titre = %v", got)
	}
	if got := userTexts(t, f, TextQuery{Sort: "length"}); got[0] != "Charlie" {
		t.Errorf("tri longueur = %v", got)
	}
	if got := userTexts(t, f, TextQuery{Sort: "oldest"}); got[0] != "Bravo" {
		t.Errorf("tri ancienneté = %v", got)
	}

	// meilleur score : seul « Bravo » a été tapé
	sid, _ := f.practice.StartSession(StartRequest{TextID: a})
	_ = f.practice.RecordKeystrokes(sid, typeString(0, 0, "Un"))
	_, _ = f.practice.FinishSession(sid, 1000)
	if got := userTexts(t, f, TextQuery{Sort: "score"}); got[0] != "Bravo" {
		t.Errorf("tri score = %v", got)
	}

	list, _ := f.texts.ListTexts(TextQuery{})
	if strings.Join(list.Tags[:2], ",") != "animaux,École" && !strings.Contains(strings.Join(list.Tags, ","), "animaux") {
		t.Errorf("étiquettes = %v", list.Tags)
	}
	if list.Total != len(list.Texts) {
		t.Errorf("total = %d, textes = %d", list.Total, len(list.Texts))
	}
}

func TestUpdateAndDuplicate(t *testing.T) {
	f := newFixture(t)
	id, _ := f.texts.CreateText(TextInput{Title: "Long", Content: "un deux trois quatre", Tags: []string{"a", "A", " "}})
	if got, _ := f.texts.GetText(id); len(got.Tags) != 1 {
		t.Errorf("étiquettes dédoublonnées = %v", got.Tags)
	}
	_ = f.store.SetProgress(t.Context(), f.practice.profileID, id, 15)

	if err := f.texts.UpdateText(id, TextInput{Title: "Court", Content: "un deux"}); err != nil {
		t.Fatal(err)
	}
	if pos, _ := f.store.GetProgress(t.Context(), f.practice.profileID, id); pos != 7 {
		t.Errorf("progression ramenée à la fin du texte = %d", pos)
	}
	if err := f.texts.UpdateText(id, TextInput{Content: " "}); err == nil {
		t.Error("contenu vide accepté")
	}

	dup, err := f.texts.DuplicateText(id)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := f.texts.GetText(dup); got.Title != "Court (copie)" || got.Content != "un deux" {
		t.Errorf("copie = %+v", got)
	}
	builtin, _ := f.texts.ListTexts(TextQuery{Search: "Dormeur du val"})
	bdup, _ := f.texts.DuplicateText(builtin.Texts[0].ID)
	if got, _ := f.texts.GetText(bdup); got.Source != "paste" {
		t.Errorf("copie d'un texte intégré : source = %q", got.Source)
	}
	if _, err := f.texts.DuplicateText(999); err == nil {
		t.Error("copie d'un texte absent")
	}
	if _, err := f.texts.CreateText(TextInput{Content: "x", Source: "web"}); err == nil {
		t.Error("source inconnue acceptée")
	}
}

func TestOpenPDF(t *testing.T) {
	f := newFixture(t)
	dir := t.TempDir()
	pdf := filepath.Join(dir, "livre.pdf")
	_ = os.WriteFile(pdf, []byte("%PDF-1.4\n..."), 0o600)
	notPDF := filepath.Join(dir, "faux.pdf")
	_ = os.WriteFile(notPDF, []byte("bonjour"), 0o600)

	f.picker.path = ""
	if got, err := f.texts.OpenPDF("Importer"); err != nil || !got.Cancelled {
		t.Errorf("annulation = %+v, %v", got, err)
	}
	f.picker.path = pdf
	if got, err := f.texts.OpenPDF("Importer"); err != nil || got.Name != "livre.pdf" || len(got.Data) == 0 {
		t.Errorf("PDF = %+v, %v", got, err)
	}
	f.picker.path = notPDF
	if _, err := f.texts.OpenPDF("Importer"); err == nil {
		t.Error("fichier non PDF accepté")
	}
	f.picker.path = filepath.Join(dir, "absent.pdf")
	if _, err := f.texts.OpenPDF("Importer"); err == nil {
		t.Error("fichier absent accepté")
	}
	f.picker.err = errors.New("boom")
	if _, err := f.texts.OpenPDF("Importer"); err == nil {
		t.Error("erreur de la boîte de dialogue ignorée")
	}
}

func TestPDFImportFlow(t *testing.T) {
	f := newFixture(t)
	cleaned := f.texts.CleanImport([]text.ImportPage{
		{Number: 3, Lines: []string{"Titre courant", "Un mot coupé en fin de li-", "gne.", "3"}},
		{Number: 4, Lines: []string{"Titre courant", "Suite du texte.", "4"}},
	}, text.DefaultCleanOptions)
	if cleaned != "Un mot coupé en fin de ligne.\n\nSuite du texte." {
		t.Fatalf("nettoyage = %q", cleaned)
	}
	id, err := f.texts.CreateText(TextInput{Content: cleaned, Source: "pdf", SourceFile: "livre.pdf", Pages: "3-4"})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := f.texts.GetText(id)
	if got.Source != "pdf" || got.Pages != "3-4" || got.SourceFile != "livre.pdf" {
		t.Errorf("texte importé = %+v", got)
	}
}

func TestExercisesAndResume(t *testing.T) {
	f := newFixture(t)
	content := "Premier paragraphe ici.\n\nDeuxième paragraphe là.\n\nTroisième."
	id, _ := f.texts.CreateText(TextInput{Content: content})

	plan, err := f.texts.Exercises(id, text.SplitParagraph, 1)
	if err != nil || len(plan.Segments) != 3 || plan.Resume.Index != 0 || plan.ResumeStart != 0 {
		t.Fatalf("plan initial = %+v, %v", plan, err)
	}
	seg2 := plan.Segments[1]

	// Exercice 1 tapé en entier : on reprend à l'exercice 2.
	s1 := plan.Segments[0]
	sid, _ := f.practice.StartSession(StartRequest{TextID: id, Start: s1.Start, End: s1.End})
	_ = f.practice.RecordKeystrokes(sid, typeString(0, 0, "Premier paragraphe ici."))
	_, _ = f.practice.FinishSession(sid, 5000)
	plan, _ = f.texts.Exercises(id, text.SplitParagraph, 1)
	if plan.Resume.Index != 1 || plan.ResumeStart != seg2.Start {
		t.Errorf("après l'exercice 1 : %+v", plan)
	}

	// Exercice 2 interrompu au milieu d'un mot : reprise au début de ce mot.
	sid, _ = f.practice.StartSession(StartRequest{TextID: id, Start: seg2.Start, End: seg2.End})
	_ = f.practice.RecordKeystrokes(sid, typeString(0, 0, "Deuxième para"))
	_, _ = f.practice.FinishSession(sid, 5000)
	plan, _ = f.texts.Exercises(id, text.SplitParagraph, 1)
	if plan.Resume.Index != 1 || plan.ResumeStart != seg2.Start+len([]rune("Deuxième ")) {
		t.Errorf("reprise en cours d'exercice : %+v", plan)
	}

	// Session sans frappe : la progression ne bouge pas.
	sid, _ = f.practice.StartSession(StartRequest{TextID: id, Start: 0})
	_, _ = f.practice.FinishSession(sid, 0)
	if again, _ := f.texts.Exercises(id, text.SplitParagraph, 1); again.Position != plan.Position {
		t.Errorf("progression modifiée sans frappe : %d -> %d", plan.Position, again.Position)
	}

	if err := f.texts.ResetProgress(id); err != nil {
		t.Fatal(err)
	}
	if plan, _ = f.texts.Exercises(id, text.SplitNone, 1); plan.ResumeStart != 0 || len(plan.Segments) != 1 {
		t.Errorf("après remise à zéro : %+v", plan)
	}
	if _, err := f.texts.Exercises(999, text.SplitNone, 1); err == nil {
		t.Error("texte absent accepté")
	}
}

func TestStartSessionBoundsAndSegmentPositions(t *testing.T) {
	f := newFixture(t)
	id, _ := f.texts.CreateText(TextInput{Content: "abc def"})
	for _, r := range []StartRequest{{TextID: id, Start: -1}, {TextID: id, Start: 4, End: 3}, {TextID: id, Start: 0, End: 99}} {
		if _, err := f.practice.StartSession(r); err == nil {
			t.Errorf("bornes invalides acceptées : %+v", r)
		}
	}
	// exercice « def » : les positions sont relatives à l'exercice
	sid, err := f.practice.StartSession(StartRequest{TextID: id, Start: 4, End: 7})
	if err != nil {
		t.Fatal(err)
	}
	_ = f.practice.RecordKeystrokes(sid, typeString(0, 0, "dxf"))
	res, _ := f.practice.FinishSession(sid, 1000)
	if len(res.Mistakes) != 1 || res.Mistakes[0].Expected != "def" || res.Mistakes[0].Type != typing.ErrSubstitution {
		t.Errorf("fautes = %+v", res.Mistakes)
	}
	se, _ := f.store.GetSession(t.Context(), sid)
	if !strings.Contains(se.Settings, `"segment":{"end":7,"start":4}`) {
		t.Errorf("réglages enregistrés = %s", se.Settings)
	}
}
