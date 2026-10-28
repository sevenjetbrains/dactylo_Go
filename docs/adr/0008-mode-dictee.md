# ADR 0008 — Mode Dictée

- Statut : accepté
- Date : 2026-10-28

## Synthèse vocale
- Interface `SpeechEngine` côté interface (`frontend/src/shared/speech/`),
  implémentée par la Web Speech API : voix françaises de Windows dans
  WebView2, hors-ligne. Une voix refusée retombe sur la voix française par
  défaut ; sans voix installée, un message indique où en ajouter une.
- Une implémentation Go (Piper, espeak-ng) pourra remplacer le moteur sous
  Linux sans toucher au mode Dictée (`internal/speech.SpeechEngine` existe).

## Cœur (`internal/core/dictation`)
- **Découpage** (F-11) : mot, groupe de N mots (sans franchir une fin de
  phrase), phrase, paragraphe. Les abréviations (« M. », « Mme », « Dr. »,
  initiales) ne terminent pas une phrase et se lisent telles quelles.
- **Ponctuation dite** (F-13) : « virgule », « point », « point-virgule »,
  « deux-points », « point d'interrogation », guillemets, parenthèses,
  « à la ligne » en fin de vers ou de paragraphe.
- **Correction** (F-16) : alignement mot à mot par distance d'édition sur
  les mots. Coûts : mot proche (accent, majuscule, ponctuation) 0,4 ; mot
  voisin (lettre manquante, en trop, inversion, moins de la moitié des
  lettres différentes) 0,8 ; mot sans rapport 1,6 ; oubli ou ajout 1 ; mot
  composé tapé en deux mots (« soixante quinze ») 0,4. Types : accent,
  majuscule, lettre manquante, lettre en trop, inversion, ponctuation,
  mot oublié, mot en trop, lettre erronée.
- Précision = mots justes / mots attendus ; une faute = un mot faux,
  oublié ou en trop.

## Interface
- Lecteur (`player.svelte.ts`), machine à états testée seule : lecture,
  pause (temps exclu), répéter (Ctrl+Espace, limite par segment, F-14),
  segment précédent ; avancement automatique (pause fixe) ou adaptatif
  (segment suivant quand les mots du segment sont tapés, F-15).
- Le découpage est figé une fois la dictée commencée ; voix, vitesse et
  pause s'appliquent en direct.
- Réglages courants en `localStorage` (`shared/dictation/settings`) ;
  profils nommés en base (F-17, migration 0006 : avancement, répétitions,
  profils « Débutant lent » et « Examen »).

## Données et événements
- Session `mode = dictation` ; la copie tapée et le nombre de répétitions
  sont conservés dans `sessions.settings` (données brutes) ; une ligne de
  `session_errors` par faute. Pas de frappes (`keystrokes`) : la copie est
  un texte libre, modifiable n'importe où.
- `SessionCompleted` porte les mots alignés : le module Révision ajoute les
  fautes aux mots difficiles (« Révision prévue demain »). L'écran de
  correction peut retirer un mot (`RemoveWordByKey`) ou lancer un exercice
  sur ces mots (`BuildExerciseFor`).

## Critère d'acceptation
- Une dictée de 50 mots avec 5 fautes connues (accent, majuscule, lettre
  manquante, lettre en trop, ponctuation) produit exactement ces 5 fautes,
  bien classées (`TestFiftyWordsFiveKnownFaults`).
