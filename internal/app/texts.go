// Service de la bibliothèque : textes, recherche, import PDF, exercices et
// boîtes de dialogue de fichiers (F-20 à F-25).

package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"dactylo/internal/core/events"
	"dactylo/internal/core/text"
	"dactylo/internal/store"
)

// MaxTextLength borne la taille d'un texte (≈ 50 000 mots).
const MaxTextLength = 300_000

// MaxPDFSize borne la taille d'un PDF importé.
const MaxPDFSize = 50 << 20

// TextInput est la saisie des formulaires « Coller un texte », « Modifier »
// et de l'import PDF (F-20, F-21).
type TextInput struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Language   string   `json:"language"`
	Tags       []string `json:"tags"`
	Source     string   `json:"source"`     // paste (défaut) ou pdf
	SourceFile string   `json:"sourceFile"` // pdf : nom du fichier
	Pages      string   `json:"pages"`      // pdf : pages importées
	Split      string   `json:"split"`      // découpage propre au texte ("" = réglage général)
	SplitSize  int      `json:"splitSize"`
}

// TextQuery filtre et trie la bibliothèque (F-23).
type TextQuery struct {
	Search string `json:"search"`
	Tag    string `json:"tag"`
	Source string `json:"source"` // paste, pdf, builtin ("" = tous)
	Sort   string `json:"sort"`   // recent (défaut), oldest, title, length, score
	// Découpage général, pour compter les exercices des textes qui n'en ont pas.
	DefaultSplit string `json:"defaultSplit"`
	DefaultSize  int    `json:"defaultSize"`
}

// TextList est le résultat d'une recherche, avec toutes les étiquettes
// existantes (pour le filtre).
type TextList struct {
	Texts []store.TextSummary `json:"texts"`
	Tags  []string            `json:"tags"`
	Total int                 `json:"total"` // nombre de textes sans filtre
}

// PDFFile est un PDF choisi par l'utilisateur ; son texte est extrait par
// l'interface (pdfjs).
type PDFFile struct {
	Cancelled bool   `json:"cancelled"`
	Name      string `json:"name"`
	Data      []byte `json:"data"` // base64 en JSON
}

// ExercisePlan découpe un texte en exercices et indique où reprendre (F-24).
type ExercisePlan struct {
	Segments    []text.Segment `json:"segments"`
	Length      int            `json:"length"`   // en caractères
	Position    int            `json:"position"` // progression mémorisée
	Resume      text.Segment   `json:"resume"`
	ResumeStart int            `json:"resumeStart"`
}

// FilePicker ouvre les boîtes de dialogue système de fichier. Les services
// ne lisent ni n'écrivent jamais un chemin venu de l'interface (NF-11).
type FilePicker interface {
	PickFile(title string, filter FileFilter) (path string, err error)              // "" si annulé
	SaveFile(title, defaultName string, filter FileFilter) (path string, err error) // "" si annulé
}

// FileFilter restreint les fichiers proposés (« PDF (*.pdf) », « *.pdf »).
type FileFilter struct {
	Name    string
	Pattern string
}

// Filtres des boîtes de dialogue.
var (
	FilterPDF    = FileFilter{Name: "PDF (*.pdf)", Pattern: "*.pdf"}
	FilterCSV    = FileFilter{Name: "CSV (*.csv)", Pattern: "*.csv"}
	FilterBackup = FileFilter{Name: "Sauvegarde Dactylo (*.db)", Pattern: "*.db"}
)

// TextService gère la bibliothèque de textes.
type TextService struct {
	store   *store.Store
	bus     *events.Bus
	profile *ActiveProfile
	picker  FilePicker
}

// NewTextService crée le service.
func NewTextService(s *store.Store, bus *events.Bus, profile *ActiveProfile, picker FilePicker) *TextService {
	return &TextService{store: s, bus: bus, profile: profile, picker: picker}
}

// CreateText nettoie et enregistre un texte, puis émet TextImported.
func (t *TextService) CreateText(in TextInput) (int64, error) {
	rec, err := toRecord(in)
	if err != nil {
		return 0, err
	}
	id, err := t.store.CreateText(context.Background(), rec)
	if err != nil {
		return 0, err
	}
	t.bus.Publish(events.Event{Name: events.TextImported, Payload: events.TextImportedPayload{TextID: id, Source: rec.Source}})
	return id, nil
}

