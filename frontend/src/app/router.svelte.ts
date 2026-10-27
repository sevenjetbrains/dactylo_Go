import { normalizePath } from './registry';

// Routage par hash : compatible avec le serveur d'assets embarqué de Wails.
let hash = $state(window.location.hash);

window.addEventListener('hashchange', () => {
  hash = window.location.hash;
});

export const router = {
  get path(): string {
    return normalizePath(hash);
  },
  /** Paramètres après « ? » (ex. #/practice?text=12). */
  get query(): URLSearchParams {
    const i = hash.indexOf('?');
    // Copie en lecture seule, recréée à chaque lecture depuis `hash` (réactif).
    // eslint-disable-next-line svelte/prefer-svelte-reactivity
    return new URLSearchParams(i >= 0 ? hash.slice(i + 1) : '');
  },
  navigate(to: string): void {
    window.location.hash = to.startsWith('#') ? to.slice(1) : to;
  },
};
