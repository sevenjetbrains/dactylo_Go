// Profil actif (F-50) et synchronisation de ses réglages (F-51).

import { get } from 'svelte/store';
import { locale } from 'svelte-i18n';
import {
  Create,
  Delete,
  Get,
  List,
  Rename,
  SaveSettings,
  SaveSettingsFor,
  Switch,
  Update,
} from '$wails/go/app/ProfileService';
import type { store } from '$wails/go/models';
import { setLocale, type Locale } from '$i18n/index';
import { getTheme, setTheme, type Theme } from '$shared/theme';
import { DEFAULTS, sanitize, typingSettings } from '$shared/typing/settings.svelte';
import {
  DICTATION_DEFAULTS,
  dictationSettings,
  sanitizeDictation,
} from '$shared/dictation/settings.svelte';

// Profil actif (F-50) et ses réglages (F-51). Les réglages vivent dans le
// profil, en base ; le stockage local du navigateur n'est qu'un cache.

export const profile = $state({
  id: 0,
  name: '',
  keyboardLayout: 'azerty',
  loaded: false,
});

interface ProfileSettings {
  typing?: unknown;
  dictation?: unknown;
  theme?: Theme;
  locale?: Locale;
}

const SAVE_DELAY_MS = 600;
let syncing = false;
let timer: ReturnType<typeof setTimeout> | null = null;

/** Réglages courants, tels qu'ils sont enregistrés dans le profil. */
function snapshot(): ProfileSettings {
  return {
    typing: $state.snapshot(typingSettings),
    dictation: $state.snapshot(dictationSettings),
    theme: getTheme(),
    locale: (get(locale) as Locale | null | undefined) ?? 'fr',
  };
}

/** Applique les réglages enregistrés dans le profil. */
function hydrate(raw: string): void {
  let s: ProfileSettings;
  try {
    s = (JSON.parse(raw || '{}') ?? {}) as ProfileSettings;
  } catch {
    return;
  }
  if (s.typing) Object.assign(typingSettings, sanitize(s.typing));
  if (s.dictation) Object.assign(dictationSettings, sanitizeDictation(s.dictation));
  if (s.theme) setTheme(s.theme);
  if (s.locale) setLocale(s.locale);
}

/** Enregistre tout de suite les réglages dans le profil actif. */
export async function persistSettings(): Promise<void> {
  if (!profile.loaded) return;
  if (timer) clearTimeout(timer);
  timer = null;
  try {
    await SaveSettings(JSON.stringify(snapshot()));
  } catch {
    // backend indisponible : les réglages restent dans le cache local
  }
}

/** Réenregistre les réglages dans le profil après chaque modification. */
function startSync(): void {
  if (syncing) return;
  syncing = true;
  $effect.root(() => {
    let first = true;
    $effect(() => {
      // lecture profonde : toute modification des réglages déclenche l'effet
      JSON.stringify(typingSettings);
      JSON.stringify(dictationSettings);
      if (first) {
        first = false;
        return;
      }
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => void persistSettings(), SAVE_DELAY_MS);
    });
  });
}

/** Charge le profil actif et ses réglages au démarrage. */
export async function loadProfile(): Promise<void> {
  try {
    const p = await Get();
    profile.id = p.id;
    profile.name = p.name;
    profile.keyboardLayout = p.keyboardLayout;
    hydrate(p.settings);
    profile.loaded = true;
    startSync();
  } catch {
    // backend indisponible (tests) : on garde les valeurs par défaut
  }
}

/** Tous les profils locaux. */
export async function listProfiles(): Promise<store.Profile[]> {
  return List();
}

/** Change de profil : toute l'interface est rechargée avec ses données. */
export async function switchProfile(id: number): Promise<void> {
  await persistSettings();
  await Switch(id);
  window.location.reload();
}

/** Crée un profil avec les réglages par défaut. */
export async function createProfile(name: string, layout: string): Promise<store.Profile> {
  const p = await Create(name, layout);
  await SaveSettingsFor(
    p.id,
    JSON.stringify({
      typing: DEFAULTS,
      dictation: DICTATION_DEFAULTS,
      theme: 'system',
      locale: 'fr',
    }),
  );
  return p;
}

/** Change le nom et la disposition d'un profil. */
export async function updateProfile(id: number, name: string, layout: string): Promise<void> {
  const p = await Update(id, name, layout);
  if (id === profile.id) {
    profile.name = p.name;
    profile.keyboardLayout = p.keyboardLayout;
  }
}

/** Supprime un profil ; supprimer le profil actif recharge l'interface. */
export async function deleteProfile(id: number): Promise<void> {
  await Delete(id);
  if (id === profile.id) window.location.reload();
}

/** Change le prénom du profil actif. */
export async function renameProfile(name: string): Promise<void> {
  const p = await Rename(name);
  profile.name = p.name;
}

/** Nom à afficher : « Moi » (nom par défaut) ne sert pas de prénom. */
export function displayName(name: string): string {
  return name === 'Moi' ? '' : name;
}
