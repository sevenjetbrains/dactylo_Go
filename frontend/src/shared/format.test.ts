import { describe, expect, it } from 'vitest';
import { formatDuration, formatNumber } from './format';

describe('formatDuration', () => {
  it.each([
    [0, '0:00'],
    [999, '0:00'],
    [61_000, '1:01'],
    [600_000, '10:00'],
    [3_723_000, '1:02:03'],
    [-5, '0:00'],
  ])('%d ms -> %s', (ms, want) => {
    expect(formatDuration(ms)).toBe(want);
  });
});

describe('formatNumber', () => {
  it('utilise la virgule en français', () => {
    expect(formatNumber(42.567, 'fr', 1)).toBe('42,6');
    expect(formatNumber(42.567, 'en', 1)).toBe('42.6');
  });
});
