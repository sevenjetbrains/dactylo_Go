# ADR 0006 — Mots difficiles, révisions et statistiques

- Statut : accepté
- Date : 2026-10-28

## Mots difficiles (F-40, F-41)
- En fin de session, `typing.WordAttempts` liste chaque mot entièrement tapé :
  forme tapée au premier essai, justesse, durée (de la frappe précédant le
  mot à sa dernière frappe), ms par caractère. Un mot juste est **lent** s'il
  dépasse `slowThreshold` (1,5 par défaut, réglable de 1,2 à 3) fois la
  moyenne par caractère de la session (au moins 3 mots).
- Clé d'un mot : NFC, minuscules, sans ponctuation autour **ni élision**
  (« l'élève » → « élève », « qu'il » → « il »). Les mots de moins de deux
  lettres et les nombres ne sont pas suivis.
- Ces tentatives voyagent dans l'événement `SessionCompleted`. Le service
  Révision s'y abonne : le mode Copie ne connaît pas les révisions.
- Un mot entre dans la liste s'il est faux ou lent ; ensuite, toute session
  où il apparaît le fait avancer ou reculer, **une fois par session** (son
  pire résultat).

## Répétition espacée (F-42)
- `srs.SM2`, variante de SM-2 derrière l'interface `Scheduler` :
  faux → série remise à zéro, révision le lendemain ; lent → série inchangée,
  révision le lendemain ; juste → 1 jour, 6 jours, puis intervalle × facilité.
- **Maîtrisé** après 3 réussites consécutives (`mastery` = série / 3).
- Les échéances sont des **jours locaux** (`AAAA-MM-JJ`) : « à réviser
  aujourd'hui » = non maîtrisé et échéance ≤ aujourd'hui.

## Exercices de révision (F-44)
- `exercise.Generator` (interface `ExerciseGenerator`) : mots mélangés
  (complétés jusqu'à 15 mots), mots répétés N fois (une ligne par mot),
  phrases de la bibliothèque contenant le mot entier (le trait d'union lie,
  l'apostrophe sépare), repli sur la liste mélangée sans phrase trouvée.
  La dictée des mots arrivera avec la Phase 4.
- Une séance de révision est une session `mode = review` sans texte
  (`text_id` NULL) ; l'exercice généré est conservé dans `settings.content`.

## Statistiques (F-33 à F-35)
- Courbes MPM net et précision : **deux graphiques à un seul axe** (pas de
  double axe), série unique, infobulle au survol.
- Carte de chaleur : 50 dernières sessions, par touche AZERTY (accents et
  chiffres rattachés à leur touche), rampe séquentielle bleue à 7 pas (pas
  propres au mode sombre), touches de moins de 5 frappes en gris, tableau
  équivalent pour l'accessibilité. Bigrammes : paires de lettres justes
  consécutives, au moins 3 occurrences, les 10 plus lentes.
- Série : jours locaux consécutifs jusqu'à aujourd'hui (ou hier).
- Export CSV par la boîte de dialogue système : `;`, virgule décimale,
  UTF-8 avec BOM (ouverture directe dans un tableur français).
- Migration 0004 : `sessions.duration_ms` (temps effectif, pauses exclues).

## Frontend
- Les composants de frappe sont dans `src/shared/typing/` (espace i18n
  `typing`) : Copie et Révision les utilisent sans s'importer.
