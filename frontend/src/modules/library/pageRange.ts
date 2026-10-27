/**
 * Lit une sélection de pages : « 3 », « 3-5 », « 1-3, 7 ». Renvoie les
 * numéros triés sans doublon, ou null si la saisie est invalide ou sort
 * du document (1 à `max`).
 */
export function parsePageRange(input: string, max: number): number[] | null {
  const parts = input
    .split(/[,;]/)
    .map((p) => p.trim())
    .filter(Boolean);
  if (parts.length === 0) return null;
  const pages = new Set<number>();
  for (const part of parts) {
    const m = /^(\d+)\s*(?:[-–à]\s*(\d+))?$/.exec(part);
    if (!m) return null;
    const from = Number(m[1]);
    const to = m[2] === undefined ? from : Number(m[2]);
    if (from < 1 || to > max || from > to) return null;
    for (let p = from; p <= to; p++) pages.add(p);
  }
  return [...pages].sort((a, b) => a - b);
}

/** Écriture compacte d'une liste de pages : [3, 4, 5, 7] → « 3-5, 7 ». */
export function formatPageRange(pages: readonly number[]): string {
  const out: string[] = [];
  for (let i = 0; i < pages.length; i++) {
    const start = pages[i]!;
    while (i + 1 < pages.length && pages[i + 1] === pages[i]! + 1) i++;
    out.push(start === pages[i] ? `${start}` : `${start}-${pages[i]}`);
  }
  return out.join(', ');
}
