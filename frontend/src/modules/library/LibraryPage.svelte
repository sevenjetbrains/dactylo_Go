<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import {
    CreateText,
    DeleteText,
    DuplicateText,
    GetText,
    ListTexts,
    UpdateText,
  } from '$wails/go/app/TextService';
  import type { app, store } from '$wails/go/models';
  import { formatNumber } from '$shared/format';
  import PdfImport from './PdfImport.svelte';
  import TextForm from './TextForm.svelte';

  type Panel =
    { kind: 'none' } | { kind: 'new' } | { kind: 'pdf' } | { kind: 'edit'; text: store.Text };
  const SORTS = ['recent', 'oldest', 'title', 'length', 'score'] as const;

  let list: app.TextList | null = $state(null);
  let loadError = $state('');
  let search = $state('');
  let tag = $state('');
  let sort: (typeof SORTS)[number] = $state('recent');
  let panel: Panel = $state({ kind: 'none' });
  let confirmDelete: number | null = $state(null);
  let highlight: number | null = $state(null);

  const lang = $derived($locale ?? 'fr');

  async function refresh() {
    try {
      list = await ListTexts({ search, tag, sort });
      loadError = '';
    } catch (e) {
      loadError = String(e);
    }
  }

  // Recherche à la frappe, avec un court délai.
  $effect(() => {
    void [search, tag, sort];
    const t = setTimeout(refresh, 150);
    return () => clearTimeout(t);
  });

  async function saved(id: number) {
    panel = { kind: 'none' };
    highlight = id;
    await refresh();
  }

  async function create(input: app.TextInput) {
    await saved(await CreateText(input));
  }

  async function edit(id: number) {
    panel = { kind: 'edit', text: await GetText(id) };
  }

  async function update(id: number, input: app.TextInput) {
    await UpdateText(id, input);
    await saved(id);
  }

  async function duplicate(id: number) {
    await saved(await DuplicateText(id));
  }

  async function remove(id: number) {
    if (confirmDelete !== id) {
      confirmDelete = id;
      return;
    }
    confirmDelete = null;
    await DeleteText(id);
    await refresh();
  }

  const progress = (t: store.TextSummary) =>
    t.length > 0 ? Math.min(100, Math.round((100 * t.position) / t.length)) : 0;
</script>

