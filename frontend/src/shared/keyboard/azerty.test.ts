import { describe, expect, it } from 'vitest';
import { strokeFor } from './azerty';

describe('strokeFor (AZERTY)', () => {
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