// UpdateText modifie un texte. La progression est ramenée dans les bornes.
func (t *TextService) UpdateText(id int64, in TextInput) error {
	ctx := context.Background()
	rec, err := toRecord(in)
	if err != nil {
		return err
	}
	rec.ID = id
	if err := t.store.UpdateText(ctx, rec); err != nil {
		return err
	}
	pos, err := t.store.GetProgress(ctx, t.profile.ID(), id)
	if err != nil {
		return err
	}
	if n := utf8.RuneCountInString(rec.Content); pos > n {
		return t.store.SetProgress(ctx, t.profile.ID(), id, n)
	}
	return nil
}

// DuplicateText copie un texte et renvoie l'id de la copie.
func (t *TextService) DuplicateText(id int64) (int64, error) {
	src, err := t.store.GetText(context.Background(), id)
	if err != nil {
		return 0, err
	}
	src.Title += " (copie)"
	if src.Source == "builtin" { // une copie appartient à l'utilisateur
		src.Source = "paste"
	}
	return t.store.CreateText(context.Background(), src)
}

// ListTexts renvoie la bibliothèque filtrée et triée.
func (t *TextService) ListTexts(q TextQuery) (TextList, error) {
	all, err := t.store.ListTexts(context.Background(), t.profile.ID())
	if err != nil {
		return TextList{}, err
	}
	tagSet := map[string]string{}
	tagCount := map[string]int{}
	out := make([]store.TextSummary, 0, len(all))
	for _, s := range all {
		for _, tag := range s.Tags {
			tagSet[text.SearchKey(tag)] = tag
			tagCount[text.SearchKey(tag)]++
		}
		if q.Tag != "" && !hasTag(s.Tags, q.Tag) {
			continue
		}
		if q.Source != "" && s.Source != q.Source {
			continue
		}
		if q.Search != "" && !text.Contains(s.Title+"\n"+strings.Join(s.Tags, " ")+"\n"+s.Content, q.Search) {
			continue
		}
		split, size := s.Split, s.SplitSize
		if split == "" {
			split, size = q.DefaultSplit, q.DefaultSize
		}
		s.Exercises = len(text.Segments(s.Content, split, size))
		s.Words = len(strings.Fields(s.Content))
		out = append(out, s)
	}
	sortTexts(out, q.Sort)

	tags := make([]string, 0, len(tagSet))
	for _, tag := range tagSet {
		tags = append(tags, tag)
	}
	// les plus utilisées d'abord (puces de filtre), puis par ordre alphabétique
	sort.Slice(tags, func(i, j int) bool {
		ki, kj := text.SearchKey(tags[i]), text.SearchKey(tags[j])
		if tagCount[ki] != tagCount[kj] {
			return tagCount[ki] > tagCount[kj]
		}
		return ki < kj
	})
	return TextList{Texts: out, Tags: tags, Total: len(all)}, nil
}

// GetText renvoie un texte complet.
func (t *TextService) GetText(id int64) (store.Text, error) {
	return t.store.GetText(context.Background(), id)
}

// DeleteText supprime un texte.
func (t *TextService) DeleteText(id int64) error {
	return t.store.DeleteText(context.Background(), id)
}

// OpenPDF fait choisir un PDF dans la boîte de dialogue système et renvoie
// son contenu. dialogTitle est traduit par l'interface.
func (t *TextService) OpenPDF(dialogTitle string) (PDFFile, error) {
	path, err := t.picker.PickFile(dialogTitle, FilterPDF)
	if err != nil {
		return PDFFile{}, err
	}
	if path == "" {
		return PDFFile{Cancelled: true}, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return PDFFile{}, err
	}
	if info.Size() > MaxPDFSize {
		return PDFFile{}, fmt.Errorf("fichier trop volumineux (%d Mo au plus)", MaxPDFSize>>20)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return PDFFile{}, err
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return PDFFile{}, errors.New("ce fichier n'est pas un PDF")
	}
	return PDFFile{Name: filepath.Base(path), Data: data}, nil
}

// CleanImport nettoie les lignes extraites d'un PDF (F-22) et rend compte
// des mots recollés et des lignes retirées.
func (t *TextService) CleanImport(pages []text.ImportPage, opts text.CleanOptions) text.CleanReport {
	return text.CleanImported(pages, opts)
}

