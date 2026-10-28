import { describe, expect, it } from 'vitest';
import { formatDuration, formatHours, formatNumber, localDay, parseDay } from './format';

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

describe('formatHours', () => {
  it.each([
    [0, '0 min'],
    [45 * 60000, '45 min'],
    [135 * 60000, '2 h 15'],
    [61 * 60000, '1 h 01'],
  ])('%d ms -> %s', (ms, want) => {
    expect(formatHours(ms)).toBe(want);
  });
});

describe('localDay', () => {
  it('donne la date locale', () => {
    expect(localDay(new Date(2026, 9, 7, 23, 30))).toBe('2026-10-07');
    expect(localDay(parseDay('2026-10-28'))).toBe('2026-10-28');
  });
});

describe('formatNumber', () => {
  it('utilise la virgule en français', () => {
    expect(formatNumber(42.567, 'fr', 1)).toBe('42,6');
    expect(formatNumber(42.567, 'en', 1)).toBe('42.6');
  });
});
