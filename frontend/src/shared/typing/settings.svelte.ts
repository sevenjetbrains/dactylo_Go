import type { Correction } from './engine';

export type SessionType = 'text' | 'time' | 'words';
export type LinePosition = 'top' | 'center';
export type ExerciseSplit = 'none' | 'paragraph' | 'lines';
export const FONTS = ['jetbrains', 'cascadia', 'courier'] as const;
export type TypingFont = (typeof FONTS)[number];

/** Polices à chasse fixe proposées pour le texte à copier (F-51). */
export const FONT_STACKS: Record<TypingFont, string> = {
  jetbrains: "'JetBrains Mono Variable', ui-monospace, monospace",
  cascadia: "'Cascadia Mono', Consolas, ui-monospace, monospace",
  courier: "'Courier New', Courier, monospace",
};

export interface PracticeSettings {
  fontSize: number; // F-01 : 14 à 32 px
  font: TypingFont; // F-51
  sounds: boolean; // F-51 : sons de frappe
  visibleLines: number; // F-03 : 1 à 10
  linePosition: LinePosition;
  correction: Correction; // F-04
  allowBackspace: boolean;
  showLiveBar: boolean; // F-05
  showKeyboard: boolean; // F-08 : clavier virtuel
  foldTypography: boolean;
  sessionType: SessionType; // F-06
  durationMin: number; // 1, 2, 5, 10
  wordCount: number;
  exerciseSplit: ExerciseSplit; // F-24
  exerciseSize: number; // paragraphes ou lignes par exercice
  slowThreshold: number; // F-40 : mot « lent » au-delà de N fois la vitesse moyenne
}

export const DURATIONS = [1, 2, 5, 10] as const;

export const DEFAULTS: PracticeSettings = {
  fontSize: 24,
  font: 'jetbrains',
  sounds: false,
  visibleLines: 3,
  linePosition: 'top',
  correction: 'free',
  allowBackspace: true,
  showLiveBar: true,
  showKeyboard: true,
  foldTypography: true,
  sessionType: 'text',
  durationMin: 1,
  wordCount: 50,
  exerciseSplit: 'paragraph',
  exerciseSize: 1,
  slowThreshold: 1.5,
};

// clé historique (Phase 1), conservée pour ne pas perdre les réglages
const STORAGE_KEY = 'dactylo.practice.settings';

const clamp = (n: unknown, min: number, max: number, fallback: number) =>
  typeof n === 'number' && Number.isFinite(n)
    ? Math.min(max, Math.max(min, Math.round(n)))
    : fallback;

/** Valide des réglages lus (stockage, ancienne version) en complétant par les défauts. */
export function sanitize(raw: unknown): PracticeSettings {
  const r = (raw && typeof raw === 'object' ? raw : {}) as Partial<
    Record<keyof PracticeSettings, unknown>
  >;
  const pick = <T>(v: unknown, allowed: readonly T[], d: T): T =>
    allowed.includes(v as T) ? (v as T) : d;
  return {
    fontSize: clamp(r.fontSize, 14, 32, DEFAULTS.fontSize),
    font: pick(r.font, FONTS, DEFAULTS.font),
    sounds: typeof r.sounds === 'boolean' ? r.sounds : DEFAULTS.sounds,
    visibleLines: clamp(r.visibleLines, 1, 10, DEFAULTS.visibleLines),
    linePosition: pick(r.linePosition, ['top', 'center'] as const, DEFAULTS.linePosition),
    correction: pick(r.correction, ['free', 'blocking'] as const, DEFAULTS.correction),
    allowBackspace:
      typeof r.allowBackspace === 'boolean' ? r.allowBackspace : DEFAULTS.allowBackspace,
    showLiveBar: typeof r.showLiveBar === 'boolean' ? r.showLiveBar : DEFAULTS.showLiveBar,
    showKeyboard: typeof r.showKeyboard === 'boolean' ? r.showKeyboard : DEFAULTS.showKeyboard,
    foldTypography:
      typeof r.foldTypography === 'boolean' ? r.foldTypography : DEFAULTS.foldTypography,
    sessionType: pick(r.sessionType, ['text', 'time', 'words'] as const, DEFAULTS.sessionType),
    durationMin: pick(r.durationMin, DURATIONS, DEFAULTS.durationMin as 1),
    wordCount: clamp(r.wordCount, 5, 5000, DEFAULTS.wordCount),
    exerciseSplit: pick(
      r.exerciseSplit,
      ['none', 'paragraph', 'lines'] as const,
      DEFAULTS.exerciseSplit,
    ),
    exerciseSize: clamp(r.exerciseSize, 1, 100, DEFAULTS.exerciseSize),
    slowThreshold:
      typeof r.slowThreshold === 'number' && Number.isFinite(r.slowThreshold)
        ? Math.min(3, Math.max(1.2, Math.round(r.slowThreshold * 10) / 10))
        : DEFAULTS.slowThreshold,
  };
}

function load(): PracticeSettings {
  try {
    return sanitize(JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '{}'));
  } catch {
    return { ...DEFAULTS };
  }
}

// Réglages partagés par l'écran de frappe et le panneau Réglages ;
// enregistrés dans le profil (F-51), avec un cache local.
export const typingSettings: PracticeSettings = $state(load());

$effect.root(() => {
  $effect(() => {
    const json = JSON.stringify(typingSettings);
    try {
      localStorage.setItem(STORAGE_KEY, json);
    } catch {
      // stockage indisponible : réglages non mémorisés
    }
  });
});
