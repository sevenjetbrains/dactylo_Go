/** Ligne affichée : caractères [start, end) du texte. */
export interface Line {
  start: number;
  end: number;
}

/**
 * Découpe le texte en lignes d'au plus `columns` caractères, en coupant
 * après une espace quand c'est possible. Une fin de ligne termine sa ligne.
 * Police monospace : le nombre de colonnes suffit, sans mesurer le DOM.
 */
export function wrapLines(chars: readonly string[], columns: number): Line[] {
  const cols = Math.max(1, Math.floor(columns));
  const lines: Line[] = [];
  let start = 0;
  let lastSpace = -1;
  for (let i = 0; i < chars.length; i++) {
    const c = chars[i]!;
    if (c === '\n') {
      lines.push({ start, end: i + 1 });
      start = i + 1;
      lastSpace = -1;
      continue;
    }
    if (c === ' ' || c === ' ') lastSpace = i;
    if (i - start + 1 > cols) {
      // l'espace de coupure reste en fin de ligne (elle se tape aussi)
      const cut = lastSpace >= start && c !== ' ' ? lastSpace + 1 : i;
      lines.push({ start, end: cut });
      start = cut;
      lastSpace = -1;
      for (let j = start; j <= i; j++) if (chars[j] === ' ') lastSpace = j;
    }
  }
  if (start < chars.length || lines.length === 0) lines.push({ start, end: chars.length });
  return lines;
}

/** Indice de la ligne qui contient la position (recherche dichotomique). */
export function lineAt(lines: readonly Line[], position: number): number {
  let lo = 0;
  let hi = lines.length - 1;
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1;
    if (lines[mid]!.start <= position) lo = mid;
    else hi = mid - 1;
  }
  return lo;
}

/** Première ligne visible pour garder la ligne active en haut ou au centre (F-03). */
export function firstVisibleLine(
  current: number,
  visible: number,
  total: number,
  position: 'top' | 'center',
): number {
  const offset = position === 'center' ? Math.floor((visible - 1) / 2) : 0;
  const max = Math.max(0, total - visible);
  return Math.min(Math.max(0, current - offset), position === 'center' ? max : current);
}
