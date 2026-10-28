-- Découpage en exercices propre à chaque texte (maquette « Bibliothèque et
-- import PDF » : « Découper en exercices » à l'import). NULL = réglage
-- général de l'utilisateur.

-- +goose Up
ALTER TABLE texts ADD COLUMN split TEXT CHECK (split IN ('none', 'paragraph', 'lines'));
ALTER TABLE texts ADD COLUMN split_size INTEGER;

-- +goose Down
ALTER TABLE texts DROP COLUMN split_size;
ALTER TABLE texts DROP COLUMN split;
