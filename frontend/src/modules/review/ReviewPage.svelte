<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import {
    AddWord,
    BuildExercise,
    BuildExerciseFor,
    ListWords,
    MarkMastered,
    RemoveWord,
    Summary,
  } from '$wails/go/app/ReviewService';
  import type { app, store } from '$wails/go/models';
  import { router } from '$app/router.svelte';
  import TypingSession from '$shared/typing/TypingSession.svelte';
  import { localDay, parseDay } from '$shared/format';

  const KINDS = ['shuffled', 'repeated', 'sentences'] as const;
  type Kind = (typeof KINDS)[number];
  const FILTERS = ['due', 'learning', 'mastered', 'all'] as const;
  type Filter = (typeof FILTERS)[number];

  // #/review?session=<type>&repeat=N : séance de révision ; sinon la liste.
  const sessionKind = $derived(router.query.get('session') as Kind | null);
  const repeatParam = $derived(Number(router.query.get('repeat')) || 3);
  // &words=a|b|c : mots choisis (« S'entraîner sur ces mots » après une dictée)
  const wordsParam = $derived(router.query.get('words')?.split('|').filter(Boolean) ?? []);

  let kind: Kind = $state('shuffled');
  let repeat = $state(3);
  let filter: Filter = $state('due');
  let summary: app.ReviewSummary | null = $state(null);
  let words: store.Word[] = $state([]);
  let newWord = $state('');
  let error = $state('');
  let round = $state(0);

  const lang = $derived($locale ?? 'fr');
  const today = localDay(new Date());

  async function refresh() {
    try {
      [summary, words] = await Promise.all([Summary(), ListWords(filter)]);
      error = '';
    } catch (e) {
      error = String(e);
    }
  }

  $effect(() => {
    void filter;
    if (!sessionKind) void refresh();
  });

  // Exercice généré à chaque entrée en séance (et à chaque « Suivant »).
  const exercise = $derived.by(() => {
    void round;
    if (!sessionKind) return null;
    return wordsParam.length > 0
      ? BuildExerciseFor(wordsParam, sessionKind, repeatParam)
      : BuildExercise(sessionKind, repeatParam);
  });

  function start() {
    router.navigate(`/review?session=${kind}&repeat=${repeat}`);
  }

  async function add(e: SubmitEvent) {
    e.preventDefault();
    try {
      await AddWord(newWord);
      newWord = '';
      filter = 'due';
      await refresh();
    } catch (err) {
      error = String(err);
    }
  }

  async function act(fn: (id: number) => Promise<void>, id: number) {
    await fn(id);
    await refresh();
  }

  function formatDate(s: string): string {
    if (!s) return '—';
    const d = parseDay(s);
    if (localDay(d) === today) return $_('review.today');
    return d.toLocaleDateString(lang, { day: 'numeric', month: 'short' });
  }
</script>

