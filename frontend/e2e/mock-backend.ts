// Simule window.go (les liaisons Wails) dans le navigateur de test.
// Script injecté tel quel dans la page : pas d'import possible ici.
export function installMockBackend(initialTexts: { title: string; content: string }[]): void {
  type Ks = { seq: number; position: number; typed: string; timeMs: number };
  const texts = initialTexts.map((t, i) => ({
    id: i + 1,
    title: t.title,
    content: t.content,
    language: 'fr',
    source: 'paste',
    tags: [] as string[],
    createdAt: '2026-10-27T10:00:00.000Z',
  }));
  const sessions = new Map<number, { textId: number; keys: Ks[] }>();
  let nextSession = 1;
  const calls: Record<string, unknown[]> = {};
  const log = (name: string, args: unknown) => (calls[name] ??= []).push(args);
  (window as unknown as { __calls: typeof calls }).__calls = calls;

  const summary = (t: { title: string; content: string; id: number }) => ({
    id: t.id,
    title: t.title,
    language: 'fr',
    source: 'paste',
    tags: [],
    length: [...t.content].length,
    preview: t.content.slice(0, 120),
    createdAt: '2026-10-27T10:00:00.000Z',
  });

  (window as unknown as { go: unknown }).go = {
    app: {
      SystemService: {
        Info: async () => ({
          version: 'e2e',
          platform: 'test',
          databasePath: ':memory:',
          schemaVersion: 2,
        }),
      },
      TextService: {
        ListTexts: async () => texts.map(summary),
        GetText: async (id: number) => {
          const t = texts.find((x) => x.id === id);
          if (!t) throw new Error('introuvable');
          return t;
        },
        CreateText: async (input: { title: string; content: string }) => {
          const id = texts.length + 1;
          texts.unshift({
            ...texts[0]!,
            id,
            title: input.title || 'Sans titre',
            content: input.content,
          });
          log('CreateText', input);
          return id;
        },
        DeleteText: async (id: number) => {
          texts.splice(
            texts.findIndex((t) => t.id === id),
            1,
          );
        },
      },
      PracticeService: {
        StartSession: async (textId: number, settings: string) => {
          log('StartSession', { textId, settings });
          sessions.set(nextSession, { textId, keys: [] });
          return nextSession++;
        },
        RecordKeystrokes: async (id: number, ks: Ks[]) => {
          log('RecordKeystrokes', ks);
          sessions.get(id)!.keys.push(...ks);
        },
        FinishSession: async (id: number, durationMs: number) => {
          const s = sessions.get(id)!;
          const text = [...texts.find((t) => t.id === s.textId)!.content];
          const chars = s.keys.filter((k) => k.typed !== '\b');
          const ok = chars.filter((k) => text[k.position] === k.typed).length;
          log('FinishSession', { id, durationMs, keys: s.keys });
          return {
            sessionId: id,
            textId: s.textId,
            summary: {
              durationMs,
              keystrokes: chars.length,
              correctKeystrokes: ok,
              backspaces: s.keys.length - chars.length,
              correctChars: ok,
              uncorrectedErrors: chars.length - ok,
              wpmGross: 42,
              wpmNet: 40,
              cpm: 200,
              accuracy: (100 * ok) / Math.max(1, chars.length),
            },
            mistakes:
              ok < chars.length
                ? [{ expected: 'être', typed: 'etre', position: 8, type: 'accent' }]
                : [],
            average: { sessions: 0, wpmNet: 0, accuracy: 0 },
          };
        },
      },
    },
  };
}
