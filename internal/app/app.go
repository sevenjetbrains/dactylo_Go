// Package app contient les services exposés à l'interface via les liaisons
// Wails : un service par module. C'est la seule couche qui importe Wails.
package app

import (
	"context"
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"dactylo/internal/core/events"
	"dactylo/internal/store"
)

// Version de l'application (remplacée au build via -ldflags).
var Version = "0.1.0-dev"

// App porte les dépendances partagées par les services et le cycle de vie
// de la fenêtre.
type App struct {
	Store *store.Store
	Bus   *events.Bus

	ctx         context.Context
	stopRelay   func()
	servicesFor []any
}

// New crée l'application et ses services, et s'assure qu'un profil existe.
func New(s *store.Store, bus *events.Bus) (*App, error) {
	profile, err := RestoreActiveProfile(context.Background(), s)
	if err != nil {
		return nil, err
	}
	a := &App{Store: s, Bus: bus}
	a.servicesFor = []any{
		NewSystemService(s),
		NewTextService(s, bus, profile, dialogPicker{a}),
		NewPracticeService(s, bus, profile),
		NewReviewService(s, bus, profile),
		NewStatsService(s, profile, dialogPicker{a}),
		NewDictationService(s, bus, profile),
		NewHomeService(s, profile),
		NewProfileService(s, profile),
		NewBackupService(s, profile, dialogPicker{a}),
		// Nouveaux services de modules : une ligne ici.
	}
	return a, nil
}

// Services renvoie les objets à lier à l'interface (options.App.Bind).
func (a *App) Services() []any {
	return a.servicesFor
}

// Startup est appelé par Wails quand la fenêtre est prête : les événements
// de domaine sont relayés à l'interface sous leur nom.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.stopRelay = a.Bus.SubscribeAll(func(e events.Event) {
		runtime.EventsEmit(a.ctx, e.Name, e.Payload)
	})
}

// dialogPicker ouvre la boîte de dialogue système de Wails.
type dialogPicker struct{ app *App }

// PickFile ouvre la boîte de dialogue système de choix de fichier (NF-11).
func (d dialogPicker) PickFile(title string, f FileFilter) (string, error) {
	if d.app.ctx == nil {
		return "", errors.New("fenêtre non prête")
	}
	return runtime.OpenFileDialog(d.app.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: []runtime.FileFilter{{DisplayName: f.Name, Pattern: f.Pattern}},
	})
}

// SaveFile ouvre la boîte de dialogue système d'enregistrement.
func (d dialogPicker) SaveFile(title, defaultName string, f FileFilter) (string, error) {
	if d.app.ctx == nil {
		return "", errors.New("fenêtre non prête")
	}
	return runtime.SaveFileDialog(d.app.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultName,
		Filters:         []runtime.FileFilter{{DisplayName: f.Name, Pattern: f.Pattern}},
	})
}

// Shutdown libère les ressources.
func (a *App) Shutdown(context.Context) {
	if a.stopRelay != nil {
		a.stopRelay()
	}
	_ = a.Store.Close()
}
