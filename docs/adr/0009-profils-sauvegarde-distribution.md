# ADR 0009 — Profils, sauvegarde, reprise et distribution (Phase 5)

- Statut : accepté
- Date : 2026-10-28

## Profils (F-50) et réglages (F-51)
- Un `ActiveProfile` (identifiant atomique) est partagé par tous les
  services Go : changer de profil change les données de tous les modules
  sans redémarrer le backend. Le dernier profil utilisé est mémorisé dans
  la table `app_settings` (migration 0007) et retrouvé au démarrage ; s'il
  n'existe plus, on prend le premier.
- Supprimer un profil supprime toutes ses données (cascade) ; le dernier
  profil ne peut pas être supprimé.
- Les réglages de l'interface (frappe, dictée, thème, langue) sont un JSON
  dans `profiles.settings`. L'interface les applique au chargement, puis
  les réenregistre 600 ms après chaque modification. Le stockage local du
  navigateur n'est plus qu'un cache. Changer de profil recharge la page :
  aucun état d'un profil ne fuit dans l'autre.
- Nouveaux réglages : police à chasse fixe du texte (JetBrains Mono,
  Cascadia Mono, Courier New — la mise en page suppose une chasse fixe) et
  sons de frappe (Web Audio synthétisé, sans fichier, désactivés par défaut).

## Clavier virtuel (F-08)
- Dispositions AZERTY, QWERTY et BÉPO décrites dans
  `shared/keyboard/layouts.ts` (caractères de chaque niveau, doigt par
  colonne). La disposition est une propriété du profil ; le clavier virtuel
  et la carte de chaleur des statistiques l'utilisent. Les accents sans
  touche directe passent par la touche morte (^ ou ¨) de la disposition.

## Sauvegarde (F-52)
- Export : `VACUUM INTO` vers le fichier choisi dans la boîte de dialogue
  système (NF-11) — copie cohérente même base ouverte.
- Restauration : le fichier est d'abord vérifié (base SQLite lisible,
  schéma pas plus récent que l'application), la base actuelle est copiée
  en `.bak`, puis remplacée et rouverte (migrations appliquées). En cas
  d'échec, la copie `.bak` est remise en place.

## Reprise après fermeture brutale (NF-06)
- Les frappes partent par lots toutes les 2 s : une session non terminée
  et non abandonnée, avec des frappes, de moins de 7 jours, est proposée à
  l'accueil. « Reprendre » recharge le texte, les réglages de la session et
  rejoue les frappes ; le chronomètre repart du temps de la dernière
  frappe, en pause. Recommencer (Ctrl+R), passer au texte suivant ou
  « Ignorer » marque la session abandonnée (`sessions.abandoned_at`).

## Distribution
- Installateur Windows NSIS (`wails build -nsis`) : 9 Mo, sous la limite
  de 15 Mo (NF-04), vérifiée dans le flux `release.yml` (étiquette `v*`).
  macOS : application zippée ; Linux : binaire.
- **Signature** : préparée dans la CI, appliquée seulement si les secrets
  `WINDOWS_CERT_PFX_BASE64` et `WINDOWS_CERT_PASSWORD` existent. Un
  certificat de signature de code (et un compte Apple Developer pour la
  notarisation macOS) doit être acquis par le porteur du projet.
- **Mise à jour automatique : non retenue en V1.** Elle exige d'interroger
  un serveur, ce que la V1 exclut (NF-11 : aucun accès réseau ; NF-12 :
  rien envoyé à l'extérieur). Alternative : la version est affichée sur
  l'accueil ; les nouvelles versions sont publiées sur la page des versions
  GitHub et s'installent par-dessus l'ancienne (les données, dans
  `%APPDATA%\Dactylo`, sont conservées et migrées au démarrage). Une
  vérification de mise à jour optionnelle, désactivée par défaut, pourra
  être ajoutée si l'exigence NF-11 est assouplie.
