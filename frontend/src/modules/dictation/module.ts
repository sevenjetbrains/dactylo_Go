import type { ModuleDefinition } from '$app/types';
import DictationPage from './DictationPage.svelte';
import fr from './i18n/fr.json';
import en from './i18n/en.json';

export default {
  id: 'dictation',
  route: '/dictation',
  labelKey: 'dictation.menu',
  // haut-parleur
  icon: 'M11 5 6 9H2v6h4l5 4zM15.5 8.5a5 5 0 0 1 0 7M19 5a10 10 0 0 1 0 14',
  order: 15,
  page: DictationPage,
  messages: { fr, en },
} satisfies ModuleDefinition;
