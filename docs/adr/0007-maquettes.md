# ADR 0007 — Intégration des maquettes (docs/design/)

- Statut : accepté
- Date : 2026-10-28

## Contexte
Des maquettes ont été fournies pour l'accueil, le mode Copie, la
bibliothèque avec l'import PDF, le mode Dictée et sa correction
(`docs/design/*.png`). Les phases 0 à 3 étaient déjà livrées.

## Décision
- **Jetons de design** dans `frontend/src/style.css` : fond gris-bleu, cartes
  blanches arrondies, texte bleu nuit, accent bleu vif, jaune pour la
  révision, barre latérale bleu nuit dans les deux thèmes ; mode sombre avec
  ses propres valeurs. Classes partagées : `.card`, `.btn-*`, `.field`,
  `.eyebrow`, `.kbd`, `.badge`.
- **Polices embarquées** (`@fontsource-variable`) : Plus Jakarta Sans pour
  l'interface, JetBrains Mono pour la frappe. Aucun appel réseau (NF-03).
- **Shell** : logo, entrées de menu avec pastille fournie par le module
  (`ModuleDefinition.badge`, ex. mots à réviser), profil en bas de la barre.
  Le prénom se règle dans Réglages (`ProfileService`, avant F-50).
- **Accueil** : `HomeService.Overview` rassemble révision du jour, reprise de
  la copie, indicateurs (moyennes sur 10 sessions et écart avec les 10
  précédentes, série et record, semaine en cours) et sessions récentes.
- **Mode Copie** : choix rapide du type de session, bandeau à grands
  chiffres avec erreurs, curseur en bloc, raccourcis, et **clavier virtuel
  (F-08) avancé depuis la Phase 5** : touche suivante, Maj/AltGr/touche morte,
  doigt conseillé (couleurs par doigt), désactivable dans les réglages.
- **Bibliothèque** : puces d'étiquettes (les plus utilisées d'abord) et
  « Importés », lignes avec record et progression, actions dans un menu,
  panneaux à droite. Import PDF en un seul panneau mis à jour en direct :
  miniatures, plage, découpage, étiquettes, nettoyage avec compteurs
  (`CleanReport`), aperçu surligné et modifiable. Le choix des lignes de
  début et de fin (F-21) reste disponible, replié.
- **Découpage par texte** (migration 0005 : `texts.split`, `split_size`),
  prioritaire sur le réglage général.
- Les maquettes **Dictée** et **Correction de la dictée** servent de
  référence pour la Phase 4.

## Conséquences
- Écarts assumés avec les maquettes : le panneau d'import affiche aussi le
  titre et le choix des lignes (exigés par F-21) ; la carte Dictée de
  l'accueil est désactivée jusqu'à la Phase 4.
- Un mot n'est plus « lent » que s'il dépasse aussi la moyenne d'au moins
  80 ms par caractère (`MinSlowGapMs`) : une hésitation de quelques
  millisecondes ne compromet plus la maîtrise d'un mot.
