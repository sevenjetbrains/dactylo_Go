# ADR 0005 — Import PDF, bibliothèque et exercices

- Statut : accepté
- Date : 2026-10-27

## Import PDF (F-21, F-22)
- **Lecture du fichier par Go, extraction par l'interface.** `TextService.OpenPDF`
  ouvre la boîte de dialogue système et lit le fichier choisi (NF-11 : aucun
  chemin ne vient de l'interface ; 50 Mo au plus, en-tête `%PDF-` vérifié).
  Les octets passent en base64 ; `pdfjs-dist` (worker embarqué, hors-ligne)
  fait l'aperçu des pages et l'extraction des lignes (`pdfLines.ts`).
- **Nettoyage dans le cœur** : `text.CleanImported` reçoit les lignes
  regroupées par page (nécessaire pour repérer les en-têtes répétés) et
  produit des paragraphes séparés par une ligne vide. Règles : numéros de
  page et lignes répétées cherchés seulement dans les 2 premières et
  dernières lignes de chaque page ; césure recollée si la ligne finit par
  « lettre- » et la suivante commence par une minuscule ; fin de paragraphe
  sur ligne vide après une phrase terminée, ou ligne courte (< 75 % du
  80e centile) finissant par une ponctuation forte.
- PDF sans texte : message « probablement scanné », OCR plus tard.

## Bibliothèque (F-23)
- Recherche, filtre et tri faits en Go (`TextService.ListTexts`), insensibles
  à la casse et aux accents (`text.SearchKey`). Le contenu complet sert à la
  recherche mais n'est pas envoyé à l'interface (`json:"-"`).
- Une copie d'un texte intégré devient un texte de l'utilisateur (`paste`).

## Exercices et reprise (F-24)
- `text.Segments` : texte entier, N paragraphes (blocs séparés par une ligne
  vide ; sans ligne vide, chaque ligne est un paragraphe) ou N lignes.
- Une session porte sur un intervalle `[start, end)` du texte
  (`PracticeService.StartSession(StartRequest)`) ; les positions des frappes
  sont relatives à l'exercice, les bornes sont enregistrées dans
  `sessions.settings.segment`.
- Fin de session : `text_progress.position` = fin de l'exercice s'il est
  complet, sinon début du dernier mot atteint. `ResumeSegment` en déduit
  l'exercice et la position de reprise ; un texte terminé reprend au début.

## Pack intégré (F-25)
- Migration `0003_builtin_texts.sql` : insertion unique (un texte supprimé
  ne revient pas). Œuvres du domaine public transcrites de mémoire : à relire
  avec une édition de référence.
