# ADR 0002 — Registre de modules et routage

- Statut : accepté
- Date : 2026-10-27

## Contexte
Ajouter une fonctionnalité doit revenir à ajouter un dossier, sans modifier
l'existant (section 3.1, critère « module factice Hello »).

## Décision
- Chaque module frontend exporte une `ModuleDefinition`
  (`frontend/src/app/types.ts`) : id, route, libellé i18n, icône, ordre,
  page, panneau de réglages optionnel, traductions.
- Le registre (`frontend/src/app/modules.ts`) est une liste explicite :
  une ligne par module. `createRegistry` refuse au démarrage les ids ou
  routes en double.
- Les traductions d'un module sont fusionnées sous son id
  (`home.title`, `settings.theme`...), donc sans collision.
- L'écran Réglages affiche les `settingsPanel` de tous les modules.
- Routage par hash (`#/settings`) avec un état `$state` : compatible avec
  le serveur d'assets Wails, sans dépendance.
- Côté Go, chaque service est créé dans `App.New` (une ligne) et lié via
  `options.App.Bind`.
- Une règle ESLint interdit l'import d'un module depuis un autre module.

## Conséquences
- Liste explicite plutôt qu'`import.meta.glob` : l'ordre de chargement et
  la présence d'un module restent lisibles dans un seul fichier.
- Le routage ne gère pas encore de paramètres (`/practice/:id`) ; à
  ajouter dans `normalizePath`/`resolve` quand le mode Copie en aura besoin.
