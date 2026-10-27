// Registre des modules : ajouter un module = ajouter son dossier dans
// src/modules/<nom> et une ligne ci-dessous. Aucun autre fichier ne change.
import { createRegistry } from './registry';
import home from '$modules/home/module';
import settings from '$modules/settings/module';

export const registry = createRegistry([home, settings]);
