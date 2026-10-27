import { normalizePath } from './registry';

// Routage par hash : compatible avec le serveur d'assets embarqué de Wails.
let current = $state(normalizePath(window.location.hash));

window.addEventListener('hashchange', () => {
  current = normalizePath(window.location.hash);
});

export const router = {
  get path(): string {
    return current;
  },
  navigate(to: string): void {
    window.location.hash = normalizePath(to);
  },
};
