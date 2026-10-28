// Dispositions de clavier (F-08) : AZERTY (France, par défaut), QWERTY (US)
// et BÉPO. Touches, caractères produits et doigt conseillé. Servent à la
// carte de chaleur (F-34) et au clavier virtuel.

export type Finger = 'l5' | 'l4' | 'l3' | 'l2' | 'r2' | 'r3' | 'r4' | 'r5' | 'thumb';
export type LayoutId = 'azerty' | 'qwerty' | 'bepo';

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

type KeyOpts = {
  shift?: string[];
  altgr?: string[];
  accents?: string[];
  base?: string;
  width?: number;
};

/** Définit une touche : libellé, doigt, caractères de chaque niveau. */
function key(label: string, finger: Finger, opts: KeyOpts = {}): KeyDef {
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

// Doigt conseillé selon la colonne (touche à gauche : auriculaire gauche…).
const NUMBER_ROW: Finger[] = [
  'l5',
  'l5',
  'l4',
  'l3',
  'l2',
  'l2',
  'r2',
  'r2',
  'r3',
  'r4',
  'r5',
  'r5',
  'r5',
];
const LETTER_ROW: Finger[] = [
  'l5',
  'l4',
  'l3',
  'l2',
  'l2',
  'r2',
  'r2',
  'r3',
  'r4',
  'r5',
  'r5',
  'r5',
];
const ISO_BOTTOM: Finger[] = ['l5', 'l5', 'l4', 'l3', 'l2', 'l2', 'r2', 'r2', 'r3', 'r4', 'r5'];
const ANSI_BOTTOM: Finger[] = ['l5', 'l4', 'l3', 'l2', 'l2', 'r2', 'r2', 'r3', 'r4', 'r5'];

/** Construit une rangée : [libellé, options] par touche, doigts par colonne. */
function row(fingers: Finger[], keys: (string | [string, KeyOpts])[]): KeyDef[] {
  return keys.map((k, i) => {
    const [label, opts] = typeof k === 'string' ? [k, {}] : k;
    return key(label, fingers[i] ?? 'r5', opts);
  });
}

const SPACE_ROW = [
  key('␣', 'thumb', { base: ' ', width: 6 }),
  key('↵', 'r5', { base: '↵', width: 2 }),
];

const AZERTY: KeyDef[][] = [
  row(NUMBER_ROW, [
    '²',
    ['&', { shift: ['1'] }],
    ['é', { shift: ['2'], altgr: ['~'] }],
    ['"', { shift: ['3'], altgr: ['#'] }],
    ["'", { shift: ['4'], altgr: ['{'], accents: ['’'] }],
    ['(', { shift: ['5'], altgr: ['['] }],
    ['-', { shift: ['6'], altgr: ['|'], accents: ['–', '—'] }],
    ['è', { shift: ['7'], altgr: ['`'] }],
    ['_', { shift: ['8'], altgr: ['\\'] }],
    ['ç', { shift: ['9'] }],
    ['à', { shift: ['0'], altgr: ['@'] }],
    [')', { shift: ['°'], altgr: [']'] }],
    ['=', { shift: ['+'], altgr: ['}'] }],
  ]),
  row(LETTER_ROW, [
    ['a', { accents: ['â', 'ä'] }],
    'z',
    ['e', { accents: ['ê', 'ë'], altgr: ['€'] }],
    'r',
    't',
    ['y', { accents: ['ÿ'] }],
    ['u', { accents: ['û', 'ü'] }],
    ['i', { accents: ['î', 'ï'] }],
    ['o', { accents: ['ô', 'ö', 'œ'] }],
    'p',
    ['^', { shift: ['¨'] }],
    ['$', { shift: ['£'], altgr: ['¤'] }],
  ]),
  row(LETTER_ROW, [
    'q',
    's',
    'd',
    'f',
    'g',
    'h',
    'j',
    'k',
    'l',
    'm',
    ['ù', { shift: ['%'] }],
    ['*', { shift: ['µ'] }],
  ]),
  row(ISO_BOTTOM, [
    ['<', { shift: ['>'] }],
    'w',
    'x',
    'c',
    'v',
    'b',
    'n',
    [',', { shift: ['?'] }],
    [';', { shift: ['.'] }],
    [':', { shift: ['/'] }],
    ['!', { shift: ['§'] }],
  ]),
  SPACE_ROW,
];

const QWERTY: KeyDef[][] = [
  row(NUMBER_ROW, [
    ['`', { shift: ['~'] }],
    ['1', { shift: ['!'] }],
    ['2', { shift: ['@'] }],
    ['3', { shift: ['#'] }],
    ['4', { shift: ['$'] }],
    ['5', { shift: ['%'] }],
    ['6', { shift: ['^'] }],
    ['7', { shift: ['&'] }],
    ['8', { shift: ['*'] }],
    ['9', { shift: ['('] }],
    ['0', { shift: [')'] }],
    ['-', { shift: ['_'], accents: ['–', '—'] }],
    ['=', { shift: ['+'] }],
  ]),
  row(LETTER_ROW, [
    'q',
    'w',
    'e',
    'r',
    't',
    'y',
    'u',
    'i',
    'o',
    'p',
    ['[', { shift: ['{'] }],
    [']', { shift: ['}'] }],
  ]),
  row(LETTER_ROW, [
    'a',
    's',
    'd',
    'f',
    'g',
    'h',
    'j',
    'k',
    'l',
    [';', { shift: [':'] }],
    ["'", { shift: ['"'], accents: ['’'] }],
  ]),
  row(ANSI_BOTTOM, [
    'z',
    'x',
    'c',
    'v',
    'b',
    'n',
    'm',
    [',', { shift: ['<'] }],
    ['.', { shift: ['>'] }],
    ['/', { shift: ['?'] }],
  ]),
  SPACE_ROW,
];

const BEPO: KeyDef[][] = [
  row(NUMBER_ROW, [
    ['$', { shift: ['#'] }],
    ['"', { shift: ['1'] }],
    ['«', { shift: ['2'] }],
    ['»', { shift: ['3'] }],
    ['(', { shift: ['4'] }],
    [')', { shift: ['5'] }],
    ['@', { shift: ['6'] }],
    ['+', { shift: ['7'] }],
    ['-', { shift: ['8'], accents: ['–', '—'] }],
    ['/', { shift: ['9'] }],
    ['*', { shift: ['0'] }],
    ['=', { shift: ['°'] }],
    ['%', { shift: ['`'] }],
  ]),
  row(LETTER_ROW, [
    'b',
    'é',
    'p',
    ['o', { accents: ['ô', 'ö', 'œ'] }],
    'è',
    ['^', { shift: ['!'] }],
    'v',
    'd',
    'l',
    'j',
    'z',
    'w',
  ]),
  row(LETTER_ROW, [
    ['a', { accents: ['â', 'ä'] }],
    ['u', { accents: ['û', 'ü', 'ù'] }],
    ['i', { accents: ['î', 'ï'] }],
    ['e', { accents: ['ë'], altgr: ['€'] }],
    [',', { shift: [';'] }],
    'c',
    't',
    's',
    'r',
    'n',
    'm',
    'ç',
  ]),
  row(ISO_BOTTOM, [
    'ê',
    'à',
    ['y', { accents: ['ÿ'] }],
    'x',
    ['.', { shift: [':'] }],
    'k',
    ["'", { shift: ['?'], accents: ['’'] }],
    'q',
    'g',
    'h',
    'f',
  ]),
  SPACE_ROW,
];

export const LAYOUTS: Record<LayoutId, KeyDef[][]> = { azerty: AZERTY, qwerty: QWERTY, bepo: BEPO };

/** Rangées de la disposition (AZERTY par défaut, F-08). */
export function layoutRows(id: string | undefined): KeyDef[][] {
  return LAYOUTS[(id as LayoutId) in LAYOUTS ? (id as LayoutId) : 'azerty'];
}

/** Touche qui produit un caractère (null si absent de la disposition). */
export function keyFor(ch: string, layout?: string): KeyDef | null {
  const c = ch === '\n' ? '↵' : ch.toLowerCase();
  for (const r of layoutRows(layout)) for (const k of r) if (k.chars.includes(c)) return k;
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

/** Gestes pour taper un caractère : touche, Maj, AltGr, touche morte. */
export function strokeFor(ch: string, layout?: string): Stroke | null {
  const k = keyFor(ch, layout);
  if (!k) return null;
  const lower = ch === '\n' ? '↵' : ch.toLowerCase();
  const upper = ch !== lower && ch.length === 1 && /\p{Lu}/u.test(ch);
  // touche morte seulement si la lettre accentuée n'a pas sa propre touche
  const viaDead =
    k.accents.includes(lower) && (CIRCUMFLEX.includes(lower) || DIAERESIS.includes(lower));
  return {
    key: k,
    shift: upper || k.shift.includes(lower),
    altgr: k.altgr.includes(lower),
    dead: viaDead ? keyFor('^', layout) : null,
    deadShift: viaDead && DIAERESIS.includes(lower),
  };
}

// Compatibilité : la disposition AZERTY reste exportée sous ce nom.
export { AZERTY };
