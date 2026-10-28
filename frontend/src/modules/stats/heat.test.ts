import { describe, expect, it } from 'vitest';
import { keyFor } from '$shared/keyboard/layouts';
import { heatByKey, scaleRange, stepFor } from './heat';

describe('keyFor (AZERTY)', () => {
  it('rattache chiffres, accents et majuscules à leur touche', () => {
    expect(keyFor('2')?.label).toBe('é');
    expect(keyFor('é')?.label).toBe('é');
    expect(keyFor('ê')?.label).toBe('e');
    expect(keyFor('A')?.label).toBe('a');
    expect(keyFor('.')?.label).toBe(';');
    expect(keyFor(' ')?.label).toBe('␣');
    expect(keyFor('↵')?.label).toBe('↵');
    expect(keyFor('ß')).toBeNull();
  });
});

describe('heatByKey', () => {
  const stats = [
    { key: 'e', count: 10, errors: 1, avgMs: 100 },
    { key: 'ê', count: 10, errors: 3, avgMs: 300 },
    { key: 'a', count: 20, errors: 0, avgMs: 150 },
    { key: 'z', count: 2, errors: 2, avgMs: 900 }, // trop peu de frappes
    { key: 'ß', count: 50, errors: 50, avgMs: 999 }, // hors disposition
  ];
  const heat = heatByKey(stats);

  it('regroupe et pondère', () => {
    const e = heat.get(keyFor('e')!)!;
    expect(e).toEqual({ count: 20, errors: 4, errorRate: 20, avgMs: 200 });
    expect(heat.size).toBe(3);
  });

  it('échelonne sur les touches assez fréquentes', () => {
    const range = scaleRange(heat, 'errors');
    expect(range).toEqual([0, 20]);
    expect(stepFor(heat.get(keyFor('e')!), 'errors', range)).toBe(6);
    expect(stepFor(heat.get(keyFor('a')!), 'errors', range)).toBe(0);
    expect(stepFor(heat.get(keyFor('z')!), 'errors', range)).toBeNull();
    expect(stepFor(undefined, 'errors', range)).toBeNull();
    expect(scaleRange(heat, 'speed')).toEqual([150, 200]);
    expect(stepFor(heat.get(keyFor('a')!), 'speed', [150, 150])).toBe(0);
    expect(scaleRange(new Map(), 'errors')).toEqual([0, 0]);
  });
});