<section class="mx-auto max-w-4xl px-8 py-10">
  <div class="flex flex-wrap items-center justify-between gap-4">
    <h1 class="text-3xl font-semibold">{$_('library.title')}</h1>
    <div class="flex gap-2">
      <button
        class="border-line bg-panel rounded-md border px-3 py-1.5 text-sm font-medium"
        onclick={() => (panel = { kind: 'new' })}
      >
        {$_('library.form.heading')}
      </button>
      <button
        class="border-line bg-panel rounded-md border px-3 py-1.5 text-sm font-medium"
        onclick={() => (panel = { kind: 'pdf' })}
      >
        {$_('library.pdf.title')}
      </button>
    </div>
  </div>

  <div class="mt-6">
    {#if panel.kind === 'new'}
      <TextForm
        heading={$_('library.form.heading')}
        submitLabel={$_('library.form.save')}
        onSubmit={create}
        onCancel={() => (panel = { kind: 'none' })}
      />
    {:else if panel.kind === 'edit'}
      {@const t = panel.text}
      {#key t.id}
        <TextForm
          heading={$_('library.editHeading', { values: { title: t.title } })}
          initial={t}
          submitLabel={$_('library.form.update')}
          onSubmit={(input) => update(t.id, input)}
          onCancel={() => (panel = { kind: 'none' })}
        />
      {/key}
    {:else if panel.kind === 'pdf'}
      <PdfImport onDone={saved} onCancel={() => (panel = { kind: 'none' })} />
    {/if}
  </div>

  <div class="mt-8 flex flex-wrap items-end gap-3" role="search">
    <label class="flex min-w-56 flex-1 flex-col gap-1 text-sm">
      <span class="text-muted">{$_('library.search')}</span>
      <input
        type="search"
        class="border-line bg-panel rounded-md border px-3 py-1.5"
        bind:value={search}
        placeholder={$_('library.searchPlaceholder')}
      />
    </label>
    <label class="flex flex-col gap-1 text-sm">
      <span class="text-muted">{$_('library.tag')}</span>
      <select class="border-line bg-panel rounded-md border px-3 py-1.5" bind:value={tag}>
        <option value="">{$_('library.allTags')}</option>
        {#each list?.tags ?? [] as t (t)}
          <option value={t}>{t}</option>
        {/each}
      </select>
    </label>
    <label class="flex flex-col gap-1 text-sm">
      <span class="text-muted">{$_('library.sort')}</span>
      <select class="border-line bg-panel rounded-md border px-3 py-1.5" bind:value={sort}>
        {#each SORTS as s (s)}
          <option value={s}>{$_(`library.sorts.${s}`)}</option>
        {/each}
      </select>
    </label>
  </div>

  {#if list}
    <h2 class="mt-6 mb-3 font-semibold">
      {list.texts.length === list.total
        ? $_('library.list', { values: { n: list.total } })
        : $_('library.filtered', { values: { n: list.texts.length, total: list.total } })}
    </h2>
  {/if}

  {#if loadError}
    <p role="alert" class="text-error">{loadError}</p>
  {:else if list && list.texts.length === 0}
    <p class="text-muted">{list.total === 0 ? $_('library.empty') : $_('library.noMatch')}</p>
  {:else if list}
    <ul class="flex flex-col gap-2">
      {#each list.texts as t (t.id)}
        <li
          class={[
            'bg-panel flex flex-wrap items-center gap-x-4 gap-y-2 rounded-lg border px-4 py-3',
            highlight === t.id ? 'border-accent' : 'border-line',
          ]}
        >
          <div class="min-w-0 flex-1">
            <p class="font-medium">
              {t.title}
              {#each t.tags as tg (tg)}
                <span class="bg-line text-muted ms-1 rounded-full px-2 py-0.5 text-xs">{tg}</span>
              {/each}
            </p>
            <p class="text-muted truncate text-sm">{t.preview}</p>
            <p class="text-muted mt-1 flex flex-wrap gap-x-4 text-xs">
              <span>{$_('library.length', { values: { n: formatNumber(t.length, lang) } })}</span>
              {#if t.source === 'pdf'}
                <span
                  >{$_('library.fromPdf', { values: { file: t.sourceFile, pages: t.pages } })}</span
                >
              {:else if t.source === 'builtin'}
                <span>{$_('library.builtin')}</span>
              {/if}
              {#if t.bestWpm != null}
                <span
                  >{$_('library.best', { values: { wpm: formatNumber(t.bestWpm, lang, 1) } })}</span
                >
              {/if}
              {#if t.position > 0}
                <span>{$_('library.progress', { values: { pct: progress(t) } })}</span>
              {/if}
            </p>
          </div>
          <div class="flex shrink-0 gap-2">
            <a
              href={`#/practice?text=${t.id}`}
              class="bg-accent text-accent-ink rounded-md px-3 py-1.5 text-sm font-medium"
            >
              {t.position > 0 && progress(t) < 100 ? $_('library.resume') : $_('library.practice')}
            </a>
            <button
              class="border-line text-muted rounded-md border px-3 py-1.5 text-sm"
              onclick={() => edit(t.id)}
              aria-label={$_('library.editLabel', { values: { title: t.title } })}
            >
              {$_('library.edit')}
            </button>
            <button
              class="border-line text-muted rounded-md border px-3 py-1.5 text-sm"
              onclick={() => duplicate(t.id)}
              aria-label={$_('library.duplicateLabel', { values: { title: t.title } })}
            >
              {$_('library.duplicate')}
            </button>
            <button
              class={[
                'rounded-md border px-3 py-1.5 text-sm',
                confirmDelete === t.id ? 'border-error text-error' : 'border-line text-muted',
              ]}
              onclick={() => remove(t.id)}
              onblur={() => (confirmDelete = null)}
              aria-label={$_(
                confirmDelete === t.id ? 'library.confirmDeleteLabel' : 'library.deleteLabel',
                { values: { title: t.title } },
              )}
            >
              {confirmDelete === t.id ? $_('library.confirmDelete') : $_('library.delete')}
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>
