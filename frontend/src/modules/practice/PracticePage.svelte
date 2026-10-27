<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { Exercises, GetText, ListTexts, ResetProgress } from '$wails/go/app/TextService';
  import { router } from '$app/router.svelte';
  import {
    DURATIONS,
    typingSettings,
    type ExerciseSplit,
    type SessionType,
  } from '$shared/typing/settings.svelte';
  import TypingSession from '$shared/typing/TypingSession.svelte';

  const textId = $derived(Number(router.query.get('text')) || 0);
  const exParam = $derived(router.query.get('ex'));
  // Incrémenté pour relancer la même page (un seul texte, rechargement).
  let round = $state(0);

  const session = $derived.by(() => {
    void round; // recharger aussi quand on relance la même adresse
    if (!textId) return null;
    return load(textId, exParam, typingSettings.exerciseSplit, typingSettings.exerciseSize);
  });
  const texts = $derived(textId ? null : ListTexts({ search: '', tag: '', sort: 'recent' }));

  const sessionTypes: SessionType[] = ['text', 'time', 'words'];
  const splits: ExerciseSplit[] = ['none', 'paragraph', 'lines'];

  /** Charge le texte et choisit l'exercice : celui demandé, sinon la reprise (F-24). */
  async function load(id: number, ex: string | null, split: string, size: number) {
    const [text, plan] = await Promise.all([GetText(id), Exercises(id, split, size)]);
    const requested = ex === null ? undefined : plan.segments[Number(ex)];
    const seg = requested ?? plan.resume;
    const start = requested ? requested.start : plan.resumeStart;
    const chars = Array.from(text.content);
    return {
      text,
      plan,
      seg,
      start,
      resumed: !requested && start > seg.start,
      content: chars.slice(start, seg.end).join(''),
    };
  }

  function go(to: string) {
    if (window.location.hash === '#' + to) round++;
    else router.navigate(to);
  }

  async function nextText() {
    const { texts: list } = await ListTexts({ search: '', tag: '', sort: 'recent' });
    if (list.length === 0) return;
    const i = list.findIndex((t) => t.id === textId);
    go(`/practice?text=${list[(i + 1) % list.length]!.id}`);
  }

  async function next() {
    const s = await session;
    if (s && s.seg.index + 1 < s.plan.segments.length) {
      go(`/practice?text=${textId}&ex=${s.seg.index + 1}`);
    } else {
      await nextText();
    }
  }

  async function restartText() {
    await ResetProgress(textId);
    go(`/practice?text=${textId}&ex=0`);
  }
</script>

