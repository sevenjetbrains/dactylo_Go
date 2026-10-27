import { addMessages, init, locale } from 'svelte-i18n';
import type { Dictionary, ModuleDefinition } from '$app/types';
import fr from './fr.json';
import en from './en.json';

export const LOCALES = ['fr', 'en'] as const;
export type Locale = (typeof LOCALES)[number];

/** Langues écrites de droite à gauche (l'arabe est prévu, NF-08). */
const RTL_LOCALES = new Set(['ar']);

const STORAGE_KEY = 'dactylo.locale';

function storedLocale(): Locale | undefined {
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    return (LOCALES as readonly string[]).includes(v ?? '') ? (v as Locale) : undefined;
  } catch {
    return undefined;
  }
}

/** Charge les traductions communes et celles de chaque module, puis initialise la langue. */
export function setupI18n(
  modules: readonly Pick<ModuleDefinition, 'id' | 'messages'>[],
  initial?: Locale,
): void {
  addMessages('fr', fr as Dictionary);
  addMessages('en', en as Dictionary);
  for (const m of modules) {
    for (const [lang, dict] of Object.entries(m.messages ?? {})) {
      if (dict) addMessages(lang, { [m.id]: dict });
    }
  }
  // Français par défaut (F-51), quelle que soit la langue du système.
  const start = initial ?? storedLocale() ?? 'fr';
  init({ fallbackLocale: 'fr', initialLocale: start });
  applyDirection(start);
}

export function setLocale(l: Locale): void {
  locale.set(l);
  applyDirection(l);
  try {
    localStorage.setItem(STORAGE_KEY, l);
  } catch {
    // stockage indisponible : la langue ne sera pas mémorisée
  }
}

function applyDirection(l: string): void {
  document.documentElement.lang = l;
  document.documentElement.dir = RTL_LOCALES.has(l) ? 'rtl' : 'ltr';
}
