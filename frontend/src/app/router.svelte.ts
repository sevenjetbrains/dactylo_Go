import { normalizePath } from './registry';

// Routage par hash : compatible avec le serveur d'assets embarqué de Wails.
let hash = $state(window.location.hash);

/** Renvoie le message à afficher si quitter la page doit être confirmé, sinon null. */
export type LeaveGuard = () => string | null;

// Lu seulement au moment d'une navigation : pas besoin de réactivité.
// eslint-disable-next-line svelte/prefer-svelte-reactivity
const guards = new Set<LeaveGuard>();
// Navigation retenue, en attente de confirmation : adresse visée et message.
let pending: { to: string; message: string } | null = $state(null);

/** Message du premier garde qui retient la navigation, sinon null. */
function blocked(): string | null {
  for (const g of guards) {
    const m = g();
    if (m) return m;
  }
  return null;
}

window.addEventListener('hashchange', () => {
  const to = window.location.hash;
  if (to === hash) return;
  const message = blocked();
  if (message) {
    // la page reste affichée : on remet l'ancienne adresse sans nouvel événement
    history.replaceState(history.state, '', hash || '#/');
    pending = { to, message };
    return;
  }
  hash = to;
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
  /**
   * Retient toute navigation tant que `guard` renvoie un message (session en
   * cours) : l'application demande confirmation. Renvoie de quoi le retirer.
   */
  guard(guard: LeaveGuard): () => void {
    guards.add(guard);
    return () => guards.delete(guard);
  },
  /** Navigation en attente de confirmation, ou null. */
  get pending(): { to: string; message: string } | null {
    return pending;
  },
  /** Confirme la navigation retenue. */
  leave(): void {
    if (!pending) return;
    const { to } = pending;
    pending = null;
    history.replaceState(history.state, '', to);
    hash = to;
  },
  /** Annule la navigation retenue : on reste sur la page. */
  stay(): void {
    pending = null;
  },
};
