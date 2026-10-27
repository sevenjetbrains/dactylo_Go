import { describe, expect, it } from 'vitest';
import { formatPageRange, parsePageRange } from './pageRange';
import { linesFromItems, type PdfTextItem } from './pdfLines';

describe('parsePageRange', () => {
  it.each([
    ['3', [3]],
    ['3-5', [3, 4, 5]],
    [' 3 - 5 ', [3, 4, 5]],
    ['1-2, 5', [1, 2, 5]],
    ['5, 1-2, 2', [1, 2, 5]],
    ['3 à 4', [3, 4]],
  ])('%s', (input, want) => {
    expect(parsePageRange(input, 10)).toEqual(want);
  });

  it.each(['', '0', '11', '5-3', 'abc', '2-', '1,,x'])('refuse « %s »', (input) => {
    expect(parsePageRange(input, 10)).toBeNull();
  });
});

describe('formatPageRange', () => {
  it('compacte les plages', () => {
    expect(formatPageRange([3, 4, 5, 7])).toBe('3-5, 7');
    expect(formatPageRange([1])).toBe('1');
    expect(formatPageRange([])).toBe('');
  });
});

const item = (str: string, x: number, y: number, width: number, hasEOL = false): PdfTextItem => ({
  str,
  transform: [12, 0, 0, 12, x, y],
  width,
  height: 12,
  hasEOL,
});

describe('linesFromItems', () => {
  it('regroupe par ligne et insère les espaces manquantes', () => {
    const lines = linesFromItems([
      item('Il était', 72, 700, 40),
      item('une fois', 115, 700, 40, true), // écart de 3 → espace
      item('Mot', 72, 685, 20),
      item('collé', 92, 685, 30), // pas d'écart → pas d'espace
      item('Nouvelle', 72, 670, 40), // changement de y sans hasEOL
      item('', 0, 0, 0, true),
    ]);
    expect(lines).toEqual(['Il était une fois', 'Motcollé', 'Nouvelle']);
  });

  it('page scannée : aucune ligne', () => {
    expect(linesFromItems([])).toEqual([]);
    expect(linesFromItems([item('', 0, 0, 0, true)])).toEqual([]);
  });

  it('normalise les espaces multiples', () => {
    expect(linesFromItems([item('a   b ', 72, 700, 30, true)])).toEqual(['a b']);
  });
});
