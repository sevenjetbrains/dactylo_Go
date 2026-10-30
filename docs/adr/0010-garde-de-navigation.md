# ADR 0010 — Garde de navigation (quitter une session en cours)

## Contexte

Changer de menu au milieu d'une session de Copie, de Révision ou de Dictée
quittait la page sans prévenir. La Copie restait proposée à la reprise
(NF-06), mais la Dictée perdait le texte tapé, et l'utilisateur ne savait
pas ce qu'il advenait de sa session.

## Décision

- Le routeur (`frontend/src/app/router.svelte.ts`) accepte des gardes :
  `router.guard(fn)` où `fn` renvoie le message à afficher si la navigation
  doit être confirmée, sinon `null`. Il renvoie la fonction qui retire le
  garde.
- Au changement d'adresse (lien du menu, `router.navigate`, saisie du hash),
  si un garde retient la navigation, l'ancienne adresse est remise
  (`history.replaceState`, sans nouvel événement) et la navigation reste en
  attente (`router.pending`). La coquille (`App.svelte`) affiche alors la
  confirmation : « Continuer la session » (`router.stay`) ou « Quitter »
  (`router.leave`).
- Les sessions posent leur garde : `TypingSession` (Copie et Révision) dès
  la première frappe et jusqu'à la fin ou l'abandon (Ctrl+R, Ctrl+Entrée) ;
  `DictationSession` dès la lecture ou la première frappe, jusqu'à la
  correction. Pendant la question, le chrono ou la voix sont mis en pause.
- Quitter ne change pas l'état enregistré : la Copie reste interrompue et
  proposée à la reprise depuis l'accueil.

## Conséquences

- Mécanisme générique : tout module peut poser un garde sans toucher aux
  autres fichiers.
- Le garde ne couvre pas la fermeture de la fenêtre ni le changement de
  profil (rechargement) : ces cas restent couverts par la reprise (NF-06).
- `shared/typing` importe désormais `$app/router.svelte`.
