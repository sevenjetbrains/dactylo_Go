import type { ModuleDefinition } from './types';

const ID_PATTERN = /^[a-z][a-z0-9-]*$/;

export interface Registry {
  readonly modules: readonly ModuleDefinition[];
  readonly menu: readonly ModuleDefinition[];
  readonly settingsPanels: readonly ModuleDefinition[];
  resolve(path: string): ModuleDefinition | undefined;
}

/** Valide les déclarations et construit le registre. Lève une erreur au démarrage si deux modules entrent en conflit. */
export function createRegistry(defs: readonly ModuleDefinition[]): Registry {
  const ids = new Set<string>();
  const routes = new Set<string>();
  for (const d of defs) {
    if (!ID_PATTERN.test(d.id)) throw new Error(`Identifiant de module invalide : "${d.id}"`);
    if (!d.route.startsWith('/')) throw new Error(`Route invalide pour "${d.id}" : "${d.route}"`);
    if (ids.has(d.id)) throw new Error(`Module en double : "${d.id}"`);
    if (routes.has(d.route)) throw new Error(`Route en double : "${d.route}"`);
    ids.add(d.id);
    routes.add(d.route);
  }

  const modules = [...defs].sort((a, b) => a.order - b.order);
  const byRoute = new Map(modules.map((m) => [m.route, m]));

  return {
    modules,
    menu: modules.filter((m) => !m.hidden),
    settingsPanels: modules.filter((m) => m.settingsPanel),
    resolve(path) {
      return byRoute.get(normalizePath(path));
    },
  };
}

/** Chemin de route normalisé : sans #, sans requête, sans / final. */
export function normalizePath(path: string): string {
  const p = path.replace(/^#/, '').split('?')[0] ?? '';
  if (p === '' || p === '/') return '/';
  return ('/' + p.replace(/^\/+/, '')).replace(/\/+$/, '');
}