// CountExercises compte les exercices d'un texte pour un découpage donné
// (aperçu de l'import).
func (t *TextService) CountExercises(content, split string, size int) int {
	return len(text.Segments(text.CleanForPractice(content), split, size))
}

// Exercises découpe un texte en exercices et indique où reprendre.
func (t *TextService) Exercises(textID int64, split string, size int) (ExercisePlan, error) {
	ctx := context.Background()
	tx, err := t.store.GetText(ctx, textID)
	if err != nil {
		return ExercisePlan{}, err
	}
	pos, err := t.store.GetProgress(ctx, t.profile.ID(), textID)
	if err != nil {
		return ExercisePlan{}, err
	}
	if tx.Split != "" { // découpage choisi pour ce texte
		split, size = tx.Split, tx.SplitSize
	}
	segs := text.Segments(tx.Content, split, size)
	if segs == nil {
		segs = []text.Segment{}
	}
	resume, start := text.ResumeSegment(segs, pos)
	return ExercisePlan{
		Segments: segs, Length: utf8.RuneCountInString(tx.Content), Position: pos,
		Resume: resume, ResumeStart: start,
	}, nil
}

// ResetProgress repart du début du texte.
func (t *TextService) ResetProgress(textID int64) error {
	return t.store.SetProgress(context.Background(), t.profile.ID(), textID, 0)
}

// toRecord valide une saisie de l'interface et la convertit en texte à enregistrer.
func toRecord(in TextInput) (store.Text, error) {
	content := text.CleanForPractice(in.Content)
	if content == "" {
		return store.Text{}, errors.New("le texte est vide")
	}
	if utf8.RuneCountInString(content) > MaxTextLength {
		return store.Text{}, errors.New("le texte est trop long")
	}
	source := in.Source
	if source == "" {
		source = "paste"
	}
	if source != "paste" && source != "pdf" {
		return store.Text{}, fmt.Errorf("source inconnue : %q", source)
	}
	title := strings.TrimSpace(text.NFC(in.Title))
	if title == "" {
		title = defaultTitle(content)
	}
	tags := make([]string, 0, len(in.Tags))
	seen := map[string]bool{}
	for _, tag := range in.Tags {
		tag = strings.TrimSpace(text.NFC(tag))
		if k := text.SearchKey(tag); tag != "" && !seen[k] {
			seen[k] = true
			tags = append(tags, tag)
		}
	}
	switch in.Split {
	case "", text.SplitNone, text.SplitParagraph, text.SplitLines:
	default:
		return store.Text{}, fmt.Errorf("découpage inconnu : %q", in.Split)
	}
	return store.Text{
		Title: title, Content: content, Language: in.Language, Tags: tags,
		Source: source, SourceFile: strings.TrimSpace(in.SourceFile), Pages: strings.TrimSpace(in.Pages),
		Split: in.Split, SplitSize: min(max(in.SplitSize, 0), 100),
	}, nil
}

// hasTag indique si une étiquette est présente, sans tenir compte des accents.
func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if text.SearchKey(t) == text.SearchKey(want) {
			return true
		}
	}
	return false
}

// sortTexts trie la liste : récents, anciens, titre ou longueur.
func sortTexts(list []store.TextSummary, by string) {
	var less func(a, b store.TextSummary) bool
	switch by {
	case "oldest":
		less = func(a, b store.TextSummary) bool {
			return a.CreatedAt < b.CreatedAt || (a.CreatedAt == b.CreatedAt && a.ID < b.ID)
		}
	case "title":
		less = func(a, b store.TextSummary) bool { return text.SearchKey(a.Title) < text.SearchKey(b.Title) }
	case "length":
		less = func(a, b store.TextSummary) bool { return a.Length < b.Length }
	case "score":
		less = func(a, b store.TextSummary) bool { // meilleurs scores d'abord, textes jamais tapés à la fin
			switch {
			case a.BestWPM == nil:
				return false
			case b.BestWPM == nil:
				return true
			}
			return *a.BestWPM > *b.BestWPM
		}
	default:
		return // déjà du plus récent au plus ancien
	}
	sort.SliceStable(list, func(i, j int) bool { return less(list[i], list[j]) })
}

// defaultTitle reprend les premiers mots du texte.
func defaultTitle(content string) string {
	words := strings.Fields(content)
	if len(words) > 6 {
		return strings.Join(words[:6], " ") + "…"
	}
	return strings.Join(words, " ")
}
