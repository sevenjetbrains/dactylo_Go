export interface Piece {
  text: string;
  hit: boolean;
}

/**
 * Découpe le texte pour surligner, dans l'ordre, les mots recollés après
 * une césure (aperçu de l'import PDF). Chaque mot est cherché après le
 * précédent : deux « exemple » recollés surlignent deux occurrences.
 */
export function highlightWords(text: string, words: readonly string[]): Piece[] {
  const out: Piece[] = [];
  let from = 0;
  for (const w of words) {
    if (!w) continue;
    const at = text.indexOf(w, from);
    if (at < 0) continue;
    if (at > from) out.push({ text: text.slice(from, at), hit: false });
    out.push({ text: w, hit: true });
    from = at + w.length;
  }
  if (from < text.length) out.push({ text: text.slice(from), hit: false });
  return out;
}

/** Fenêtre de miniatures (au plus `size` pages) autour de la première page choisie. */
export function thumbWindow(first: number, total: number, size = 6): number[] {
  const start = Math.max(1, Math.min(first - 2, total - size + 1));
  const out: number[] = [];
  for (let p = start; p <= Math.min(total, start + size - 1); p++) out.push(p);
  return out;
}