<section class="mx-auto max-w-6xl px-10 py-10">
  {#if exercise}
    {#await exercise then ex}
      {#key `${ex.content}-${round}`}
        <TypingSession
          title={$_('review.sessionTitle', { values: { n: ex.words.length } })}
          content={ex.content}
          onNext={() => round++}
          back={{ href: '#/review', label: $_('review.backToList') }}
        />
      {/key}
    {:catch e}
      <p role="alert" class="text-error">{String(e)}</p>
      <a href="#/review" class="text-accent underline">{$_('review.backToList')}</a>
    {/await}
  {:else}
    <h1 class="text-4xl font-bold tracking-tight">{$_('review.title')}</h1>
    <p class="text-muted mt-2">{$_('review.subtitle')}</p>

    {#if error}<p role="alert" class="text-error mt-4 text-sm">{error}</p>{/if}

    {#if summary}
      <div class="border-line bg-panel mt-8 flex flex-wrap items-end gap-6 rounded-lg border p-5">
        <div class="min-w-48 flex-1">
          <p class="text-3xl font-semibold tabular-nums" data-testid="due-count">
            {$_('review.dueToday', { values: { n: summary.due } })}
          </p>
          <p class="text-muted mt-1 text-sm">
            {$_('review.counts', {
              values: { learning: summary.learning, mastered: summary.mastered },
            })}
          </p>
        </div>
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">{$_('review.kind')}</span>
          <select class="field" bind:value={kind}>
            {#each KINDS as k (k)}
              <option value={k}>{$_(`review.kinds.${k}`)}</option>
            {/each}
            <option disabled>{$_('review.kinds.dictation')}</option>
          </select>
        </label>
        {#if kind === 'repeated'}
          <label class="flex flex-col gap-1 text-sm">
            <span class="text-muted">{$_('review.repeat')}</span>
            <input type="number" min="1" max="20" class="field w-20" bind:value={repeat} />
          </label>
        {/if}
        <button
          class="btn btn-primary"
          onclick={start}
          disabled={summary.due + summary.learning === 0}
        >
          {summary.due > 0 ? $_('review.start') : $_('review.startLearning')}
        </button>
      </div>
    {/if}

    <div class="mt-8 flex flex-wrap items-end justify-between gap-4">
      <div role="tablist" class="flex gap-1" aria-label={$_('review.filterLabel')}>
        {#each FILTERS as f (f)}
          <button
            role="tab"
            aria-selected={filter === f}
            class={[
              'rounded-md px-3 py-1.5 text-sm',
              filter === f
                ? 'bg-accent text-accent-ink font-medium'
                : 'text-muted hover:bg-line/60',
            ]}
            onclick={() => (filter = f)}
          >
            {$_(`review.filters.${f}`)}
          </button>
        {/each}
      </div>
      <form class="flex items-end gap-2" onsubmit={add}>
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">{$_('review.addLabel')}</span>
          <input class="field w-48" bind:value={newWord} required />
        </label>
        <button type="submit" class="btn btn-outline">
          {$_('review.add')}
        </button>
      </form>
    </div>

    {#if words.length === 0}
      <p class="text-muted mt-6">{$_(`review.empty.${filter}`)}</p>
    {:else}
      <div class="border-line mt-4 overflow-x-auto rounded-lg border">
        <table class="bg-panel w-full text-sm">
          <thead class="text-muted text-left text-xs">
            <tr class="border-line border-b">
              <th class="px-3 py-2 font-medium">{$_('review.cols.word')}</th>
              <th class="px-3 py-2 text-right font-medium">{$_('review.cols.encounters')}</th>
              <th class="px-3 py-2 text-right font-medium">{$_('review.cols.errors')}</th>
              <th class="px-3 py-2 font-medium">{$_('review.cols.forms')}</th>
              <th class="px-3 py-2 font-medium">{$_('review.cols.mistake')}</th>
              <th class="px-3 py-2 font-medium">{$_('review.cols.lastError')}</th>
              <th class="px-3 py-2 font-medium">{$_('review.cols.mastery')}</th>
              <th class="px-3 py-2 font-medium">{$_('review.cols.next')}</th>
              <th class="px-3 py-2"><span class="sr-only">{$_('review.cols.actions')}</span></th>
            </tr>
          </thead>
          <tbody>
            {#each words as w (w.id)}
              <tr class="border-line border-b last:border-0">
                <td class="px-3 py-2 font-mono font-semibold">{w.shown}</td>
                <td class="px-3 py-2 text-right tabular-nums">{w.encounters}</td>
                <td class="px-3 py-2 text-right tabular-nums">{w.errors}</td>
                <td class="px-3 py-2">
                  {#each w.forms as f (f)}
                    <span class="text-error me-1 font-mono line-through">{f}</span>
                  {:else}
                    <span class="text-muted">{w.errors === 0 ? $_('review.slowOnly') : '—'}</span>
                  {/each}
                </td>
                <td class="text-muted px-3 py-2">
                  {w.dominantError ? $_(`typing.errorTypes.${w.dominantError}`) : '—'}
                </td>
                <td class="text-muted px-3 py-2 whitespace-nowrap">{formatDate(w.lastErrorAt)}</td>
                <td class="px-3 py-2">
                  <div class="flex items-center gap-2">
                    <div
                      class="bg-line h-1.5 w-16 overflow-hidden rounded-full"
                      role="meter"
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-valuenow={Math.round(w.mastery * 100)}
                      aria-label={$_('review.cols.mastery')}
                    >
                      <div class="bg-accent h-full" style:width="{w.mastery * 100}%"></div>
                    </div>
                    <span class="text-muted text-xs whitespace-nowrap tabular-nums"
                      >{Math.round(w.mastery * 100)} %</span
                    >
                  </div>
                </td>
                <td class="px-3 py-2 whitespace-nowrap">
                  {w.mastery >= 1 ? $_('review.masteredLabel') : formatDate(w.nextReview)}
                </td>
                <td class="px-3 py-2 whitespace-nowrap">
                  {#if w.mastery < 1}
                    <button
                      class="text-accent me-3 text-xs underline"
                      onclick={() => act(MarkMastered, w.id)}
                      aria-label={$_('review.markMasteredLabel', { values: { word: w.shown } })}
                    >
                      {$_('review.markMastered')}
                    </button>
                  {/if}
                  <button
                    class="text-muted text-xs underline"
                    onclick={() => act(RemoveWord, w.id)}
                    aria-label={$_('review.removeLabel', { values: { word: w.shown } })}
                  >
                    {$_('review.remove')}
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {/if}
</section>
