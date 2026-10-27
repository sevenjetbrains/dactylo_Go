// Moteur de frappe côté interface : sert uniquement à l'affichage immédiat
// (F-02). Les mesures officielles sont recalculées par le cœur Go à partir
// des frappes brutes (internal/core/typing).

export type Correction = 'free' | 'blocking';

export interface EngineOptions {
  correction: Correction;
  allowBackspace: boolean;
  foldTypography: boolean;
  /** Position de fin (exclue), pour les sessions à nombre de mots fixe. */
  end?: number;
}

/** Frappe brute, au format attendu par PracticeService.RecordKeystrokes. */
export interface Keystroke {
  seq: number;
  position: number;
  typed: string;
  timeMs: number;
}

export const BACKSPACE = '\b';

export const enum CharState {
  Pending = 0,
  Correct = 1,
  Error = 2,
  /** Juste, mais après une erreur à cette position. */
  Corrected = 3,
}

const SPACE_LIKE = new Set([' ', '\t', ' ', ' ', ' ', ' ']);
const FOLD: Record<string, string> = {
  '’': "'",
  '‘': "'",
  '‛': "'",
  ʼ: "'",
  '´': "'",
  '`': "'",
  '«': '"',
  '»': '"',
  '“': '"',
  '”': '"',
  '„': '"',
  '‟': '"',
  '–': '-',
  '—': '-',
  '−': '-',
};

/** Mêmes règles que typing.Match côté Go. */
export function matches(expected: string, typed: string, foldTypography: boolean): boolean {
  if (expected === typed) return true;
  if (expected === '\n') return typed === ' ';
  if (SPACE_LIKE.has(expected)) return SPACE_LIKE.has(typed);
  if (foldTypography) return (FOLD[expected] ?? expected) === (FOLD[typed] ?? typed);
  return false;
}

/** Positions de fin (exclues) de chaque mot. */
export function wordEnds(chars: readonly string[]): number[] {
  const ends: number[] = [];
  let inWord = false;
  chars.forEach((c, i) => {
    const blank = /\s/.test(c);
    if (blank && inWord) ends.push(i);
    inWord = !blank;
  });
  if (inWord) ends.push(chars.length);
  return ends;
}

export type TypeResult = 'correct' | 'error' | 'ignored';

export class TypingEngine {
  readonly chars: string[];
  readonly state: Uint8Array;
  readonly end: number;
  position = 0;
  keystrokes = 0;
  correctKeystrokes = 0;

  private seq = 0;
  private pending: Keystroke[] = [];
  private hadError: Uint8Array;

  constructor(
    text: string,
    private readonly opts: EngineOptions,
  ) {
    this.chars = Array.from(text.normalize('NFC'));
    this.state = new Uint8Array(this.chars.length);
    this.hadError = new Uint8Array(this.chars.length);
    this.end = Math.min(opts.end ?? this.chars.length, this.chars.length);
  }

  get done(): boolean {
    return this.position >= this.end;
  }

  /** Progression entre 0 et 1. */
  get progress(): number {
    return this.end === 0 ? 1 : this.position / this.end;
  }

  get accuracy(): number {
    return this.keystrokes === 0 ? 100 : (100 * this.correctKeystrokes) / this.keystrokes;
  }

  /** Traite un caractère tapé (une seule unité de code point, en NFC). */
  type(ch: string, timeMs: number): TypeResult {
    if (this.done) return 'ignored';
    const p = this.position;
    const ok = matches(this.chars[p]!, ch, this.opts.foldTypography);
    this.record(p, ch, timeMs);
    this.keystrokes++;

    if (ok) {
      this.correctKeystrokes++;
      this.state[p] = this.hadError[p] ? CharState.Corrected : CharState.Correct;
      this.position++;
      this.skipBlankLines();
      return 'correct';
    }
    this.hadError[p] = 1;
    this.state[p] = CharState.Error;
    if (this.opts.correction === 'free') {
      this.position++;
      this.skipBlankLines();
    }
    return 'error';
  }

  /** Retour arrière : renvoie false s'il est interdit ou impossible. */
  backspace(timeMs: number): boolean {
    if (!this.opts.allowBackspace || this.position === 0 || this.done) return false;
    // revenir aussi sur les lignes vides sautées automatiquement
    while (this.position > 1 && this.isSkippedBlank(this.position - 1)) {
      this.position--;
      this.state[this.position] = CharState.Pending;
    }
    this.position--;
    this.state[this.position] = CharState.Pending;
    this.record(this.position, BACKSPACE, timeMs);
    return true;
  }

  /** Renvoie et vide les frappes pas encore envoyées au backend. */
  take(): Keystroke[] {
    const out = this.pending;
    this.pending = [];
    return out;
  }

  private record(position: number, typed: string, timeMs: number): void {
    this.pending.push({ seq: this.seq++, position, typed, timeMs: Math.round(timeMs) });
  }

  // « \n\n » : la seconde fin de ligne (paragraphe) se saute toute seule.
  private isSkippedBlank(i: number): boolean {
    return this.chars[i] === '\n' && this.chars[i - 1] === '\n';
  }

  private skipBlankLines(): void {
    while (this.position < this.end && this.isSkippedBlank(this.position)) {
      this.state[this.position] = CharState.Correct;
      this.position++;
    }
  }
}
