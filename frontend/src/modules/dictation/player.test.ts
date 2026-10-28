import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { SpeechEngine } from '$shared/speech/engine';
import { DictationPlayer, type PlayerOptions } from './player.svelte';

/** Moteur factice : chaque lecture se termine quand on appelle finish(). */
class FakeEngine implements SpeechEngine {
  spoken: string[] = [];
  cancels = 0;
  private pending: (() => void) | null = null;
  available() {
    return true;
  }
  async voices() {
    return [];
  }
  speak(text: string) {
    this.spoken.push(text);
    return new Promise<void>((resolve) => (this.pending = resolve));
  }
  cancel() {
    this.cancels++;
    const p = this.pending;
    this.pending = null;
    p?.(); // une lecture annulée se résout aussi
  }
  async finish() {
    const p = this.pending;
    this.pending = null;
    p?.();
    await Promise.resolve();
    await Promise.resolve();
  }
}

const chunks = [
  { spoken: 'un deux', words: 2 },
  { spoken: 'trois quatre', words: 2 },
  { spoken: 'cinq', words: 1 },
];

let engine: FakeEngine;
let options: PlayerOptions & { mode: 'adaptive' | 'auto'; max: number };
const make = () => new DictationPlayer(chunks, engine, options, () => Date.now());

beforeEach(() => {
  vi.useFakeTimers();
  engine = new FakeEngine();
  options = {
    mode: 'auto',
    max: 0,
    speak: () => ({ rate: 1, pitch: 1, volume: 1 }),
    advance: () => options.mode,
    pauseMs: () => 1000,
    maxRepeats: () => options.max,
  };
});
afterEach(() => vi.useRealTimers());

describe('DictationPlayer', () => {
  it('avancement automatique : lit chaque segment après la pause (F-15)', async () => {
    const p = make();
    p.toggle();
    expect(p.status).toBe('speaking');
    await engine.finish();
    expect(p.status).toBe('waiting');
    vi.advanceTimersByTime(999);
    expect(engine.spoken).toHaveLength(1);
    vi.advanceTimersByTime(1);
    expect(p.index).toBe(1);
    expect(engine.spoken).toEqual(['un deux', 'trois quatre']);
    await engine.finish();
    vi.advanceTimersByTime(1000);
    await engine.finish();
    expect(p.index).toBe(2);
    expect(p.status).toBe('waiting'); // dernier segment lu : on attend la correction
    vi.advanceTimersByTime(5000);
    expect(engine.spoken).toHaveLength(3);
  });

  it('avancement adaptatif : attend que les mots du segment soient tapés', async () => {
    options.mode = 'adaptive';
    const p = make();
    p.toggle();
    await engine.finish();
    vi.advanceTimersByTime(10_000);
    expect(p.index).toBe(0); // rien tapé
    p.notifyTyped(1);
    vi.advanceTimersByTime(10_000);
    expect(p.index).toBe(0);
    p.notifyTyped(2);
    vi.advanceTimersByTime(1000);
    expect(p.index).toBe(1);
    expect(p.wordsThroughCurrent).toBe(4);
  });

  it('répéter : compte et respecte la limite (F-14)', async () => {
    options.max = 1;
    const p = make();
    expect(p.repeat()).toBe(false); // pas encore commencé
    p.toggle();
    await engine.finish();
    expect(p.repeat()).toBe(true);
    expect(engine.spoken).toEqual(['un deux', 'un deux']);
    expect(p.canRepeat).toBe(false);
    expect(p.repeat()).toBe(false);
    expect(p.totalRepeats).toBe(1);
    // nouveau segment : compteur remis à zéro
    await engine.finish();
    vi.advanceTimersByTime(1000);
    expect(p.repeats).toBe(0);
    expect(p.totalRepeats).toBe(1);
  });

  it('pause, reprise et segment précédent', async () => {
    const p = make();
    p.toggle();
    await engine.finish();
    vi.advanceTimersByTime(1000); // segment 2 en lecture
    p.toggle();
    expect(p.status).toBe('paused');
    vi.advanceTimersByTime(60_000);
    expect(p.index).toBe(1); // pas d'avancée pendant la pause
    p.toggle(); // reprise : relit le segment courant
    expect(engine.spoken.at(-1)).toBe('trois quatre');
    p.previous();
    expect(p.index).toBe(0);
    expect(engine.spoken.at(-1)).toBe('un deux');
    p.stop();
    expect(p.status).toBe('done');
    // une fin de lecture annulée n'a pas d'effet
    await engine.finish();
    expect(p.status).toBe('done');
  });

  it('temps effectif : pauses exclues', async () => {
    const p = make();
    p.toggle();
    vi.advanceTimersByTime(3000);
    p.toggle(); // pause
    vi.advanceTimersByTime(10_000);
    expect(p.elapsed()).toBe(3000);
    p.toggle();
    vi.advanceTimersByTime(2000);
    expect(p.elapsed()).toBe(5000);
  });
});
