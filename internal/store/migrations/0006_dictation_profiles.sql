-- Profils de dictée (F-17) : avancement (F-15) et répétitions autorisées
-- (F-14), plus deux profils prêts à l'emploi.

-- +goose Up
ALTER TABLE dictation_profiles ADD COLUMN advance TEXT NOT NULL DEFAULT 'adaptive'
    CHECK (advance IN ('adaptive', 'auto'));
ALTER TABLE dictation_profiles ADD COLUMN max_repeats INTEGER NOT NULL DEFAULT 0; -- 0 = illimité

INSERT INTO dictation_profiles (name, rate, chunking, group_size, pause_ms, speak_punctuation, advance, max_repeats)
VALUES ('Débutant lent', 0.9, 'group', 4, 2000, 1, 'adaptive', 0),
       ('Examen', 1.0, 'sentence', 4, 3000, 1, 'auto', 2);

-- +goose Down
DELETE FROM dictation_profiles WHERE name IN ('Débutant lent', 'Examen');
ALTER TABLE dictation_profiles DROP COLUMN max_repeats;
ALTER TABLE dictation_profiles DROP COLUMN advance;
