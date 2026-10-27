import type { ModuleDefinition } from '$app/types';
import LibraryPage from './LibraryPage.svelte';
import fr from './i18n/fr.json';
import en from './i18n/en.json';

export default {
  id: 'library',
  route: '/library',
  labelKey: 'library.menu',
  // livre ouvert
  icon: 'M2 5h7a3 3 0 0 1 3 3v12a2 2 0 0 0-2-2H2zM22 5h-7a3 3 0 0 0-3 3v12a2 2 0 0 1 2-2h8z',
  order: 20,
  page: LibraryPage,
  messages: { fr, en },
} satisfies ModuleDefinition;
