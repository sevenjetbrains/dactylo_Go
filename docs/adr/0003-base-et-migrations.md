# ADR 0003 — Base SQLite et migrations

- Statut : accepté
- Date : 2026-10-27

## Décision
- Fichier : `os.UserConfigDir()/Dactylo/dactylo.db`
  (`%APPDATA%\Dactylo\dactylo.db` sous Windows).
- Pragmas : `foreign_keys(1)`, `busy_timeout(5000)`, `journal_mode(WAL)`.
  Une seule connexion ouverte (`SetMaxOpenConns(1)`) : SQLite n'a qu'un
  écrivain, et cela garde une base `:memory:` unique dans les tests.
- Migrations : `goose v3` (API `Provider`, sans état global), fichiers SQL
  embarqués dans `internal/store/migrations/`, appliquées à chaque
  ouverture. Table de suivi nommée `schema_migrations` (section 3.3) ;
  ses colonnes sont celles de goose (`version_id`, `is_applied`, `tstamp`).
- Dates stockées en texte ISO 8601 UTC ; champs extensibles en JSON
  (`settings`, `tags`, `srs_params`).

## Conséquences
- Une migration publiée n'est jamais modifiée : toute évolution passe par
  `0002_xxx.sql`, etc.
- La sauvegarde (F-52) pourra utiliser `VACUUM INTO` pour copier la base
  de façon cohérente même en mode WAL.
