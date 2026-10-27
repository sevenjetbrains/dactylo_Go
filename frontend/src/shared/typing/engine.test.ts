import { describe, expect, it } from 'vitest';
import {
  BACKSPACE,
  CharState,
  matches,
  TypingEngine,
  wordEnds,
  type EngineOptions,
} from './engine';

const free: EngineOptions = { correction: 'free', allowBackspace: true, foldTypography: true };
const blocking: EngineOptions = { ...free, correction: 'blocking' };

function typeAll(e: TypingEngine, s: string) {
  for (const ch of s) e.type(ch, 0);
}

describe('matches', () => {
  it('suit les règles du cœur Go', () => {
    expect(matches('ê', 'ê', false)).toBe(true);
    expect(matches('ê', 'e', false)).toBe(false);
    expect(matches('\n', ' ', false)).toBe(true);
    expect(matches(' ', ' ', false)).toBe(true);
    expect(matches('’', "'", false)).toBe(false);
    expect(matches('’', "'", true)).toBe(true);
    expect(matches('«', '"', true)).toBe(true);
  });
});

describe('TypingEngine', () => {
  it('compte une lettre accentuée comme un seul caractère juste (touche morte)', () => {
    const e = new TypingEngine('être', free); // texte décomposé
    expect(e.chars).toHaveLength(4);
    expect(e.type('ê', 10)).toBe('correct');
    expect(e.keystrokes).toBe(1);
    expect(e.state[0]).toBe(CharState.Correct);
  });

  it('mode libre : avance sur une erreur', () => {
    const e = new TypingEngine('chat', free);
    typeAll(e, 'chot');
    expect(e.done).toBe(true);
    expect(Array.from(e.state)).toEqual([1, 1, 2, 1]);
    expect(e.accuracy).toBe(75);
  });

  it('mode bloquant : reste sur l’erreur puis marque la correction', () => {
    const e = new TypingEngine('ab', blocking);
    expect(e.type('x', 0)).toBe('error');
    expect(e.position).toBe(0);
    expect(e.state[0]).toBe(CharState.Error);
    e.type('a', 0);
    expect(e.position).toBe(1);
    expect(e.state[0]).toBe(CharState.Corrected);
  });

  it('retour arrière autorisé ou non', () => {
    const e = new TypingEngine('ab', free);
    typeAll(e, 'x');
    expect(e.backspace(5)).toBe(true);
    expect(e.position).toBe(0);
    expect(e.state[0]).toBe(CharState.Pending);
    e.type('a', 6);
    expect(e.state[0]).toBe(CharState.Corrected);

    const strict = new TypingEngine('ab', { ...free, allowBackspace: false });
    strict.type('x', 0);
    expect(strict.backspace(1)).toBe(false);
    expect(strict.position).toBe(1);
    expect(new TypingEngine('ab', free).backspace(0)).toBe(false);
  });

  it('produit les frappes brutes numérotées, une seule fois', () => {
    const e = new TypingEngine('abc', free);
    e.type('a', 100.4);
    e.type('x', 200);
    e.backspace(300);
    expect(e.take()).toEqual([
      { seq: 0, position: 0, typed: 'a', timeMs: 100 },
      { seq: 1, position: 1, typed: 'x', timeMs: 200 },
      { seq: 2, position: 1, typed: BACKSPACE, timeMs: 300 },
    ]);
    expect(e.take()).toEqual([]);
  });

  it('saute la ligne vide entre deux paragraphes', () => {
    const e = new TypingEngine('a\n\nb', free);
    e.type('a', 0);
    e.type('\n', 0);
    expect(e.position).toBe(3);
    e.backspace(0);
    expect(e.position).toBe(1);
  });

  it('s’arrête à la position de fin et ignore la suite', () => {
    const e = new TypingEngine('un deux trois', { ...free, end: 2 });
    typeAll(e, 'un');
    expect(e.done).toBe(true);
    expect(e.progress).toBe(1);
    expect(e.type(' ', 0)).toBe('ignored');
    expect(e.backspace(0)).toBe(false);
  });
});

describe('wordEnds', () => {
  it('donne la fin de chaque mot', () => {
    expect(wordEnds(Array.from(' un  deux\ntrois'))).toEqual([3, 9, 15]);
    expect(wordEnds([])).toEqual([]);
  });
});
