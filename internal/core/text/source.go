// Package text regroupe le découpage, le nettoyage et la normalisation des
// textes, ainsi que l'interface des sources de texte.
package text

import "context"

// Document est un texte prêt à être enregistré dans la bibliothèque.
type Document struct {
	Title    string
	Content  string
	Language string
	Source   string // "paste", "pdf"...
	File     string
	Pages    string
}

// TextSource produit un document à partir d'une origine (collage, PDF,
// plus tard OCR, EPUB, site web).
type TextSource interface {
	Name() string
	Load(ctx context.Context) (Document, error)
}
