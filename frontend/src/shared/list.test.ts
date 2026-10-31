import { describe, expect, it } from 'vitest';
import { fold, matches, pageNumbers, paginate } from './list';

describe('matches', () => {
  it('ignore accents et majuscules', () => {
    expect(fold('Élève Ça')).toBe('eleve ca');
    expect(matches('ELEVE', 'élève')).toBe(true);
    expect(matches('netre', 'Fenêtre')).toBe(true);
    expect(matches('chat', 'chien', 'oiseau')).toBe(false);
  });

  it('une recherche vide correspond à tout', () => {
    expect(matches('  ', 'mot')).toBe(true);
  });
});

describe('paginate', () => {
  const list = Array.from({ length: 53 }, (_, i) => i);

  it('découpe en pages', () => {
    const p = paginate(list, 2, 25);
    expect(p).toMatchObject({ page: 2, pages: 3, total: 53 });
    expect(p.items).toEqual([50, 51, 52]);
  });

  it('ramène la page dans les bornes', () => {
    expect(paginate(list, 9, 25).page).toBe(2);
    expect(paginate(list, -1, 25).page).toBe(0);
    expect(paginate([], 3, 25)).toMatchObject({ page: 0, pages: 1, total: 0, items: [] });
  });
});

describe('pageNumbers', () => {
  it('affiche toutes les pages quand il y en a peu', () => {
    expect(pageNumbers(0, 1)).toEqual([0]);
    expect(pageNumbers(1, 4)).toEqual([0, 1, 2, 3]);
  });

  it('coupe autour de la page courante', () => {
    expect(pageNumbers(0, 15)).toEqual([0, 1, null, 14]);
    expect(pageNumbers(7, 15)).toEqual([0, null, 6, 7, 8, null, 14]);
    expect(pageNumbers(14, 15)).toEqual([0, null, 13, 14]);
  });

  it('ne coupe pas pour une seule page cachée', () => {
    expect(pageNumbers(3, 15)).toEqual([0, 1, 2, 3, 4, null, 14]);
  });
});
