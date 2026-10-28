export type Theme = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'dactylo.theme';

/** Thème mémorisé (clair, sombre ou celui du système). */
export function getTheme(): Theme {
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    if (v === 'light' || v === 'dark' || v === 'system') return v;
  } catch {
    // stockage indisponible
  }
  return 'system';
}

/** Le thème donne-t-il un affichage sombre ? */
export function isDark(theme: Theme): boolean {
  if (theme === 'system')
    return window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? false;
  return theme === 'dark';
}

/** Applique le thème au document. */
export function applyTheme(theme: Theme): void {
  document.documentElement.classList.toggle('dark', isDark(theme));
}

/** Change de thème et le mémorise. */
export function setTheme(theme: Theme): void {
  applyTheme(theme);
  try {
    localStorage.setItem(STORAGE_KEY, theme);
  } catch {
    // stockage indisponible : le thème ne sera pas mémorisé
  }
}
