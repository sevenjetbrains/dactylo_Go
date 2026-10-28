// Disposition AZERTY (France) : touches, caractères produits, doigt conseillé.
// Sert à la carte de chaleur (F-34) et au clavier virtuel (F-08).

export type Finger = 'l5' | 'l4' | 'l3' | 'l2' | 'r2' | 'r3' | 'r4' | 'r5' | 'thumb';

export interface KeyDef {
  /** Libellé principal affiché sur la touche. */
  label: string;
  /** Caractère sans modificateur. */
  base: string;
  /** Caractères avec Maj. */
  shift: string[];
  /** Caractères avec AltGr. */
  altgr: string[];
  /** Lettres accentuées obtenues par touche morte puis cette touche. */
  accents: string[];
  /** Tous les caractères attribués à la touche (minuscules), pour les statistiques. */
  chars: string[];
  finger: Finger;
  /** Largeur relative (1 = touche standard). */
  width?: number;
}

function key(
  label: string,
  finger: Finger,
  opts: {
    shift?: string[];
    altgr?: string[];
    accents?: string[];
    base?: string;
    width?: number;
  } = {},
): KeyDef {
  const base = opts.base ?? label;
  const shift = opts.shift ?? [];
  const altgr = opts.altgr ?? [];
  const accents = opts.accents ?? [];
  return {
    label,
    base,
    shift,
    altgr,
    accents,
    chars: [base, ...accents, ...shift, ...altgr],
    finger,
    width: opts.width,
  };
}

export const AZERTY: KeyDef[][] = [
  [
    key('²', 'l5'),
    key('&', 'l5', { shift: ['1'] }),
    key('é', 'l4', { shift: ['2'], altgr: ['~'] }),
    key('"', 'l3', { shift: ['3'], altgr: ['#'] }),
    key("'", 'l2', { shift: ['4'], altgr: ['{'], accents: ['’'] }),
    key('(', 'l2', { shift: ['5'], altgr: ['['] }),
    key('-', 'r2', { shift: ['6'], altgr: ['|'], accents: ['–', '—'] }),
    key('è', 'r2', { shift: ['7'], altgr: ['`'] }),
    key('_', 'r3', { shift: ['8'], altgr: ['\\'] }),
    key('ç', 'r4', { shift: ['9'] }),
    key('à', 'r5', { shift: ['0'], altgr: ['@'] }),
    key(')', 'r5', { shift: ['°'], altgr: [']'] }),
    key('=', 'r5', { shift: ['+'], altgr: ['}'] }),
  ],
  [
    key('a', 'l5', { accents: ['â', 'ä'] }),
    key('z', 'l4'),
    key('e', 'l3', { accents: ['ê', 'ë'], altgr: ['€'] }),
    key('r', 'l2'),
    key('t', 'l2'),
    key('y', 'r2', { accents: ['ÿ'] }),
    key('u', 'r2', { accents: ['û', 'ü'] }),
    key('i', 'r3', { accents: ['î', 'ï'] }),
    key('o', 'r4', { accents: ['ô', 'ö', 'œ'] }),
    key('p', 'r5'),
    key('^', 'r5', { shift: ['¨'] }),
    key('$', 'r5', { shift: ['£'], altgr: ['¤'] }),
  ],
  [
    key('q', 'l5'),
    key('s', 'l4'),
    key('d', 'l3'),
    key('f', 'l2'),
    key('g', 'l2'),
    key('h', 'r2'),
    key('j', 'r2'),
    key('k', 'r3'),
    key('l', 'r4'),
    key('m', 'r5'),
    key('ù', 'r5', { shift: ['%'] }),
    key('*', 'r5', { shift: ['µ'] }),
  ],
  [
    key('<', 'l5', { shift: ['>'] }),
    key('w', 'l5'),
    key('x', 'l4'),
    key('c', 'l3'),
    key('v', 'l2'),
    key('b', 'l2'),
    key('n', 'r2'),
    key(',', 'r3', { shift: ['?'] }),
    key(';', 'r4', { shift: ['.'] }),
    key(':', 'r5', { shift: ['/'] }),
    key('!', 'r5', { shift: ['§'] }),
  ],
  [key('␣', 'thumb', { base: ' ', width: 6 }), key('↵', 'r5', { base: '↵', width: 2 })],
];

/** Touche qui produit un caractère (null si absent de la disposition). */
export function keyFor(ch: string): KeyDef | null {
  const c = ch === '\n' ? '↵' : ch.toLowerCase();
  for (const row of AZERTY) for (const k of row) if (k.chars.includes(c)) return k;
  return null;
}

/** Ce qu'il faut presser pour un caractère (clavier virtuel, F-08). */
export interface Stroke {
  key: KeyDef;
  shift: boolean;
  altgr: boolean;
  /** Touche morte à presser avant (^ ou ¨), et si elle demande Maj. */
  dead: KeyDef | null;
  deadShift: boolean;
}

const CIRCUMFLEX = 'âêîôû';
const DIAERESIS = 'äëïöüÿ';

export function strokeFor(ch: string): Stroke | null {
  const k = keyFor(ch);
  if (!k) return null;
  const lower = ch === '\n' ? '↵' : ch.toLowerCase();
  const upper = ch !== lower && ch.length === 1 && /\p{Lu}/u.test(ch);
  const dead = CIRCUMFLEX.includes(lower) || DIAERESIS.includes(lower) ? keyFor('^') : null;
  return {
    key: k,
    shift: upper || k.shift.includes(lower),
    altgr: k.altgr.includes(lower),
    dead,
    deadShift: DIAERESIS.includes(lower),
  };
}
