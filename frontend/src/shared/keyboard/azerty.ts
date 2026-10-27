// Disposition AZERTY (France) : touches, caractères produits et largeur.
// Sert à la carte de chaleur (F-34) et servira au clavier virtuel (F-08).

export interface KeyDef {
  /** Libellé principal affiché sur la touche. */
  label: string;
  /** Caractères attribués à la touche (minuscules ; accents morts inclus). */
  chars: string[];
  /** Largeur relative (1 = touche standard). */
  width?: number;
}

const k = (label: string, ...more: string[]): KeyDef => ({ label, chars: [label, ...more] });

export const AZERTY: KeyDef[][] = [
  [
    k('²'),
    k('&', '1'),
    k('é', '2', '~'),
    k('"', '3', '#'),
    k("'", '4', '{', '’'),
    k('(', '5', '['),
    k('-', '6', '|', '–', '—'),
    k('è', '7', '`'),
    k('_', '8', '\\'),
    k('ç', '9', '^'),
    k('à', '0', '@'),
    k(')', '°', ']'),
    k('=', '+', '}'),
  ],
  [
    k('a', 'â', 'ä'),
    k('z'),
    k('e', 'ê', 'ë', '€'),
    k('r'),
    k('t'),
    k('y', 'ÿ'),
    k('u', 'û', 'ü'),
    k('i', 'î', 'ï'),
    k('o', 'ô', 'ö', 'œ'),
    k('p'),
    k('^', '¨'),
    k('$', '£', '¤'),
  ],
  [
    k('q'),
    k('s'),
    k('d'),
    k('f'),
    k('g'),
    k('h'),
    k('j'),
    k('k'),
    k('l'),
    k('m'),
    k('ù', '%'),
    k('*', 'µ'),
  ],
  [
    k('<', '>'),
    k('w'),
    k('x'),
    k('c'),
    k('v'),
    k('b'),
    k('n'),
    k(',', '?'),
    k(';', '.'),
    k(':', '/'),
    k('!', '§'),
  ],
  [
    { label: '␣', chars: [' '], width: 6 },
    { label: '↵', chars: ['↵'], width: 2 },
  ],
];

/** Touche qui produit un caractère (null si absent de la disposition). */
export function keyFor(ch: string): KeyDef | null {
  const c = ch.toLowerCase();
  for (const row of AZERTY) for (const key of row) if (key.chars.includes(c)) return key;
  return null;
}
