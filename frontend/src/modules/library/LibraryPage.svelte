<!--
  Bibliothèque : recherche, filtres, tri, création, modification,
  duplication et suppression des textes (F-23).
-->
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
  import { router } from '$app/router.svelte';
  import Icon from '$shared/Icon.svelte';
  import { formatNumber } from '$shared/format';
  import { typingSettings } from '$shared/typing/settings.svelte';
  import PdfImport from './PdfImport.svelte';
  import TextForm from './TextForm.svelte';

  type Panel =
    { kind: 'none' } | { kind: 'new' } | { kind: 'pdf' } | { kind: 'edit'; text: store.Text };
  const SORTS = ['recent', 'oldest', 'title', 'length', 'score'] as const;
  const MAX_CHIPS = 6;
  const ICON_SEARCH = 'M11 4a7 7 0 1 0 0 14 7 7 0 0 0 0-14zM21 21l-5-5';
  const ICON_PLUS = 'M12 5v14M5 12h14';

  let list: app.TextList | null = $state(null);
  let loadError = $state('');
  let search = $state('');
  // filtre : '' = tous, 'tag:<x>' = étiquette, 'pdf' = textes importés
  let filter = $state('');
  let sort: (typeof SORTS)[number] = $state('recent');
  let panel: Panel = $state(initialPanel());
  let confirmDelete: number | null = $state(null);
  let highlight: number | null = $state(null);

  const lang = $derived($locale ?? 'fr');

  /** #/library?new=paste|pdf ouvre directement le panneau (boutons de l'accueil). */
  function initialPanel(): Panel {
    const q = router.query.get('new');
    return q === 'pdf' ? { kind: 'pdf' } : q === 'paste' ? { kind: 'new' } : { kind: 'none' };
  }

  /** Recharge la liste selon la recherche, les filtres et le tri. */
  async function refresh() {
    try {
      list = await ListTexts({
        search,
        tag: filter.startsWith('tag:') ? filter.slice(4) : '',
        source: filter === 'pdf' ? 'pdf' : '',
        sort,
        defaultSplit: typingSettings.exerciseSplit,
        defaultSize: typingSettings.exerciseSize,
      });
      loadError = '';
    } catch (e) {
      loadError = String(e);
    }
  }

  // Recherche à la frappe, avec un court délai.
  $effect(() => {
    void [search, filter, sort];
    const t = setTimeout(refresh, 150);
    return () => clearTimeout(t);
  });

  /** Ferme le panneau de création ou de modification. */
  function close() {
    panel = { kind: 'none' };
    if (router.query.get('new')) router.navigate('/library');
  }

  /** Après un enregistrement : fermer, recharger et surligner le texte. */
  async function saved(id: number) {
    close();
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

  /** Duplique un texte (copie modifiable). */
  async function duplicate(id: number) {
    await saved(await DuplicateText(id));
  }

  /** Supprime un texte après confirmation. */
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

  /** Ligne d'informations d'un texte : longueur, source, date. */
  function meta(t: store.TextSummary): string {
    const parts = [...t.tags];
    parts.push($_('library.exercises', { values: { n: t.exercises } }));
    if (t.source === 'pdf' && t.pages)
      parts.unshift($_('library.pagesShort', { values: { pages: t.pages } }));
    if (t.source === 'builtin') parts.push($_('library.builtin'));
    return parts.join(' · ');
  }
</script>

<section class="mx-auto flex max-w-6xl gap-6 px-10 py-10">
  <div class="min-w-0 flex-1">
    <div class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <!-- compteur affiché une fois la liste chargée (pas de « Aucun texte » fugace) -->
        <p class="eyebrow min-h-4">
          {list ? $_('library.count', { values: { n: list.total } }) : ''}
        </p>
        <h1 class="mt-1 text-4xl font-bold tracking-tight">{$_('library.title')}</h1>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-outline" onclick={() => (panel = { kind: 'new' })}>
          <Icon d={ICON_PLUS} size={16} />
          {$_('library.form.heading')}
        </button>
        {#if panel.kind !== 'pdf'}
          <button class="btn btn-outline" onclick={() => (panel = { kind: 'pdf' })}>
            {$_('library.pdf.title')}
          </button>
        {/if}
      </div>
    </div>

    <div class="mt-6 flex flex-wrap items-center gap-3" role="search">
      <label class="field flex min-w-56 flex-1 items-center gap-2 py-2.5">
        <Icon d={ICON_SEARCH} size={18} class="text-muted" />
        <span class="sr-only">{$_('library.search')}</span>
        <input
          type="search"
          class="w-full bg-transparent outline-none"
          bind:value={search}
          placeholder={$_('library.searchPlaceholder')}
        />
      </label>
      <label class="text-muted flex items-center gap-2 text-sm">
        {$_('library.sort')}
        <select class="field text-ink py-2.5" bind:value={sort}>
          {#each SORTS as s (s)}
            <option value={s}>{$_(`library.sorts.${s}`)}</option>
          {/each}
        </select>
      </label>
    </div>

    <div class="mt-4 flex flex-wrap gap-2" role="radiogroup" aria-label={$_('library.tag')}>
      {#each [{ id: '', label: $_('library.allTags') }, ...(list?.tags ?? [])
          .slice(0, MAX_CHIPS)
          .map( (t) => ({ id: `tag:${t}`, label: t }) ), { id: 'pdf', label: $_('library.imported') }] as chip (chip.id)}
        <button
          role="radio"
          aria-checked={filter === chip.id}
          class={[
            'rounded-full border px-4 py-1.5 text-sm font-semibold',
            filter === chip.id
              ? 'bg-nav border-nav text-white'
              : 'border-line bg-panel hover:border-accent',
          ]}
          onclick={() => (filter = chip.id)}
        >
          {chip.label}
        </button>
      {/each}
    </div>

    {#if list && list.texts.length !== list.total}
      <p class="text-muted mt-4 text-sm">
        {$_('library.filtered', { values: { n: list.texts.length, total: list.total } })}
      </p>
    {/if}

    {#if loadError}
      <p role="alert" class="text-error mt-6">{loadError}</p>
    {:else if list && list.texts.length === 0}
      <p class="text-muted mt-6">
        {list.total === 0 ? $_('library.empty') : $_('library.noMatch')}
      </p>
    {:else if list}
      <ul class="mt-5 flex flex-col gap-3">
        {#each list.texts as t (t.id)}
          <li
            class={[
              'card flex items-center gap-4 px-5 py-4',
              highlight === t.id && 'border-accent',
            ]}
          >
            <span
              class={[
                'grid h-11 w-11 shrink-0 place-items-center rounded-lg font-mono text-xs font-bold',
                t.source === 'pdf' ? 'bg-error-soft text-error' : 'bg-accent-soft text-accent',
              ]}
              aria-hidden="true">{t.source === 'pdf' ? 'PDF' : 'TXT'}</span
            >
            <div class="min-w-0 flex-1">
              <p class="truncate font-bold">{t.title}</p>
              <p class="text-muted truncate text-sm">{meta(t)}</p>
            </div>
            <div class="w-28 shrink-0 text-right">
              <p class="font-mono text-sm tabular-nums">
                {t.bestWpm != null ? `${formatNumber(t.bestWpm, lang)} MPM` : '—'}
              </p>
              <div
                class="bg-line mt-1.5 h-1 overflow-hidden rounded-full"
                role="progressbar"
                aria-valuenow={progress(t)}
                aria-valuemin={0}
                aria-valuemax={100}
                aria-label={$_('library.progress', { values: { pct: progress(t) } })}
              >
                <div class="bg-accent h-full" style:width="{progress(t)}%"></div>
              </div>
            </div>
            <a href={`#/practice?text=${t.id}`} class="btn btn-soft shrink-0">
              {t.position > 0 && progress(t) < 100 ? $_('library.resume') : $_('library.open')}
            </a>
            <details class="relative shrink-0">
              <summary
                class="text-muted hover:text-ink cursor-pointer list-none rounded-lg px-2 py-1 text-lg"
                aria-label={$_('library.more', { values: { title: t.title } })}>⋯</summary
              >
              <div class="card absolute right-0 z-10 mt-1 flex w-44 flex-col p-1 text-sm shadow-lg">
                <button
                  class="hover:bg-sunken rounded-lg px-3 py-2 text-left"
                  onclick={() => edit(t.id)}
                  aria-label={$_('library.editLabel', { values: { title: t.title } })}
                  >{$_('library.edit')}</button
                >
                <button
                  class="hover:bg-sunken rounded-lg px-3 py-2 text-left"
                  onclick={() => duplicate(t.id)}
                  aria-label={$_('library.duplicateLabel', { values: { title: t.title } })}
                  >{$_('library.duplicate')}</button
                >
                <button
                  class={[
                    'hover:bg-sunken rounded-lg px-3 py-2 text-left',
                    confirmDelete === t.id && 'text-error font-semibold',
                  ]}
                  onclick={() => remove(t.id)}
                  aria-label={$_(
                    confirmDelete === t.id ? 'library.confirmDeleteLabel' : 'library.deleteLabel',
                    { values: { title: t.title } },
                  )}
                  >{confirmDelete === t.id
                    ? $_('library.confirmDelete')
                    : $_('library.delete')}</button
                >
              </div>
            </details>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  {#if panel.kind !== 'none'}
    <aside class="card sticky top-6 w-[23rem] shrink-0 self-start p-6 shadow-lg">
      {#if panel.kind === 'new'}
        <TextForm
          heading={$_('library.form.heading')}
          submitLabel={$_('library.form.save')}
          onSubmit={create}
          onCancel={close}
        />
      {:else if panel.kind === 'edit'}
        {@const t = panel.text}
        {#key t.id}
          <TextForm
            heading={$_('library.editHeading', { values: { title: t.title } })}
            initial={t}
            submitLabel={$_('library.form.update')}
            onSubmit={(input) => update(t.id, input)}
            onCancel={close}
          />
        {/key}
      {:else}
        <PdfImport onDone={saved} onCancel={close} />
      {/if}
    </aside>
  {/if}
</section>
