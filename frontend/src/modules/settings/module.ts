import type { ModuleDefinition } from '$app/types';
import SettingsPage from './SettingsPage.svelte';
import AppearancePanel from './AppearancePanel.svelte';
import fr from './i18n/fr.json';
import en from './i18n/en.json';

export default {
  id: 'settings',
  route: '/settings',
  labelKey: 'settings.menu',
  // curseurs de réglage
  icon: 'M4 6h10M18 6h2M4 12h4M12 12h8M4 18h12M20 18h0M16 4v4M10 10v4M16 16v4',
  order: 100,
  page: SettingsPage,
  settingsPanel: AppearancePanel,
  messages: { fr, en },
} satisfies ModuleDefinition;
