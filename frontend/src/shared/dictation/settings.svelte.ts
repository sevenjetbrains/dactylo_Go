// Réglages de lecture de la dictée (F-11 à F-15), mémorisés localement ;
// les profils nommés (F-17) sont en base (DictationService).

export type Chunking = 'word' | 'group' | 'sentence' | 'paragraph';
export type Advance = 'adaptive' | 'auto';

export interface DictationSettings {
  voice: string;
  rate: number; // 0,5 à 2
  pitch: number; // 0 à 2
  volume: number; // 0 à 1
  chunking: Chunking;
  groupSize: number;
  pauseMs: number; // 0 à 10 000
  speakPunctuation: boolean;
  advance: Advance;
  maxRepeats: number; // 0 = illimité
  profileId: number; // profil choisi (0 = aucun)
}

export const DICTATION_DEFAULTS: DictationSettings = {
  voice: '',
  rate: 0.9,
  pitch: 1,
  volume: 1,
  chunking: 'group',
  groupSize: 4,
  pauseMs: 2000,
  speakPunctuation: true,
  advance: 'adaptive',
  maxRepeats: 0,
  profileId: 0,
};

const STORAGE_KEY = 'dactylo.dictation.settings';

const num = (v: unknown, min: number, max: number, d: number) =>
  typeof v === 'number' && Number.isFinite(v) ? Math.min(max, Math.max(min, v)) : d;

export function sanitizeDictation(raw: unknown): DictationSettings {
  const r = (raw && typeof raw === 'object' ? raw : {}) as Record<string, unknown>;
  const d = DICTATION_DEFAULTS;
  const pick = <T>(v: unknown, allowed: readonly T[], def: T): T =>
    allowed.includes(v as T) ? (v as T) : def;
  return {
    voice: typeof r.voice === 'string' ? r.voice : d.voice,
    rate: num(r.rate, 0.5, 2, d.rate),
    pitch: num(r.pitch, 0, 2, d.pitch),
    volume: num(r.volume, 0, 1, d.volume),
    chunking: pick(r.chunking, ['word', 'group', 'sentence', 'paragraph'] as const, d.chunking),
    groupSize: Math.round(num(r.groupSize, 1, 20, d.groupSize)),
    pauseMs: Math.round(num(r.pauseMs, 0, 10000, d.pauseMs)),
    speakPunctuation:
      typeof r.speakPunctuation === 'boolean' ? r.speakPunctuation : d.speakPunctuation,
    advance: pick(r.advance, ['adaptive', 'auto'] as const, d.advance),
    maxRepeats: Math.round(num(r.maxRepeats, 0, 10, d.maxRepeats)),
    profileId: Math.round(num(r.profileId, 0, Number.MAX_SAFE_INTEGER, 0)),
  };
}

function load(): DictationSettings {
  try {
    return sanitizeDictation(JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '{}'));
  } catch {
    return { ...DICTATION_DEFAULTS };
  }
}

export const dictationSettings: DictationSettings = $state(load());

$effect.root(() => {
  $effect(() => {
    const json = JSON.stringify(dictationSettings);
    try {
      localStorage.setItem(STORAGE_KEY, json);
    } catch {
      // stockage indisponible : réglages non mémorisés
    }
  });
});
