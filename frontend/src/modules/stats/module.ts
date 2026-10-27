import type { ModuleDefinition } from '$app/types';
import StatsPage from './StatsPage.svelte';
import fr from './i18n/fr.json';
import en from './i18n/en.json';

export default {
  id: 'stats',
  route: '/stats',
  labelKey: 'stats.menu',
  // courbe montante
  icon: 'M3 3v18h18M7 15l4-4 3 3 5-6',
  order: 40,
  page: StatsPage,
  messages: { fr, en },
} satisfies ModuleDefinition;
