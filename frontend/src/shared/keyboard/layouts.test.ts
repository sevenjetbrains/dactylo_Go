import { describe, expect, it } from 'vitest';
import { keyFor, layoutRows, strokeFor } from './layouts';

describe('strokeFor (AZERTY, par défaut)', () => {
  it('lettre simple, doigt conseillé', () => {
    const s = strokeFor('q')!;
    expect(s.key.label).toBe('q');
    expect(s.key.finger).toBe('l5');
    expect(s.shift || s.altgr || s.dead).toBeFalsy();
  });

  it('majuscules et chiffres demandent Maj', () => {
    expect(strokeFor('Q')).toMatchObject({ shift: true });
    expect(strokeFor('2')).toMatchObject({ shift: true });
    expect(strokeFor('2')!.key.label).toBe('é');
    expect(strokeFor('.')).toMatchObject({ shift: true });
    expect(strokeFor('é')).toMatchObject({ shift: false });
  });

  it('AltGr', () => {
    expect(strokeFor('@')).toMatchObject({ altgr: true });
    expect(strokeFor('@')!.key.label).toBe('à');
  });

  it('touches mortes pour les accents', () => {
    const e = strokeFor('ê')!;
    expect(e.key.label).toBe('e');
    expect(e.dead?.label).toBe('^');
    expect(e.deadShift).toBe(false);
    expect(strokeFor('ï')).toMatchObject({ deadShift: true });
    expect(strokeFor('^')!.dead).toBeNull();
  });

  it('espace, fin de ligne, caractère inconnu', () => {
    expect(strokeFor(' ')!.key.finger).toBe('thumb');
    expect(strokeFor('\n')!.key.label).toBe('↵');
    expect(strokeFor('ß')).toBeNull();
  });
});

describe('QWERTY et BÉPO (F-08)', () => {
  it('QWERTY : chiffres directs, pas d’accents', () => {
    expect(strokeFor('2', 'qwerty')).toMatchObject({ shift: false });
    expect(strokeFor('a', 'qwerty')!.key.finger).toBe('l5');
    expect(strokeFor('!', 'qwerty')).toMatchObject({ shift: true });
    expect(strokeFor('é', 'qwerty')).toBeNull();
  });

  it('BÉPO : lettres accentuées directes, chiffres avec Maj', () => {
    expect(strokeFor('é', 'bepo')).toMatchObject({ shift: false, dead: null });
    expect(strokeFor('ê', 'bepo')!.dead).toBeNull(); // touche ê dédiée
    expect(strokeFor('e', 'bepo')!.key.finger).toBe('l2');
    expect(strokeFor('1', 'bepo')).toMatchObject({ shift: true });
    expect(strokeFor('1', 'bepo')!.key.label).toBe('"');
    expect(strokeFor('â', 'bepo')!.dead?.label).toBe('^');
  });

  it('disposition inconnue : AZERTY', () => {
    expect(layoutRows('dvorak')).toBe(layoutRows('azerty'));
    expect(keyFor('z', 'dvorak')!.finger).toBe('l4');
  });

  it('chaque caractère n’a qu’une touche par disposition', () => {
    for (const id of ['azerty', 'qwerty', 'bepo']) {
      const seen = new Map<string, string>();
      for (const r of layoutRows(id))
        for (const k of r)
          for (const c of k.chars) {
            expect(
              seen.get(c),
              `${id} : « ${c} » sur ${seen.get(c)} et ${k.label}`,
            ).toBeUndefined();
            seen.set(c, k.label);
          }
    }
  });
});
