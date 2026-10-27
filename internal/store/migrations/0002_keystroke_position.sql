-- Position (en runes) du caractère visé par chaque frappe : nécessaire pour
-- rejouer la session et retrouver les mots fautifs (F-31). La colonne idx
-- reste le numéro de séquence de la frappe.

-- +goose Up
ALTER TABLE keystrokes ADD COLUMN position INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE keystrokes DROP COLUMN position;
