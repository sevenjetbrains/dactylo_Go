import { keyFor, layoutRows, type KeyDef } from '$shared/keyboard/layouts';

export interface KeyStatLike {
  key: string;
  count: number;
  errors: number;
  avgMs: number;
}

export interface KeyHeat {
  count: number;
  errors: number;
  errorRate: number; // %
  avgMs: number;
}

export type HeatMetric = 'errors' | 'speed';

/** En dessous, une touche n'a pas assez de frappes pour être colorée. */
export const MIN_COUNT = 5;
/** Nombre de pas de la rampe séquentielle (--viz-seq-0 à 6). */
export const STEPS = 7;

/** Regroupe les mesures par touche du clavier (é et 2 sur la même touche...). */
export function heatByKey(stats: readonly KeyStatLike[], layout?: string): Map<KeyDef, KeyHeat> {
  const out = new Map<KeyDef, KeyHeat>();
  for (const s of stats) {
    const key = keyFor(s.key, layout);
    if (!key) continue;
    const h = out.get(key) ?? { count: 0, errors: 0, errorRate: 0, avgMs: 0 };
    // moyenne pondérée par le nombre de frappes
    h.avgMs = (h.avgMs * h.count + s.avgMs * s.count) / (h.count + s.count);
    h.count += s.count;
    h.errors += s.errors;
    h.errorRate = (100 * h.errors) / h.count;
    out.set(key, h);
  }
  return out;
}

export function metricValue(h: KeyHeat, metric: HeatMetric): number {
  return metric === 'errors' ? h.errorRate : h.avgMs;
}

/** Bornes de l'échelle sur les touches assez fréquentes. */
export function scaleRange(heat: Map<KeyDef, KeyHeat>, metric: HeatMetric): [number, number] {
  const values = [...heat.values()]
    .filter((h) => h.count >= MIN_COUNT)
    .map((h) => metricValue(h, metric));
  if (values.length === 0) return [0, 0];
  return [Math.min(...values), Math.max(...values)];
}

/** Pas de la rampe (0 à STEPS-1) ; null si pas assez de données. */
export function stepFor(
  h: KeyHeat | undefined,
  metric: HeatMetric,
  [min, max]: [number, number],
): number | null {
  if (!h || h.count < MIN_COUNT) return null;
  if (max <= min) return 0;
  const t = (metricValue(h, metric) - min) / (max - min);
  return Math.min(STEPS - 1, Math.max(0, Math.round(t * (STEPS - 1))));
}

export { layoutRows };
