<script lang="ts">
  import { onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { CleanImport, CountExercises, CreateText, OpenPDF } from '$wails/go/app/TextService';
  import type { text } from '$wails/go/models';
  import { typingSettings } from '$shared/typing/settings.svelte';
  import { loadPdf, pageLines, renderPage, type PdfDocument } from './pdf';
  import { formatPageRange, parsePageRange } from './pageRange';
  import { highlightWords, thumbWindow } from './highlight';

  let { onDone, onCancel }: { onDone: (id: number) => void; onCancel: () => void } = $props();

  let error = $state('');
  let busy = $state(false);

  // Document
  let fileName = $state('');
  let doc = $state.raw<PdfDocument | null>(null);
  let closeDoc: (() => Promise<void>) | null = null;
  let hasText = $state(true);
  const numPages = $derived(doc?.numPages ?? 0);

  // Choix
  let pagesInput = $state('1');
  const selected = $derived(parsePageRange(pagesInput, numPages));
  let firstLine = $state(1);
  let lastLine = $state(0); // 0 = jusqu'à la fin
  let split = $state(typingSettings.exerciseSplit);
  let splitSize = $state(typingSettings.exerciseSize);
  let tags = $state('');
  let title = $state('');
  let joinHyphenated = $state(true);
  let removeEdges = $state(true); // numéros de page + en-têtes et pieds répétés
  let normalize = $state(true);

  // Résultat
  let totalLines = $state(0);
  let report = $state.raw<text.CleanReport | null>(null);
  let exercises = $state(0);
  let edited: string | null = $state(null);
  const content = $derived(edited ?? report?.text ?? '');

  // cache des lignes extraites par page : pas d'état réactif
  // eslint-disable-next-line svelte/prefer-svelte-reactivity
  const pageCache = new Map<number, string[]>();
  async function linesOf(n: number): Promise<string[]> {
    if (!doc) return [];
    if (!pageCache.has(n)) pageCache.set(n, await pageLines(doc, n));
    return pageCache.get(n)!;
  }

  async function open() {
    error = '';
    busy = true;
    try {
      const file = await OpenPDF($_('library.pdf.dialogTitle'));
      if (file.cancelled) return;
      // []byte arrive en base64 (le type généré dit number[])
      const loaded = await loadPdf(file.data as unknown as string);
      void closeDoc?.();
      pageCache.clear();
      doc = loaded.doc;
      closeDoc = loaded.close;
      fileName = file.name;
      title = file.name.replace(/\.pdf$/i, '');
      pagesInput = numPages > 1 ? `1-${Math.min(numPages, 3)}` : '1';
      // texte détecté sur l'une des premières pages ?
      hasText = false;
      for (let p = 1; p <= Math.min(3, numPages) && !hasText; p++) {
        hasText = (await linesOf(p)).some(Boolean);
      }
    } catch (e) {
      error = $_('library.pdf.openError', { values: { error: String(e) } });
    } finally {
      busy = false;
    }
  }

  // Nouvelle plage de pages : toutes les lignes, modifications perdues.
  $effect(() => {
    void selected;
    firstLine = 1;
    lastLine = 0;
    edited = null;
  });

  // Extraction + nettoyage + nombre d'exercices, recalculés à chaque réglage.
  $effect(() => {
    const pages = selected;
    const opts = {
      joinHyphenated,
      removePageNumbers: removeEdges,
      removeRepeatedLines: removeEdges,
      normalizeApostrophes: normalize,
    };
    const [from, to, sp, size] = [firstLine, lastLine, split, splitSize];
    if (!doc || !pages) return;
    const timer = setTimeout(async () => {
      try {
        const extracted = [];
        for (const n of pages) extracted.push({ number: n, lines: await linesOf(n) });
        totalLines = extracted.reduce((s, p) => s + p.lines.length, 0);
        const end = to || totalLines;
        let i = 0;
        const chosen = extracted.map((p) => ({
          number: p.number,
          lines: p.lines.filter(() => {
            i++;
            return i >= from && i <= end;
          }),
        }));
        report = await CleanImport(chosen, opts);
        exercises = report.text ? await CountExercises(report.text, sp, size) : 0;
        error = report.text ? '' : $_('library.pdf.scanned');
      } catch (e) {
        error = String(e);
      }
    }, 250);
    return () => clearTimeout(timer);
  });

  // Miniatures
  const thumbs = $derived(selected && numPages ? thumbWindow(selected[0]!, numPages) : []);
  function thumb(canvas: HTMLCanvasElement, page: number) {
    if (doc) void renderPage(doc, page, canvas, 44).catch(() => {});
  }
  function pickPage(p: number) {
    const s = selected;
    if (s && s.length === 1 && s[0] !== p) {
      pagesInput = `${Math.min(s[0]!, p)}-${Math.max(s[0]!, p)}`;
    } else {
      pagesInput = String(p);
    }
  }

  async function save() {
    busy = true;
    try {
      const id = await CreateText({
        title,
        content,
        language: 'fr',
        tags: tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
        source: 'pdf',
        sourceFile: fileName,
        pages: formatPageRange(selected ?? []),
        split,
        splitSize: split === 'none' ? 0 : splitSize,
      });
      onDone(id);
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
    }
  }

  onDestroy(() => void closeDoc?.());
</script>

<section class="flex flex-col gap-5" aria-label={$_('library.pdf.title')}>
  <header class="flex items-center justify-between">
    <h2 class="text-lg font-bold">{$_('library.pdf.title')}</h2>
    <button
      class="text-muted hover:text-ink text-xl leading-none"
      onclick={onCancel}
      aria-label={$_('library.close')}>×</button
    >
  </header>

  {#if !doc}
    <p class="text-muted text-sm">{$_('library.pdf.intro')}</p>
    <button class="btn btn-primary" onclick={open} disabled={busy}
      >{$_('library.pdf.choose')}</button
    >
    {#if error}<p role="alert" class="text-error text-sm">{error}</p>{/if}
  {:else}
    <div class="bg-sunken flex items-center gap-3 rounded-xl p-3">
      <span class="badge bg-error text-white">PDF</span>
      <div class="min-w-0">
        <p class="truncate text-sm font-bold">{fileName}</p>
        <p class="text-muted text-xs">
          {$_('library.pdf.pagesCount', { values: { n: numPages } })} ·
          {hasText ? $_('library.pdf.textFound') : $_('library.pdf.noText')}
        </p>
      </div>
    </div>

    <div>
      <div class="flex items-center justify-between">
        <span class="text-sm font-semibold">{$_('library.pdf.pagesLabel')}</span>
        <label class="text-muted flex items-center gap-2 text-sm">
          {$_('library.pdf.range')}
          <input
            class="field w-24 py-1"
            bind:value={pagesInput}
            aria-invalid={selected === null}
            aria-label={$_('library.pdf.rangeLabel', { values: { total: numPages } })}
          />
        </label>
      </div>
      {#if selected === null}
        <p class="text-error mt-1 text-sm">{$_('library.pdf.badRange')}</p>
      {/if}
      <ol class="mt-3 flex gap-2" data-testid="thumbs">
        {#each thumbs as p (p)}
          {@const on = selected?.includes(p)}
          <li>
            <button
              class="flex flex-col items-center gap-1"
              onclick={() => pickPage(p)}
              aria-pressed={on}
              aria-label={$_('library.pdf.page', { values: { n: p } })}
            >
              <canvas
                {@attach (c) => thumb(c, p)}
                class={['rounded border-2 bg-white', on ? 'border-accent' : 'border-line']}
              ></canvas>
              <span class={['text-xs', on ? 'text-accent font-bold' : 'text-muted']}>{p}</span>
            </button>
          </li>
        {/each}
      </ol>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <label class="flex flex-col gap-1.5 text-sm">
        <span class="font-semibold">{$_('library.split.label')}</span>
        <select class="field" bind:value={split}>
          <option value="none">{$_('library.split.none')}</option>
          <option value="paragraph">{$_('library.split.paragraph')}</option>
          <option value="lines">{$_('library.split.lines')}</option>
        </select>
      </label>
      <label class="flex flex-col gap-1.5 text-sm">
        <span class="font-semibold">{$_('library.form.tagsShort')}</span>
        <input class="field" bind:value={tags} placeholder={$_('library.form.tagsPlaceholder')} />
      </label>
    </div>

    <fieldset class="flex flex-col gap-2 text-sm">
      <legend class="mb-1 font-semibold">{$_('library.pdf.cleaning')}</legend>
      <label class="flex items-center gap-2">
        <input type="checkbox" bind:checked={joinHyphenated} />
        {$_('library.pdf.options.joinHyphenated')}
        {#if report && joinHyphenated}<span class="text-muted">({report.joined.length})</span>{/if}
      </label>
      <label class="flex items-center gap-2">
        <input type="checkbox" bind:checked={removeEdges} />
        {$_('library.pdf.options.removeEdges')}
        {#if report && removeEdges}<span class="text-muted">({report.removedLines})</span>{/if}
      </label>
      <label class="flex items-center gap-2">
        <input type="checkbox" bind:checked={normalize} />
        {$_('library.pdf.options.normalize')}
      </label>
    </fieldset>

    <details class="text-sm">
      <summary class="text-accent cursor-pointer font-semibold">{$_('library.pdf.lines')}</summary>
      <div class="mt-2 flex flex-wrap items-end gap-3">
        <label class="flex flex-col gap-1">
          <span class="text-muted">{$_('library.pdf.firstLine')}</span>
          <input
            class="field w-20 py-1"
            type="number"
            min="1"
            max={totalLines}
            bind:value={firstLine}
          />
        </label>
        <label class="flex flex-col gap-1">
          <span class="text-muted">{$_('library.pdf.lastLine')}</span>
          <input
            class="field w-20 py-1"
            type="number"
            min={firstLine}
            max={totalLines}
            value={lastLine || totalLines}
            onchange={(e) => (lastLine = Number(e.currentTarget.value))}
          />
        </label>
        <span class="text-muted">{$_('library.pdf.totalLines', { values: { n: totalLines } })}</span
        >
      </div>
    </details>

    <label class="flex flex-col gap-1.5 text-sm">
      <span class="font-semibold">{$_('library.form.title')}</span>
      <input class="field" bind:value={title} />
    </label>

    <div>
      <div class="flex items-center justify-between">
        <span class="text-sm font-semibold">{$_('library.pdf.preview')}</span>
        {#if report?.text}
          <button
            class="text-accent text-xs font-semibold"
            onclick={() => (edited = edited === null ? content : null)}
          >
            {edited === null ? $_('library.pdf.edit') : $_('library.pdf.reset')}
          </button>
        {/if}
      </div>
      {#if edited !== null}
        <textarea
          class="field bg-sunken mt-2 min-h-40 w-full font-mono text-sm leading-relaxed"
          bind:value={edited}
          aria-label={$_('library.pdf.preview')}></textarea>
      {:else}
        <div
          class="border-line bg-sunken mt-2 max-h-48 overflow-y-auto rounded-xl border p-4 font-mono text-sm leading-relaxed whitespace-pre-wrap"
          data-testid="pdf-preview"
        >
          {#each highlightWords(report?.text ?? '', report?.joined ?? []) as piece, i (i)}
            {#if piece.hit}<mark class="bg-ok-soft text-ok rounded px-0.5">{piece.text}</mark
              >{:else}{piece.text}{/if}
          {/each}
        </div>
        {#if report?.joined.length}
          <p class="text-muted mt-1 text-xs">{$_('library.pdf.joinedHint')}</p>
        {/if}
      {/if}
    </div>

    {#if error}<p role="alert" class="text-error text-sm">{error}</p>{/if}

    <button class="btn btn-primary w-full py-3" onclick={save} disabled={busy || !content.trim()}>
      {$_('library.pdf.save', {
        values: { pages: selected?.length ?? 0, exercises },
      })}
    </button>
  {/if}
</section>
