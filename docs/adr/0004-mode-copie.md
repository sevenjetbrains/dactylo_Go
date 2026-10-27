# ADR 0004 — Mode Copie : correction locale, calcul officiel dans le cœur

- Statut : accepté
- Date : 2026-10-27

## Contexte
La frappe doit s'afficher en moins de 16 ms (NF-01), sans aller-retour vers
Go à chaque touche, alors que la logique métier doit vivre dans
`internal/core`.

## Décision
- **Deux comparaisons, une seule autorité.** `frontend/.../practice/engine.ts`
  colore les caractères en direct ; ses règles (`matches`) reprennent celles
  de `typing.Match`. Le cœur Go **recorrige** chaque frappe à partir du
  texte (`typing.Score`) : la correction envoyée par l'interface n'est pas
  reprise. Les mesures (MPM, précision, mots fautifs) sont calculées
  uniquement par `stats.Summarize` et `typing.MistypedWords`.
- **Frappes brutes** : `{seq, position, typed, timeMs}`, `typed = "\b"` pour
  le retour arrière, envoyées par lots toutes les 2 s. La migration 0002
  ajoute `keystrokes.position` (la colonne `idx` est le numéro de séquence).
- **Session créée à la première frappe** : ouvrir un texte sans taper ne
  crée rien. Ctrl+R ou un changement de texte laisse la session en cours
  non terminée (frappes conservées) : la reprise viendra avec NF-06.
- **Affichage** : police monospace, découpage en lignes par nombre de
  colonnes (`layout.ts`), seules les lignes visibles sont rendues.
  Les classes d'état sont préfixées `ch-` (« fixed », « block »… sont aussi
  des utilitaires Tailwind générés en lisant le code source).
- **Saisie** : zone de texte cachée, lecture de `input` / `compositionend`,
  normalisation NFC ; `beforeinput` pour Retour arrière, Entrée et le
  blocage du collage.
- **Réglages du mode Copie** en `localStorage` pour l'instant ; ils
  rejoindront `profiles.settings` avec les profils (F-50, F-51).
- **MPM net** = brut − erreurs non corrigées par minute ;
  **CPM** = caractères justes en fin de session par minute.
- Tests de bout en bout : `frontend/e2e/` (Playwright, liaisons Go
  simulées, Edge installé en local, Chromium en CI) plutôt que
  `tests/e2e/` à la racine, pour résoudre `@playwright/test` depuis
  `frontend/node_modules`.
