import { describe, expect, it } from 'vitest';
import { DEFAULTS, sanitize } from './settings.svelte';

describe('sanitize', () => {
  it('complète avec les valeurs par défaut', () => {
    expect(sanitize(undefined)).toEqual(DEFAULTS);
    expect(sanitize({ visibleLines: 5 })).toEqual({ ...DEFAULTS, visibleLines: 5 });
  });

  it('borne et filtre les valeurs invalides', () => {
    const s = sanitize({
      fontSize: 99,
      visibleLines: 0,
      correction: 'autre',
      durationMin: 3,
      allowBackspace: 'oui',
      wordCount: 2.6,
    });
    expect(s.fontSize).toBe(32);
    expect(s.visibleLines).toBe(1);
    expect(s.correction).toBe('free');
    expect(s.durationMin).toBe(1);
    expect(s.allowBackspace).toBe(true);
    expect(s.wordCount).toBe(5);
  });

  it('police et sons de frappe (F-51)', () => {
    expect(sanitize({ font: 'courier', sounds: true })).toMatchObject({
      font: 'courier',
      sounds: true,
    });
    expect(sanitize({ font: 'comic', sounds: 1 })).toMatchObject({
      font: 'jetbrains',
      sounds: false,
    });
  });
});
