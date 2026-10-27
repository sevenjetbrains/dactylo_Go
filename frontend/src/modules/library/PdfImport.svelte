<script lang="ts">
  import { onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { CleanImport, CreateText, OpenPDF } from '$wails/go/app/TextService';
  import type { app, text } from '$wails/go/models';
  import { loadPdf, pageLines, renderPage, type PdfDocument } from './pdf';
  import { formatPageRange, parsePageRange } from './pageRange';
  import TextForm from './TextForm.svelte';

  let { onDone, onCancel }: { onDone: (id: number) => void; onCancel: () => void } = $props();

  type Step = 'open' | 'pages' | 'lines' | 'review';
  let step: Step = $state('open');
  let error = $state('');
  let busy = $state(false);

  // Étape « pages »
  let fileName = $state('');
  let doc = $state.raw<PdfDocument | null>(null);
  let closeDoc: (() => Promise<void>) | null = null;
  let pagesInput = $state('1');
  let previewPage = $state(1);
  let canvas: HTMLCanvasElement | undefined = $state();
  const numPages = $derived(doc?.numPages ?? 0);
  const selectedPages = $derived(parsePageRange(pagesInput, numPages));

  // Étape « lignes » : lignes extraites, numérotées de 1 à n sur toutes les pages
  let extracted: { page: number; lines: string[] }[] = $state.raw([]);
  const allLines = $derived(extracted.flatMap((p) => p.lines.map((l) => ({ page: p.page, l }))));
  let firstLine = $state(1);
  let lastLine = $state(1);
  const lineCount = $derived(Math.max(0, lastLine - firstLine + 1));
  let options: text.CleanOptions = $state({
    removePageNumbers: true,
    removeRepeatedLines: true,
    joinHyphenated: true,
    normalizeApostrophes: true,
  });

  // Étape « vérification »
  let draft: Partial<app.TextInput> = $state({});

  async function open() {
    error = '';
    busy = true;
    try {
      const file = await OpenPDF($_('library.pdf.dialogTitle'));
      if (file.cancelled) return;
      // []byte arrive en base64 (le type généré dit number[])
      const loaded = await loadPdf(file.data as unknown as string);
      void closeDoc?.();
      doc = loaded.doc;
      closeDoc = loaded.close;
      fileName = file.name;
      pagesInput = numPages > 1 ? `1-${Math.min(numPages, 3)}` : '1';
      previewPage = 1;
      step = 'pages';
    } catch (e) {
      error = $_('library.pdf.openError', { values: { error: String(e) } });
    } finally {
      busy = false;
    }
  }

  // Aperçu de la page (F-21)
  $effect(() => {
    if (step !== 'pages' || !doc || !canvas) return;
    const page = Math.min(Math.max(1, previewPage), numPages);
    void renderPage(doc, page, canvas, 300).catch(() => {});
  });

  async function extract() {
    if (!doc || !selectedPages) return;
    busy = true;
    error = '';
    try {
      const pages = [];
      for (const n of selectedPages) pages.push({ page: n, lines: await pageLines(doc, n) });
      extracted = pages;
      const total = pages.reduce((s, p) => s + p.lines.filter(Boolean).length, 0);
      if (total === 0) {
        error = $_('library.pdf.scanned');
        return;
      }
      firstLine = 1;
      lastLine = allLines.length;
      step = 'lines';
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
    }
  }

  function setCount(n: number) {
    lastLine = Math.min(allLines.length, Math.max(firstLine, firstLine + n - 1));
  }

  async function clean() {
    busy = true;
    error = '';
    try {
      // On garde le regroupement par page : le nettoyage repère les en-têtes répétés.
      let i = 0;
      const pages = extracted.map((p) => ({
        number: p.page,
        lines: p.lines.filter(() => {
          i++;
          return i >= firstLine && i <= lastLine;
        }),
      }));
      const content = await CleanImport(pages, options);
      const range = formatPageRange(selectedPages ?? []);
      draft = {
        title: `${fileName.replace(/\.pdf$/i, '')} (p. ${range})`,
        content,
        source: 'pdf',
        sourceFile: fileName,
        pages: range,
        tags: [],
        language: 'fr',
      };
      step = 'review';
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
    }
  }

  async function save(input: app.TextInput) {
    onDone(await CreateText(input));
  }

  onDestroy(() => void closeDoc?.());
</script>

<section
  class="border-line bg-panel flex flex-col gap-5 rounded-lg border p-5"
  aria-label={$_('library.pdf.title')}
>
  <header class="flex items-center justify-between gap-4">
    <h2 class="font-semibold">
      {$_('library.pdf.title')}
      {#if fileName}<span class="text-muted font-normal"> · {fileName}</span>{/if}
    </h2>
    <button class="text-muted text-sm underline" onclick={onCancel}>{$_('library.cancel')}</button>
  </header>

  <ol class="text-muted flex gap-4 text-xs">
    {#each ['open', 'pages', 'lines', 'review'] as s, i (s)}
      <li class={step === s ? 'text-ink font-semibold' : ''}>
        {i + 1}. {$_(`library.pdf.steps.${s}`)}
      </li>
    {/each}
  </ol>

  {#if error}<p role="alert" class="text-error text-sm">{error}</p>{/if}

  {#if step === 'open'}
    <p class="text-sm">{$_('library.pdf.intro')}</p>
    <div>
      <button
        class="bg-accent text-accent-ink rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
        onclick={open}
        disabled={busy}
      >
        {$_('library.pdf.choose')}
      </button>
    </div>
  {:else if step === 'pages'}
    <div class="flex flex-wrap gap-6">
      <div class="flex flex-col gap-2">
        <canvas bind:this={canvas} class="border-line rounded border bg-white"></canvas>
        <div class="flex items-center gap-2 text-sm">
          <button
            class="border-line rounded border px-2 disabled:opacity-40"
            onclick={() => previewPage--}
            disabled={previewPage <= 1}
            aria-label={$_('library.pdf.prevPage')}>‹</button
          >
          <span>{$_('library.pdf.pageOf', { values: { n: previewPage, total: numPages } })}</span>
          <button
            class="border-line rounded border px-2 disabled:opacity-40"
            onclick={() => previewPage++}
            disabled={previewPage >= numPages}
            aria-label={$_('library.pdf.nextPage')}>›</button
          >
        </div>
      </div>
      <div class="flex min-w-60 flex-1 flex-col gap-3">
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">{$_('library.pdf.pages', { values: { total: numPages } })}</span>
          <input
            class="border-line bg-surface w-48 rounded-md border px-3 py-1.5"
            bind:value={pagesInput}
            aria-invalid={selectedPages === null}
          />
        </label>
        {#if selectedPages === null}
          <p class="text-error text-sm">{$_('library.pdf.badRange')}</p>
        {/if}
        <div>
          <button
            class="bg-accent text-accent-ink rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            onclick={extract}
            disabled={busy || selectedPages === null}
          >
            {$_('library.pdf.extract')}
          </button>
        </div>
      </div>
    </div>
  {:else if step === 'lines'}
    <div class="flex flex-wrap items-end gap-4 text-sm">
      <label class="flex flex-col gap-1">
        <span class="text-muted">{$_('library.pdf.firstLine')}</span>
        <input
          type="number"
          min="1"
          max={allLines.length}
          class="border-line bg-surface w-24 rounded-md border px-2 py-1"
          bind:value={firstLine}
          onchange={() => (lastLine = Math.max(lastLine, firstLine))}
        />
      </label>
      <label class="flex flex-col gap-1">
        <span class="text-muted">{$_('library.pdf.lastLine')}</span>
        <input
          type="number"
          min={firstLine}
          max={allLines.length}
          class="border-line bg-surface w-24 rounded-md border px-2 py-1"
          bind:value={lastLine}
        />
      </label>
      <label class="flex flex-col gap-1">
        <span class="text-muted">{$_('library.pdf.lineCount')}</span>
        <input
          type="number"
          min="1"
          max={allLines.length}
          class="border-line bg-surface w-24 rounded-md border px-2 py-1"
          value={lineCount}
          onchange={(e) => setCount(Number(e.currentTarget.value))}
        />
      </label>
      <span class="text-muted">
        {$_('library.pdf.selected', { values: { n: lineCount, total: allLines.length } })}
      </span>
    </div>

    <ol
      class="border-line bg-surface max-h-72 overflow-y-auto rounded-md border p-2 font-mono text-xs"
      data-testid="pdf-lines"
    >
      {#each allLines as line, i (i)}
        {@const inSel = i + 1 >= firstLine && i + 1 <= lastLine}
        {#if i === 0 || allLines[i - 1]?.page !== line.page}
          <li class="text-muted mt-2 first:mt-0">
            — {$_('library.pdf.page', { values: { n: line.page } })} —
          </li>
        {/if}
        <li class={['flex gap-3', !inSel && 'opacity-35']}>
          <span class="text-muted w-8 shrink-0 text-right tabular-nums">{i + 1}</span>
          <span class="whitespace-pre-wrap">{line.l}</span>
        </li>
      {/each}
    </ol>

    <fieldset class="flex flex-wrap gap-x-6 gap-y-2 text-sm">
      <legend class="mb-1 font-semibold">{$_('library.pdf.cleaning')}</legend>
      {#each ['removePageNumbers', 'removeRepeatedLines', 'joinHyphenated', 'normalizeApostrophes'] as const as opt (opt)}
        <label class="flex items-center gap-2">
          <input type="checkbox" bind:checked={options[opt]} />
          {$_(`library.pdf.options.${opt}`)}
        </label>
      {/each}
    </fieldset>

    <div class="flex gap-3">
      <button
        class="bg-accent text-accent-ink rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
        onclick={clean}
        disabled={busy || lineCount === 0}
      >
        {$_('library.pdf.clean')}
      </button>
      <button
        class="border-line rounded-md border px-4 py-2 text-sm"
        onclick={() => (step = 'pages')}
      >
        {$_('library.pdf.back')}
      </button>
    </div>
  {:else}
    <p class="text-muted text-sm">{$_('library.pdf.reviewHint')}</p>
    <TextForm
      heading={$_('library.pdf.steps.review')}
      initial={draft}
      submitLabel={$_('library.form.save')}
      onSubmit={save}
      onCancel={() => (step = 'lines')}
    />
  {/if}
</section>
