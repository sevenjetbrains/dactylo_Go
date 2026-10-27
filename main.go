package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"dactylo/internal/app"
	"dactylo/internal/core/events"
	"dactylo/internal/store"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dbPath, err := store.DefaultPath()
	if err != nil {
		log.Fatal(err)
	}
	s, err := store.Open(context.Background(), dbPath)
	if err != nil {
		log.Fatal(err)
	}

	a := app.New(s, events.NewBus())

	err = wails.Run(&options.App{
		Title:     "Dactylo",
		Width:     1100,
		Height:    760,
		MinWidth:  720,
		MinHeight: 520,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 250, G: 250, B: 249, A: 1},
		OnStartup:        a.Startup,
		OnShutdown:       a.Shutdown,
		Bind:             a.Services(),
	})
	if err != nil {
		log.Fatal(err)
	}
}
