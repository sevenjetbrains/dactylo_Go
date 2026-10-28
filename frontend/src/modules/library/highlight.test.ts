import { describe, expect, it } from 'vitest';
import { highlightWords, thumbWindow } from './highlight';

describe('highlightWords', () => {
  it('surligne les mots dans l’ordre', () => {
    expect(highlightWords('un exemple et un autre exemple', ['exemple', 'exemple'])).toEqual([
      { text: 'un ', hit: false },
      { text: 'exemple', hit: true },
      { text: ' et un autre ', hit: false },
      { text: 'exemple', hit: true },
    ]);
  });

  it('ignore les mots absents ou vides', () => {
    expect(highlightWords('abc', ['zz', ''])).toEqual([{ text: 'abc', hit: false }]);
    expect(highlightWords('', [])).toEqual([]);
  });
});

describe('thumbWindow', () => {
  it.each([
    [1, 412, [1, 2, 3, 4, 5, 6]],
    [3, 412, [1, 2, 3, 4, 5, 6]],
    [100, 412, [98, 99, 100, 101, 102, 103]],
    [411, 412, [407, 408, 409, 410, 411, 412]],
    [2, 3, [1, 2, 3]],
  ])('page %d sur %d', (first, total, want) => {
    expect(thumbWindow(first, total)).toEqual(want);
  });
});