<section class="mx-auto max-w-4xl px-8 py-10">
  {#if session}
    {#await session then s}
      {#if s.plan.segments.length > 1}
        <nav
          class="mb-4 flex flex-wrap items-center gap-3 text-sm"
          aria-label={$_('practice.exercise.nav')}
        >
          <label class="flex items-center gap-2">
            <span class="text-muted">{$_('practice.exercise.label')}</span>
            <select
              class="border-line bg-panel rounded-md border px-2 py-1"
              value={s.seg.index}
              onchange={(e) => go(`/practice?text=${textId}&ex=${e.currentTarget.value}`)}
            >
              {#each s.plan.segments as seg (seg.index)}
                <option value={seg.index}>
                  {$_('practice.exercise.option', {
                    values: { n: seg.index + 1, total: s.plan.segments.length },
                  })}
                </option>
              {/each}
            </select>
          </label>
          <span class="text-muted">
            {$_('practice.exercise.progress', {
              values: {
                pct: Math.round((100 * Math.min(s.plan.position, s.plan.length)) / s.plan.length),
              },
            })}
          </span>
          <button class="text-muted underline" onclick={restartText}>
            {$_('practice.exercise.restartText')}
          </button>
        </nav>
      {/if}
      {#if s.resumed}
        <p class="text-muted mb-4 text-sm" data-testid="resumed">
          {$_('practice.exercise.resumed')}
        </p>
      {/if}
      {#key `${s.text.id}-${s.start}-${s.seg.end}-${round}`}
        <TypingSession
          textId={s.text.id}
          title={s.text.title}
          start={s.start}
          end={s.seg.end}
          content={s.content}
          onNext={next}
          back={{ href: '#/practice', label: $_('practice.actions.chooseText') }}
        />
      {/key}
    {:catch}
      <p role="alert" class="text-error">{$_('practice.textNotFound')}</p>
      <a href="#/practice" class="text-accent underline">{$_('practice.actions.chooseText')}</a>
    {/await}
  {:else}
    <h1 class="text-3xl font-semibold">{$_('practice.title')}</h1>
    <p class="text-muted mt-2">{$_('practice.subtitle')}</p>

    <fieldset class="mt-8 flex flex-wrap items-end gap-4">
      <legend class="mb-2 text-sm font-semibold">{$_('practice.sessionType.label')}</legend>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-muted">{$_('practice.sessionType.mode')}</span>
        <select
          class="border-line bg-panel rounded-md border px-3 py-1.5"
          bind:value={typingSettings.sessionType}
        >
          {#each sessionTypes as st (st)}
            <option value={st}>{$_(`practice.sessionType.${st}`)}</option>
          {/each}
        </select>
      </label>
      {#if typingSettings.sessionType === 'time'}
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">{$_('practice.sessionType.duration')}</span>
          <select
            class="border-line bg-panel rounded-md border px-3 py-1.5"
            bind:value={typingSettings.durationMin}
          >
            {#each DURATIONS as d (d)}
              <option value={d}>{$_('practice.sessionType.minutes', { values: { n: d } })}</option>
            {/each}
          </select>
        </label>
      {:else if typingSettings.sessionType === 'words'}
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">{$_('practice.sessionType.words')}</span>
          <input
            type="number"
            min="5"
            max="5000"
            step="5"
            class="border-line bg-panel w-28 rounded-md border px-3 py-1.5"
            bind:value={typingSettings.wordCount}
          />
        </label>
      {/if}
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-muted">{$_('practice.exercise.split')}</span>
        <select
          class="border-line bg-panel rounded-md border px-3 py-1.5"
          bind:value={typingSettings.exerciseSplit}
        >
          {#each splits as sp (sp)}
            <option value={sp}>{$_(`practice.exercise.splits.${sp}`)}</option>
          {/each}
        </select>
      </label>
      {#if typingSettings.exerciseSplit !== 'none'}
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">
            {$_(`practice.exercise.size.${typingSettings.exerciseSplit}`)}
          </span>
          <input
            type="number"
            min="1"
            max="100"
            class="border-line bg-panel w-24 rounded-md border px-3 py-1.5"
            bind:value={typingSettings.exerciseSize}
          />
        </label>
      {/if}
    </fieldset>

    <h2 class="mt-10 mb-3 font-semibold">{$_('practice.chooseText')}</h2>
    {#await texts then list}
      {#if !list || list.texts.length === 0}
        <p class="text-muted">
          {$_('practice.noTexts')}
          <a href="#/library" class="text-accent underline">{$_('practice.addText')}</a>
        </p>
      {:else}
        <ul class="flex flex-col gap-2">
          {#each list.texts as t (t.id)}
            <li>
              <a
                href={`#/practice?text=${t.id}`}
                class="border-line bg-panel hover:border-accent block rounded-lg border px-4 py-3"
              >
                <span class="font-medium">{t.title}</span>
                <span class="text-muted ms-2 text-xs">
                  {$_('practice.length', { values: { n: t.length } })}
                </span>
                <span class="text-muted mt-1 block truncate text-sm">{t.preview}</span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    {:catch e}
      <p role="alert" class="text-error">{String(e)}</p>
    {/await}
  {/if}
</section>
