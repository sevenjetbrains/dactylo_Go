import type { ModuleDefinition } from '$app/types';
import PracticePage from './PracticePage.svelte';
import PracticeSettingsPanel from './PracticeSettingsPanel.svelte';
import fr from './i18n/fr.json';
import en from './i18n/en.json';

export default {
  id: 'practice',
  route: '/practice',
  labelKey: 'practice.menu',
  // clavier
  icon: 'M3 6h18v12H3zM7 10h.01M11 10h.01M15 10h.01M7 14h10',
  order: 10,
  page: PracticePage,
  settingsPanel: PracticeSettingsPanel,
  messages: { fr, en },
} satisfies ModuleDefinition;
