import { describe, expect, it, vi } from 'vitest';
import { WebSpeechEngine } from './engine';

type Listener = () => void;

function fakeSynth(voices: Partial<SpeechSynthesisVoice>[]) {
  const spoken: SpeechSynthesisUtterance[] = [];
  const listeners: Listener[] = [];
  const synth = {
    getVoices: vi.fn(() => voices as SpeechSynthesisVoice[]),
    addEventListener: (_: string, l: Listener) => listeners.push(l),
    speak: (u: SpeechSynthesisUtterance) => {
      spoken.push(u);
      queueMicrotask(() => u.onend?.(new Event('end') as SpeechSynthesisEvent));
    },
    cancel: vi.fn(),
  };
  return { synth: synth as unknown as SpeechSynthesis, spoken, listeners };
}

class FakeUtterance {
  voice: SpeechSynthesisVoice | null = null;
  lang = '';
  rate = 1;
  pitch = 1;
  volume = 1;
  onend: ((e: SpeechSynthesisEvent) => void) | null = null;
  onerror: ((e: SpeechSynthesisErrorEvent) => void) | null = null;
  constructor(public text: string) {}
}
vi.stubGlobal('SpeechSynthesisUtterance', FakeUtterance);

describe('WebSpeechEngine', () => {
  const voices = [
    { voiceURI: 'en', name: 'Zira', lang: 'en-US', localService: true },
    { voiceURI: 'fr-paul', name: 'Paul', lang: 'fr-FR', localService: true },
    { voiceURI: 'fr-denise', name: 'Denise', lang: 'fr-FR', localService: false },
  ];

  it('voix françaises d’abord', async () => {
    const { synth } = fakeSynth(voices);
    const list = await new WebSpeechEngine(synth).voices();
    expect(list.map((v) => v.name)).toEqual(['Denise', 'Paul', 'Zira']);
  });

  it('lit avec la voix et les réglages choisis', async () => {
    const { synth, spoken } = fakeSynth(voices);
    const engine = new WebSpeechEngine(synth);
    await engine.speak('Bonjour', { voiceId: 'fr-paul', rate: 0.9, pitch: 1.1, volume: 0.5 });
    const u = spoken[0]!;
    expect(u.text).toBe('Bonjour');
    expect(u.voice?.name).toBe('Paul');
    expect([u.rate, u.pitch, u.volume, u.lang]).toEqual([0.9, 1.1, 0.5, 'fr-FR']);
    engine.cancel();
    expect(synth.cancel).toHaveBeenCalled();
  });

  it('une lecture interrompue se résout, une erreur réelle est rejetée', async () => {
    const { synth } = fakeSynth(voices);
    synth.speak = (u: SpeechSynthesisUtterance) =>
      queueMicrotask(() =>
        u.onerror?.({
          error: u.text === 'stop' ? 'interrupted' : 'synthesis-failed',
        } as SpeechSynthesisErrorEvent),
      );
    const engine = new WebSpeechEngine(synth);
    await expect(engine.speak('stop', { rate: 1, pitch: 1, volume: 1 })).resolves.toBeUndefined();
    await expect(engine.speak('x', { rate: 1, pitch: 1, volume: 1 })).rejects.toThrow(
      'synthesis-failed',
    );
  });

  it('sans synthèse vocale', async () => {
    const engine = new WebSpeechEngine(undefined);
    expect(engine.available()).toBe(false);
    expect(await engine.voices()).toEqual([]);
    await expect(engine.speak('x', { rate: 1, pitch: 1, volume: 1 })).resolves.toBeUndefined();
  });
});
