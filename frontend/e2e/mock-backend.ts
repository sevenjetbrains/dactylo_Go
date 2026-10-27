// Simule window.go (les liaisons Wails) dans le navigateur de test.
// Script injecté tel quel dans la page : pas d'import possible ici.
export function installMockBackend(args: {
  texts: { title: string; content: string }[];
  pdf?: string;
}): void {
  type Ks = { seq: number; position: number; typed: string; timeMs: number };
  type Txt = {
    id: number;
    title: string;
    content: string;
    language: string;
    source: string;
    sourceFile: string;
    pages: string;
    tags: string[];
    createdAt: string;
  };
  const texts: Txt[] = args.texts.map((t, i) => ({
    id: i + 1,
    title: t.title,
    content: t.content,
    language: 'fr',
    source: 'paste',
    sourceFile: '',
    pages: '',
    tags: [],
    createdAt: '2026-10-27T10:00:00.000Z',
  }));
  const sessions = new Map<number, { textId: number; start: number; keys: Ks[] }>();
  let nextSession = 1;
  let nextText = texts.length + 1;
  const calls: Record<string, unknown[]> = {};
  const log = (name: string, a: unknown) => (calls[name] ??= []).push(a);
  (window as unknown as { __calls: typeof calls }).__calls = calls;

  const summary = (t: Txt) => ({
    ...t,
    length: [...t.content].length,
    preview: t.content.slice(0, 120),
    position: 0,
  });

  (window as unknown as { go: unknown }).go = {
    app: {
      SystemService: {
        Info: async () => ({
          version: 'e2e',
          platform: 'test',
          databasePath: ':memory:',
          schemaVersion: 3,
        }),
      },
      TextService: {
        ListTexts: async (q: { search: string }) => {
          const all = [...texts].reverse().map(summary);
          const filtered = all.filter((t) =>
            (t.title + t.content).toLowerCase().includes((q.search ?? '').toLowerCase()),
          );
          return { texts: filtered, tags: [], total: all.length };
        },
        GetText: async (id: number) => {
          const t = texts.find((x) => x.id === id);
          if (!t) throw new Error('introuvable');
          return t;
        },
        CreateText: async (input: Partial<Txt>) => {
          log('CreateText', input);
          const id = nextText++;
          texts.push({
            id,
            title: input.title || 'Sans titre',
            content: input.content ?? '',
            language: 'fr',
            source: input.source || 'paste',
            sourceFile: input.sourceFile ?? '',
            pages: input.pages ?? '',
            tags: input.tags ?? [],
            createdAt: '2026-10-27T11:00:00.000Z',
          });
          return id;
        },
        UpdateText: async (id: number, input: Partial<Txt>) => {
          Object.assign(
            texts.find((t) => t.id === id)!,
            input,
          );
        },
        DuplicateText: async (id: number) => {
          const t = texts.find((x) => x.id === id)!;
          const copy = { ...t, id: nextText++, title: t.title + ' (copie)' };
          texts.push(copy);
          return copy.id;
        },
        DeleteText: async (id: number) => {
          texts.splice(
            texts.findIndex((t) => t.id === id),
            1,
          );
        },
        Exercises: async (id: number) => {
          const n = [...texts.find((t) => t.id === id)!.content].length;
          const seg = { index: 0, start: 0, end: n };
          return { segments: [seg], length: n, position: 0, resume: seg, resumeStart: 0 };
        },
        ResetProgress: async () => {},
        OpenPDF: async () => ({ cancelled: false, name: 'livre.pdf', data: args.pdf ?? '' }),
        CleanImport: async (pages: { number: number; lines: string[] }[], opts: unknown) => {
          log('CleanImport', { pages, opts });
          return pages.flatMap((p) => p.lines).join(' ');
        },
      },
      ReviewService: {
        Summary: async () => ({ due: 0, learning: 0, mastered: 0 }),
        ListWords: async () => [],
      },
      StatsService: {
        Dashboard: async () => ({
          points: [],
          totals: { sessions: 0, totalMs: 0, bestWpm: 0 },
          streak: 0,
          keys: [],
          bigrams: [],
        }),
      },
      PracticeService: {
        StartSession: async (req: { textId: number; start: number; end: number }) => {
          log('StartSession', req);
          sessions.set(nextSession, { textId: req.textId, start: req.start, keys: [] });
          return nextSession++;
        },
        RecordKeystrokes: async (id: number, ks: Ks[]) => {
          log('RecordKeystrokes', ks);
          sessions.get(id)!.keys.push(...ks);
        },
        FinishSession: async (id: number, durationMs: number) => {
          const s = sessions.get(id)!;
          const text = [...texts.find((t) => t.id === s.textId)!.content].slice(s.start);
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
