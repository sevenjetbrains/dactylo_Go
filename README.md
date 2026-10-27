# Dactylo

Application desktop hors-ligne d'entraînement à la dactylographie et à
l'orthographe en français (Wails + Go + Svelte + SQLite).

Spécifications : [docs/cahier-des-charges.md](docs/cahier-des-charges.md) ·
Règles du projet : [CLAUDE.md](CLAUDE.md) · Décisions : [docs/adr/](docs/adr/)

## Prérequis

- Go 1.25+ (`winget install GoLang.Go`)
- Node.js 22+ et pnpm 9 (`npm install -g pnpm@9`)
- Wails CLI : `go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0`
- Windows : WebView2 (déjà présent sur Windows 10/11). Vérifier avec `wails doctor`.

## Développement

```sh
wails dev              # application avec rechargement à chaud
wails build            # exécutable dans build/bin/
wails build -nsis      # installeur Windows (nécessite NSIS)
```

Vérifications avant commit :

```sh
gofmt -l . ; golangci-lint run ; go test ./...
cd frontend ; pnpm lint ; pnpm check ; pnpm test
```

## Structure

| Dossier | Rôle |
| --- | --- |
| `internal/core/` | Logique métier en Go pur (sans Wails), testée en TDD |
| `internal/app/` | Services exposés à l'interface (un par module) |
| `internal/store/` | SQLite et migrations embarquées |
| `internal/speech/` | Interface `SpeechEngine` |
| `frontend/src/app/` | Shell, routage, registre des modules |
| `frontend/src/modules/` | Un dossier par fonctionnalité |
| `frontend/wailsjs/` | Liaisons générées par Wails (ne pas modifier) |
