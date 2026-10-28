import type { SpeakOptions, SpeechEngine } from '$shared/speech/engine';
import { Clock } from '$shared/typing/clock';

// Lecteur de dictée (F-10, F-14, F-15) : lit les segments un par un, avance
// automatiquement (temps fixe) ou de façon adaptative (quand l'élève a tapé
// les mots du segment), répète, revient en arrière, se met en pause.

export interface PlayerChunk {
  spoken: string;
  words: number;
}

export interface PlayerOptions {
  speak: () => SpeakOptions;
  advance: () => 'adaptive' | 'auto';
  pauseMs: () => number;
  /** Répétitions autorisées par segment (0 = illimité). */
  maxRepeats: () => number;
}

export type PlayerStatus = 'idle' | 'speaking' | 'waiting' | 'paused' | 'done';

/** Délai minimal avant le segment suivant en mode adaptatif. */
const MIN_ADAPTIVE_MS = 300;

export class DictationPlayer {
  index = $state(0);
  status: PlayerStatus = $state('idle');
  /** Répétitions du segment en cours, et de toute la dictée. */
  repeats = $state(0);
  totalRepeats = $state(0);

  private typedWords = 0;
  private generation = 0; // ignore les fins de lecture annulées
  private timer: ReturnType<typeof setTimeout> | null = null;
  private readonly clock: Clock;

  constructor(
    private readonly chunks: readonly PlayerChunk[],
    private readonly engine: SpeechEngine,
    private readonly opts: PlayerOptions,
    now: () => number = () => performance.now(),
  ) {
    this.clock = new Clock(now);
  }

  get count(): number {
    return this.chunks.length;
  }

  /** Mots attendus jusqu'au segment courant inclus (avancement adaptatif). */
  get wordsThroughCurrent(): number {
    return this.chunks.slice(0, this.index + 1).reduce((s, c) => s + c.words, 0);
  }

  get canRepeat(): boolean {
    const max = this.opts.maxRepeats();
    return max === 0 || this.repeats < max;
  }

  /** Temps de dictée effectif (pauses exclues). */
  elapsed(): number {
    return this.clock.elapsed();
  }

  /** Lecture / pause (bouton principal). */
  toggle(): void {
    if (this.status === 'idle') {
      this.clock.start();
      this.play();
    } else if (this.status === 'paused') {
      this.clock.resume();
      this.play();
    } else if (this.status === 'speaking' || this.status === 'waiting') {
      this.halt();
      this.clock.pause();
      this.status = 'paused';
    }
  }

  /** Relit le segment courant (Ctrl+Espace). */
  repeat(): boolean {
    if (this.status === 'idle' || !this.canRepeat) return false;
    if (this.status === 'paused') this.clock.resume();
    this.repeats++;
    this.totalRepeats++;
    this.play();
    return true;
  }

  /** Revient au segment précédent et le lit. */
  previous(): void {
    if (this.status === 'idle') return;
    if (this.status === 'paused') this.clock.resume();
    this.index = Math.max(0, this.index - 1);
    this.repeats = 0;
    this.play();
  }

  /** Nombre de mots tapés par l'élève (avancement adaptatif). */
  notifyTyped(words: number): void {
    this.typedWords = words;
    if (this.status === 'waiting' && this.opts.advance() === 'adaptive') this.scheduleIfReady();
  }

  /** Arrête tout (correction). */
  stop(): void {
    this.halt();
    this.clock.pause();
    this.status = 'done';
  }

  private halt(): void {
    this.generation++;
    if (this.timer) clearTimeout(this.timer);
    this.timer = null;
    this.engine.cancel();
  }

  private play(): void {
    this.halt();
    const gen = this.generation;
    this.status = 'speaking';
    const chunk = this.chunks[this.index];
    if (!chunk) {
      this.status = 'done';
      return;
    }
    void this.engine
      .speak(chunk.spoken, this.opts.speak())
      .catch(() => {}) // voix indisponible : on avance quand même
      .then(() => {
        if (gen === this.generation) this.afterSpeak();
      });
  }

  private afterSpeak(): void {
    this.status = 'waiting';
    if (this.index >= this.chunks.length - 1) return; // dernier segment lu
    if (this.opts.advance() === 'auto') {
      this.timer = setTimeout(() => this.next(), this.opts.pauseMs());
    } else {
      this.scheduleIfReady();
    }
  }

  private scheduleIfReady(): void {
    if (this.timer || this.typedWords < this.wordsThroughCurrent) return;
    this.timer = setTimeout(() => this.next(), Math.max(MIN_ADAPTIVE_MS, this.opts.pauseMs()));
  }

  private next(): void {
    this.timer = null;
    this.index++;
    this.repeats = 0;
    this.play();
  }
}
