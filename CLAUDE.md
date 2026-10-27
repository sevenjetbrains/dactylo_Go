# Dactylo — règles du projet

## Référence
- Spécifications : docs/cahier-des-charges.md (exigences F-xx et NF-xx).
- Toute décision d'architecture nouvelle = un fichier dans docs/adr/.

## Pile
Wails v2, Go (internal/core + internal/app), Svelte 5 + TypeScript strict,
Vite, Tailwind + shadcn-svelte, Chart.js, SQLite via modernc.org/sqlite
(sans CGO), migrations goose embarquées, pdfjs-dist, Web Speech API,
svelte-i18n.

## Règles d'architecture
- La logique métier vit dans internal/core (Go pur, aucun import Wails).
  Jamais dans Svelte.
- L'interface n'appelle le backend que via les liaisons générées
  (frontend/wailsjs) ; ne jamais modifier ces fichiers à la main.
- Chaque fonctionnalité = un module dans frontend/src/modules/<nom>
  + un service dans internal/app/<nom>. Un module s'enregistre dans le
  registre (route, entrée de menu, réglages) ; aucun autre fichier ne
  doit changer.
- Les modules ne s'importent pas entre eux : ils communiquent par les
  événements de domaine (SessionCompleted, WordMistyped...).
- Toute modification du schéma = nouvelle migration, jamais d'édition
  d'une migration existante.
- Synthèse vocale, planificateur de révision, import de source :
  toujours derrière une interface (SpeechEngine, Scheduler, TextSource).
- Pas de dépendance qui exige CGO ou un compilateur C.

## Saisie clavier
- Utiliser beforeinput / input / compositionend, pas seulement keydown
  (touches mortes AZERTY : ê, ë...). Comparer en Unicode NFC.

## Qualité
- TDD sur internal/core. Couverture >= 80 % sur core.
- Avant chaque commit : gofmt, golangci-lint run, go test ./...,
  pnpm lint, pnpm check, pnpm test.
- Commits conventionnels, référençant l'exigence : feat(practice): F-03 ...
- Une phase à la fois (section 6.2). Ne pas commencer la suivante
  sans validation.

## Langue
- Code et noms en anglais ; interface en français via i18n ;
  commentaires et docs en français.

## Repères pratiques
- Ajouter un module frontend : dossier `frontend/src/modules/<id>/` avec
  `module.ts` (export default `ModuleDefinition`, voir `src/app/types.ts`),
  ses pages et `i18n/fr.json` + `i18n/en.json` (clés préfixées par l'id),
  puis une ligne dans `frontend/src/app/modules.ts`.
- Ajouter un service Go : `internal/app/<nom>.go`, puis une ligne dans
  `App.New` (`internal/app/app.go`) ; régénérer les liaisons avec
  `wails generate module`.
- Alias d'import : `$app`, `$modules`, `$shared`, `$i18n`, `$wails`.
- Tests d'un composant qui appelle le backend : `vi.mock('$wails/go/app/<Service>')`.
- Base de développement : `%APPDATA%\Dactylo\dactylo.db` (Windows).
- Commandes : `wails dev` (dev), `wails build` (exécutable),
  `go test ./...`, `golangci-lint run`, et dans `frontend/` :
  `pnpm lint`, `pnpm check`, `pnpm test`, `pnpm format`,
  `pnpm test:e2e` (Playwright, Edge local, liaisons simulées dans
  `frontend/e2e/mock-backend.ts`).
- Classes CSS maison : éviter les noms d'utilitaires Tailwind (`fixed`,
  `block`, `hidden`...) ; préfixer (ex. `ch-ok`).
