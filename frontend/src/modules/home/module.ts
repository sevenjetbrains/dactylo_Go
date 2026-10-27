import type { ModuleDefinition } from '$app/types';
import HomePage from './HomePage.svelte';
import fr from './i18n/fr.json';
import en from './i18n/en.json';

export default {
  id: 'home',
  route: '/',
  labelKey: 'home.menu',
  // maison
  icon: 'M3 10.5 12 3l9 7.5V20a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1z',
  order: 0,
  page: HomePage,
  messages: { fr, en },
} satisfies ModuleDefinition;
