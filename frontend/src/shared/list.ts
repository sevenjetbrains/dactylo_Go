// Outils des listes longues : recherche sans accents ni majuscules, et
// découpage en pages.

/** Forme de comparaison : minuscules, sans accents (« Élève » → « eleve »). */
export function fold(s: string): string {
  return s.normalize('NFD').replace(/\p{M}/gu, '').toLocaleLowerCase('fr');
}

/** Vrai si l'un des champs contient la recherche (vide : tout correspond). */
export function matches(query: string, ...fields: string[]): boolean {
  const q = fold(query.trim());
  return q === '' || fields.some((f) => fold(f).includes(q));
}

export interface Page<T> {
  items: T[];
  /** Page affichée (0 = première), ramenée dans les bornes. */
  page: number;
  pages: number;
  total: number;
}

/**
 * Numéros de page à afficher (0 = première) : la première, la dernière et
 * les voisines de la page courante ; null marque une coupure (« … »).
 */
export function pageNumbers(page: number, pages: number, around = 1): (number | null)[] {
  const out: (number | null)[] = [];
  for (let p = 0; p < pages; p++) {
    if (p === 0 || p === pages - 1 || Math.abs(p - page) <= around) out.push(p);
    else if (out.at(-1) !== null) out.push(null);
  }
  // une coupure qui ne cache qu'une page : afficher la page à la place
  return out.map((p, i) =>
    p === null && out[i - 1] !== null && out[i + 1] === (out[i - 1] as number) + 2
      ? (out[i - 1] as number) + 1
      : p,
  );
}

/** Découpe la liste et renvoie la page demandée (bornée). */
export function paginate<T>(items: readonly T[], page: number, size: number): Page<T> {
  const pages = Math.max(1, Math.ceil(items.length / size));
  const p = Math.min(Math.max(0, Math.floor(page)), pages - 1);
  return { items: items.slice(p * size, (p + 1) * size), page: p, pages, total: items.length };
}
