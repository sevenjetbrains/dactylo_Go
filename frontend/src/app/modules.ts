// Registre des modules : ajouter un module = ajouter son dossier dans
// src/modules/<nom> et une ligne ci-dessous. Aucun autre fichier ne change.
import { createRegistry } from './registry';
import home from '$modules/home/module';
import settings from '$modules/settings/module';
import practice from '$modules/practice/module';
import library from '$modules/library/module';
import review from '$modules/review/module';
import stats from '$modules/stats/module';

export const registry = createRegistry([home, practice, library, review, stats, settings]);
