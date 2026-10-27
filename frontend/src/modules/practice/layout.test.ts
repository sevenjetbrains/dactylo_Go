import { describe, expect, it } from 'vitest';
import { firstVisibleLine, lineAt, wrapLines, type Line } from './layout';

const text = (lines: Line[], s: string) => lines.map((l) => s.slice(l.start, l.end));

describe('wrapLines', () => {
  it('coupe après une espace sans dépasser la largeur', () => {
    const s = 'le petit chat dort';
    const lines = wrapLines(Array.from(s), 10);
    expect(text(lines, s)).toEqual(['le petit ', 'chat dort']);
    for (const l of lines) expect(l.end - l.start).toBeLessThanOrEqual(10);
  });

  it('couvre tout le texte, sans trou ni chevauchement', () => {
    const s = 'Un texte un peu plus long, avec des mots variés.\nEt une seconde ligne ici.';
    const lines = wrapLines(Array.from(s), 12);
    expect(lines[0]!.start).toBe(0);
    expect(lines.at(-1)!.end).toBe(s.length);
    for (let i = 1; i < lines.length; i++) expect(lines[i]!.start).toBe(lines[i - 1]!.end);
    expect(text(lines, s).join('')).toBe(s);
  });

  it('coupe un mot trop long et respecte les fins de ligne', () => {
    const s = 'abcdefghij\nxy';
    expect(text(wrapLines(Array.from(s), 4), s)).toEqual(['abcd', 'efgh', 'ij\n', 'xy']);
  });

  it('gère le texte vide', () => {
    expect(wrapLines([], 10)).toEqual([{ start: 0, end: 0 }]);
  });

  it('reste rapide sur 10 000 mots (NF-01)', () => {
    const chars = Array.from('lorem ipsum dolor sit amet '.repeat(2000));
    const t0 = performance.now();
    const lines = wrapLines(chars, 60);
    expect(performance.now() - t0).toBeLessThan(100);
    expect(lineAt(lines, chars.length - 1)).toBe(lines.length - 1);
  });
});

describe('lineAt', () => {
  const lines = [
    { start: 0, end: 5 },
    { start: 5, end: 9 },
    { start: 9, end: 12 },
  ];
  it.each([
    [0, 0],
    [4, 0],
    [5, 1],
    [11, 2],
    [12, 2],
  ])('position %d -> ligne %d', (p, want) => {
    expect(lineAt(lines, p)).toBe(want);
  });
});

describe('firstVisibleLine', () => {
  it('garde la ligne active en haut', () => {
    expect(firstVisibleLine(4, 3, 10, 'top')).toBe(4);
  });
  it('centre la ligne active sans dépasser les bords', () => {
    expect(firstVisibleLine(0, 3, 10, 'center')).toBe(0);
    expect(firstVisibleLine(5, 3, 10, 'center')).toBe(4);
    expect(firstVisibleLine(9, 3, 10, 'center')).toBe(7);
  });
});
