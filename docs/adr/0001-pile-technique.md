# ADR 0001 — Pile technique : Wails + Go + Svelte + SQLite

- Statut : accepté
- Date : 2026-10-27

## Contexte
Application desktop hors-ligne, Windows en priorité, peu d'espace disque
disponible pour développer, installeur < 15 Mo (NF-04). Voir le comparatif
de la section 2.1 du cahier des charges.

## Décision
- Shell desktop : **Wails v2.17** (Go 1.27), WebView2 sous Windows.
- Interface : **Svelte 5** (runes) + TypeScript strict + **Vite 8**,
  **Tailwind CSS 4** (plugin Vite, jetons de couleur dans `src/style.css`).
- Données : **SQLite** via `modernc.org/sqlite` (Go pur, sans CGO).
- Traductions : `svelte-i18n`. Graphiques : `chart.js`.
- Gestionnaire de paquets frontend : **pnpm 9**.
- Qualité : golangci-lint v2, ESLint 10 (flat config) + Prettier,
  Vitest + Testing Library.

## Conséquences
- Aucun compilateur C requis ; `go build` en quelques secondes.
- Pas de version mobile directe (plan B : Tauri / Capacitor).
- `shadcn-svelte` n'est pas encore installé : il sera ajouté en Phase 1,
  au premier besoin de composants (boutons, boîtes de dialogue), via sa CLI.
- Sous Ubuntu 24.04, le build utilise le tag `webkit2_41`.
