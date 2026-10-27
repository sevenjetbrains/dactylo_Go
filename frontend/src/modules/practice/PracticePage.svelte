<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { GetText, ListTexts } from '$wails/go/app/TextService';
  import { router } from '$app/router.svelte';
  import { DURATIONS, practiceSettings, type SessionType } from './settings.svelte';
  import TypingSession from './TypingSession.svelte';

  const textId = $derived(Number(router.query.get('text')) || 0);
  const text = $derived(textId ? GetText(textId) : null);
  const texts = $derived(textId ? null : ListTexts());
  // Incrémenté pour relancer le même texte quand il n'y en a qu'un.
  let round = $state(0);

  const sessionTypes: SessionType[] = ['text', 'time', 'words'];

  async function next() {
    const list = await ListTexts();
    if (list.length === 0) return;
    const i = list.findIndex((t) => t.id === textId);
    const nextText = list[(i + 1) % list.length]!;
    if (nextText.id === textId) round++;
    else router.navigate(`/practice?text=${nextText.id}`);
  }
</script>

<section class="mx-auto max-w-4xl px-8 py-10">
  {#if text}
    {#await text then t}
      {#key `${t.id}-${round}`}
        <TypingSession textId={t.id} title={t.title} content={t.content} onNext={next} />
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
          bind:value={practiceSettings.sessionType}
        >
          {#each sessionTypes as st (st)}
            <option value={st}>{$_(`practice.sessionType.${st}`)}</option>
          {/each}
        </select>
      </label>
      {#if practiceSettings.sessionType === 'time'}
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">{$_('practice.sessionType.duration')}</span>
          <select
            class="border-line bg-panel rounded-md border px-3 py-1.5"
            bind:value={practiceSettings.durationMin}
          >
            {#each DURATIONS as d (d)}
              <option value={d}>{$_('practice.sessionType.minutes', { values: { n: d } })}</option>
            {/each}
          </select>
        </label>
      {:else if practiceSettings.sessionType === 'words'}
        <label class="flex flex-col gap-1 text-sm">
          <span class="text-muted">{$_('practice.sessionType.words')}</span>
          <input
            type="number"
            min="5"
            max="5000"
            step="5"
            class="border-line bg-panel w-28 rounded-md border px-3 py-1.5"
            bind:value={practiceSettings.wordCount}
          />
        </label>
      {/if}
    </fieldset>

    <h2 class="mt-10 mb-3 font-semibold">{$_('practice.chooseText')}</h2>
    {#await texts then list}
      {#if !list || list.length === 0}
        <p class="text-muted">
          {$_('practice.noTexts')}
          <a href="#/library" class="text-accent underline">{$_('practice.addText')}</a>
        </p>
      {:else}
        <ul class="flex flex-col gap-2">
          {#each list as t (t.id)}
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
