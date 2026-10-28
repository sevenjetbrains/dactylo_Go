<!--
  Page du mode Copie : choix du texte, exercices du texte (F-24), session
  de frappe et reprise d'une session interrompue.
-->
<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { Exercises, GetText, ListTexts, ResetProgress } from '$wails/go/app/TextService';
  import { Resume } from '$wails/go/app/PracticeService';
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
  // Reprise d'une session interrompue (NF-06), proposée par l'accueil.
  const resumeId = $derived(Number(router.query.get('resume')) || 0);
  const resumed = $derived(resumeId ? Resume(resumeId) : null);
  // Incrémenté pour relancer la même page (un seul texte, rechargement).
  let round = $state(0);

  const session = $derived.by(() => {
    void round; // recharger aussi quand on relance la même adresse
    if (!textId || resumeId) return null;
    return load(textId, exParam, typingSettings.exerciseSplit, typingSettings.exerciseSize);
  });
  const texts = $derived(textId || resumeId ? null : ListTexts(allTexts()));

  const sessionTypes: SessionType[] = ['text', 'time', 'words'];
  const splits: ExerciseSplit[] = ['none', 'paragraph', 'lines'];

  /** Tous les textes, du plus récent au plus ancien. */
  function allTexts() {
    return {
      search: '',
      tag: '',
      source: '',
      sort: 'recent',
      defaultSplit: typingSettings.exerciseSplit,
      defaultSize: typingSettings.exerciseSize,
    };
  }

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

  /** Navigue ; même adresse : relance la session. */
  function go(to: string) {
    if (window.location.hash === '#' + to) round++;
    else router.navigate(to);
  }

  /** Passe au texte suivant de la bibliothèque. */
  async function nextText() {
    const { texts: list } = await ListTexts(allTexts());
    if (list.length === 0) return;
    const i = list.findIndex((t) => t.id === textId);
    go(`/practice?text=${list[(i + 1) % list.length]!.id}`);
  }

  /** Exercice suivant du texte, sinon texte suivant (Ctrl+Entrée). */
  async function next() {
    const s = await session;
    if (s && s.seg.index + 1 < s.plan.segments.length) {
      go(`/practice?text=${textId}&ex=${s.seg.index + 1}`);
    } else {
      await nextText();
    }
  }

  /** Efface la progression et reprend le texte au premier exercice. */
  async function restartText() {
    await ResetProgress(textId);
    go(`/practice?text=${textId}&ex=0`);
  }
</script>

<section class="mx-auto max-w-6xl px-10 py-10">
  {#if resumed}
    {#await resumed then r}
      <p class="text-muted mb-4 text-sm" data-testid="session-resumed">
        {$_('practice.resume.notice')}
      </p>
      {#key r.sessionId}
        <TypingSession
          textId={r.textId}
          title={r.title}
          eyebrow={$_('practice.resume.eyebrow')}
          start={r.start}
          end={r.end}
          content={r.content}
          resume={r}
          onNext={() => go(`/practice?text=${r.textId}`)}
          back={{ href: '#/practice', label: $_('practice.actions.chooseText') }}
        />
      {/key}
    {:catch}
      <p role="alert" class="text-error">{$_('practice.resume.failed')}</p>
      <a href="#/practice" class="text-accent underline">{$_('practice.actions.chooseText')}</a>
    {/await}
  {:else if session}
    {#await session then s}
      {#if s.plan.segments.length > 1}
        <nav
          class="text-muted relative z-10 -mb-6 flex flex-wrap items-center justify-end gap-3 text-sm"
          aria-label={$_('practice.exercise.nav')}
        >
          <label class="flex items-center gap-2">
            <span class="text-muted">{$_('practice.exercise.label')}</span>
            <select
              class="field py-1"
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
          eyebrow={$_('practice.eyebrow', {
            values: { n: s.seg.index + 1, total: s.plan.segments.length },
          })}
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
    <p class="eyebrow">{$_('practice.menu')}</p>
    <h1 class="mt-1 text-3xl font-bold tracking-tight">{$_('practice.title')}</h1>
    <p class="text-muted mt-2">{$_('practice.subtitle')}</p>

    <fieldset class="mt-8 flex flex-wrap items-end gap-4">
      <legend class="mb-2 text-sm font-semibold">{$_('practice.sessionType.label')}</legend>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-muted">{$_('practice.sessionType.mode')}</span>
        <select class="field" bind:value={typingSettings.sessionType}>
          {#each sessionTypes as st (st)}
            <option value={st}>{$_(`practice.sessionType.${st}`)}</option>
          {/each}
        </select>
      </label>
      {#if typingSettings.sessionType === 'time'}
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">{$_('practice.sessionType.duration')}</span>
          <select class="field" bind:value={typingSettings.durationMin}>
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
            class="field w-28"
            bind:value={typingSettings.wordCount}
          />
        </label>
      {/if}
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-muted">{$_('practice.exercise.split')}</span>
        <select class="field" bind:value={typingSettings.exerciseSplit}>
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
            class="field w-24"
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
              <a href={`#/practice?text=${t.id}`} class="card hover:border-accent block px-5 py-4">
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
