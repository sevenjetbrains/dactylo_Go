import '@testing-library/jest-dom/vitest';
import type { Dictionary, ModuleDefinition } from '$app/types';
import { setupI18n } from '$i18n/index';

// Charge uniquement les traductions des modules (sans importer leurs pages,
// pour que chaque test puisse simuler les liaisons Wails avec vi.mock).
const files = import.meta.glob<Dictionary>('../modules/*/i18n/*.json', {
  eager: true,
  import: 'default',
});
const byModule = new Map<string, Pick<ModuleDefinition, 'id' | 'messages'>>();
for (const [path, dict] of Object.entries(files)) {
  const [, id, lang] = /modules\/([^/]+)\/i18n\/([^/]+)\.json$/.exec(path) ?? [];
  if (!id || !lang) continue;
  const m = byModule.get(id) ?? { id, messages: {} };
  m.messages = { ...m.messages, [lang]: dict };
  byModule.set(id, m);
}

setupI18n([...byModule.values()], 'fr');
