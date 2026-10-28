/** Durée en m:ss (ou h:mm:ss au-delà d'une heure). */
export function formatDuration(ms: number): string {
  const total = Math.max(0, Math.floor(ms / 1000));
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  const ss = String(s).padStart(2, '0');
  return h > 0 ? `${h}:${String(m).padStart(2, '0')}:${ss}` : `${m}:${ss}`;
}

/** Durée longue : « 2 h 15 », « 45 min ». */
export function formatHours(ms: number): string {
  const min = Math.round(Math.max(0, ms) / 60000);
  if (min < 60) return `${min} min`;
  return `${Math.floor(min / 60)} h ${String(min % 60).padStart(2, '0')}`;
}

/** Date locale AAAA-MM-JJ (même convention que les révisions côté Go). */
export function localDay(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}

/** Lit « AAAA-MM-JJ » (jour local) ou un horodatage ISO. */
export function parseDay(s: string): Date {
  return s.length === 10 ? new Date(`${s}T12:00:00`) : new Date(s);
}

/** Nombre arrondi, au format de la langue (virgule décimale en français). */
export function formatNumber(n: number, locale: string, digits = 0): string {
  return new Intl.NumberFormat(locale, {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(n);
}
