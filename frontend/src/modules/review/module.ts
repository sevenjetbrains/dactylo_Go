import type { ModuleDefinition } from '$app/types';
import ReviewPage from './ReviewPage.svelte';
import fr from './i18n/fr.json';
import en from './i18n/en.json';

export default {
  id: 'review',
  route: '/review',
  labelKey: 'review.menu',
  // flèches circulaires (répétition)
  icon: 'M3 12a9 9 0 0 1 15.5-6.2L21 8M21 3v5h-5M21 12a9 9 0 0 1-15.5 6.2L3 16M3 21v-5h5',
  order: 30,
  page: ReviewPage,
  messages: { fr, en },
} satisfies ModuleDefinition;
