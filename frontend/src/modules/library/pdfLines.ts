// Reconstitution des lignes d'une page à partir des fragments de texte de
// pdfjs (getTextContent). Fonction pure, sans dépendance à pdfjs.

export interface PdfTextItem {
  str: string;
  /** Matrice [a, b, c, d, x, y] : x = transform[4], y = transform[5]. */
  transform: number[];
  width: number;
  height: number;
  hasEOL: boolean;
}

/** Regroupe les fragments en lignes, dans l'ordre de lecture du flux. */
export function linesFromItems(items: readonly PdfTextItem[]): string[] {
  const lines: string[] = [];
  let cur = '';
  let prev: PdfTextItem | null = null;

  const push = () => {
    lines.push(cur.replace(/\s+/g, ' ').trim());
    cur = '';
    prev = null;
  };

  for (const it of items) {
    if (prev) {
      const fontSize = Math.hypot(it.transform[0] ?? 0, it.transform[1] ?? 0) || it.height || 10;
      const dy = Math.abs((it.transform[5] ?? 0) - (prev.transform[5] ?? 0));
      if (dy > fontSize * 0.5) {
        push(); // changement de ligne sans marque de fin de ligne
      } else if (it.str) {
        const gap = (it.transform[4] ?? 0) - ((prev.transform[4] ?? 0) + prev.width);
        if (gap > fontSize * 0.15 && !cur.endsWith(' ') && !it.str.startsWith(' ')) cur += ' ';
      }
    }
    cur += it.str;
    if (it.str || it.width) prev = it;
    if (it.hasEOL) push();
  }
  if (cur.trim()) push();
  // supprimer les lignes vides en fin de page
  while (lines.length > 0 && lines.at(-1) === '') lines.pop();
  return lines;
}
