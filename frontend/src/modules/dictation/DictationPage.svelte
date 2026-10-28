<!--
  Page du mode Dictée : choix du texte et des réglages, puis session.
-->
<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { GetText, ListTexts } from '$wails/go/app/TextService';
  import { router } from '$app/router.svelte';
  import { typingSettings } from '$shared/typing/settings.svelte';
  import DictationSession from './DictationSession.svelte';
  import DictationSettingsCard from './DictationSettingsCard.svelte';

  const textId = $derived(Number(router.query.get('text')) || 0);
  const text = $derived(textId ? GetText(textId) : null);
  const texts = $derived(
    textId
      ? null
      : ListTexts({
          search: '',
          tag: '',
          source: '',
          sort: 'recent',
          defaultSplit: typingSettings.exerciseSplit,
          defaultSize: typingSettings.exerciseSize,
        }),
  );
  // « Refaire la dictée » : nouvelle séance sur le même texte
  let round = $state(0);
</script>

<section class="mx-auto max-w-6xl px-10 py-10">
  {#if text}
    {#await text then t}
      {#key `${t.id}-${round}`}
        <DictationSession textId={t.id} title={t.title} onRedo={() => round++} />
      {/key}
    {:catch}
      <p role="alert" class="text-error">{$_('dictation.textNotFound')}</p>
    {/await}
  {:else}
    <p class="eyebrow">{$_('dictation.menu')}</p>
    <h1 class="mt-1 text-4xl font-bold tracking-tight">{$_('dictation.title')}</h1>
    <p class="text-muted mt-2">{$_('dictation.subtitle')}</p>

    <div class="mt-8 grid items-start gap-6 lg:grid-cols-[1fr_22rem]">
      <div>
        <h2 class="mb-3 font-bold">{$_('dictation.chooseText')}</h2>
        {#await texts then list}
          {#if !list || list.texts.length === 0}
            <p class="text-muted">{$_('dictation.noTexts')}</p>
          {:else}
            <ul class="flex flex-col gap-3">
              {#each list.texts as t (t.id)}
                <li class="card flex items-center gap-4 px-5 py-4">
                  <div class="min-w-0 flex-1">
                    <p class="truncate font-bold">{t.title}</p>
                    <p class="text-muted truncate text-sm">
                      {[...t.tags, $_('dictation.words', { values: { n: t.words } })].join(' · ')}
                    </p>
                  </div>
                  <a href={`#/dictation?text=${t.id}`} class="btn btn-soft shrink-0">
                    {$_('dictation.start')}
                  </a>
                </li>
              {/each}
            </ul>
          {/if}
        {/await}
      </div>
      <DictationSettingsCard />
    </div>
  {/if}
</section>
