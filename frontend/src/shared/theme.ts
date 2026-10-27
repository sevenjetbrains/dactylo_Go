export type Theme = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'dactylo.theme';

export function getTheme(): Theme {
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    if (v === 'light' || v === 'dark' || v === 'system') return v;
  } catch {
    // stockage indisponible
  }
  return 'system';
}

export function isDark(theme: Theme): boolean {
  if (theme === 'system')
    return window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? false;
  return theme === 'dark';
}

export function applyTheme(theme: Theme): void {
  document.documentElement.classList.toggle('dark', isDark(theme));
}

export function setTheme(theme: Theme): void {
  applyTheme(theme);
  try {
    localStorage.setItem(STORAGE_KEY, theme);
  } catch {
    // stockage indisponible : le thème ne sera pas mémorisé
  }
}
