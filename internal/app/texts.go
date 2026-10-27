package app

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"dactylo/internal/core/events"
	"dactylo/internal/core/text"
	"dactylo/internal/store"
)

// MaxTextLength borne la taille d'un texte collé (≈ 50 000 mots).
const MaxTextLength = 300_000

// TextInput est la saisie du formulaire « Coller un texte » (F-20).
type TextInput struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Language string   `json:"language"`
	Tags     []string `json:"tags"`
}

// TextService gère la bibliothèque de textes.
type TextService struct {
	store *store.Store
	bus   *events.Bus
}

// NewTextService crée le service.
func NewTextService(s *store.Store, bus *events.Bus) *TextService {
	return &TextService{store: s, bus: bus}
}

// CreateText nettoie et enregistre un texte collé, puis émet TextImported.
func (t *TextService) CreateText(in TextInput) (int64, error) {
	content := text.CleanForPractice(in.Content)
	if content == "" {
		return 0, errors.New("le texte est vide")
	}
	if utf8.RuneCountInString(content) > MaxTextLength {
		return 0, errors.New("le texte est trop long")
	}
	title := strings.TrimSpace(text.NFC(in.Title))
	if title == "" {
		title = defaultTitle(content)
	}
	tags := make([]string, 0, len(in.Tags))
	for _, tag := range in.Tags {
		if tag = strings.TrimSpace(text.NFC(tag)); tag != "" {
			tags = append(tags, tag)
		}
	}

	id, err := t.store.CreateText(context.Background(), store.Text{
		Title: title, Content: content, Language: in.Language, Source: "paste", Tags: tags,
	})
	if err != nil {
		return 0, err
	}
	t.bus.Publish(events.Event{Name: events.TextImported, Payload: events.TextImportedPayload{TextID: id, Source: "paste"}})
	return id, nil
}

// ListTexts renvoie la bibliothèque.
func (t *TextService) ListTexts() ([]store.TextSummary, error) {
	return t.store.ListTexts(context.Background())
}

// GetText renvoie un texte complet.
func (t *TextService) GetText(id int64) (store.Text, error) {
	return t.store.GetText(context.Background(), id)
}

// DeleteText supprime un texte.
func (t *TextService) DeleteText(id int64) error {
	return t.store.DeleteText(context.Background(), id)
}

// defaultTitle reprend les premiers mots du texte.
func defaultTitle(content string) string {
	words := strings.Fields(content)
	if len(words) > 6 {
		words = words[:6]
		return strings.Join(words, " ") + "…"
	}
	return strings.Join(words, " ")
}
