<!--
  Écran de fin de session : résultats, comparaison à la moyenne, mots fautifs.
-->
<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import type { app } from '$wails/go/models';
  import { formatDuration, formatNumber } from '$shared/format';

  let {
    result,
    onRestart,
    onNext,
    back,
  }: {
    result: app.SessionResult;
    onRestart: () => void;
    onNext: () => void;
    /** Lien de sortie fourni par le module appelant. */
    back?: { href: string; label: string };
  } = $props();

  const lang = $derived($locale ?? 'fr');
  const s = $derived(result.summary);
  const avg = $derived(result.average);

  /** Écart signé à la moyenne des sessions précédentes. */
  function delta(value: number, reference: number): string {
    const d = value - reference;
    return (d >= 0 ? '+' : '−') + formatNumber(Math.abs(d), lang, 1);
  }
</script>

<section class="flex flex-col gap-8" aria-labelledby="result-title">
  <h2 id="result-title" class="text-2xl font-bold">{$_('typing.result.title')}</h2>

  <dl class="grid grid-cols-2 gap-4 sm:grid-cols-3">
    {#each [{ key: 'wpmNet', value: formatNumber(s.wpmNet, lang, 1) }, { key: 'accuracy', value: formatNumber(s.accuracy, lang, 1) + ' %' }, { key: 'wpmGross', value: formatNumber(s.wpmGross, lang, 1) }, { key: 'cpm', value: formatNumber(s.cpm, lang) }, { key: 'duration', value: formatDuration(s.durationMs) }, { key: 'errors', value: String(s.uncorrectedErrors) }] as stat (stat.key)}
      <div class="card p-5">
        <dt class="text-muted text-sm">{$_(`typing.result.${stat.key}`)}</dt>
        <dd class="mt-1 text-3xl font-bold tabular-nums">{stat.value}</dd>
      </div>
    {/each}
  </dl>

  <p class="text-sm" data-testid="comparison">
    {#if avg.sessions > 0}
      {$_('typing.result.comparison', {
        values: {
          n: avg.sessions,
          wpm: delta(s.wpmNet, avg.wpmNet),
          accuracy: delta(s.accuracy, avg.accuracy),
        },
      })}
    {:else}
      {$_('typing.result.firstSession')}
    {/if}
  </p>

  <div>
    <h3 class="mb-3 font-semibold">{$_('typing.result.mistakes')}</h3>
    {#if result.mistakes.length === 0}
      <p class="text-muted text-sm">{$_('typing.result.noMistakes')}</p>
    {:else}
      <ul class="flex flex-wrap gap-2">
        {#each result.mistakes as m (m.position)}
          <li class="card px-3 py-1.5 text-sm">
            <span class="font-mono font-semibold">{m.expected}</span>
            <span class="text-muted">←</span>
            <span class="text-error font-mono line-through">{m.typed}</span>
            <span class="text-muted ms-1 text-xs">{$_(`typing.errorTypes.${m.type}`)}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  <div class="flex flex-wrap gap-3">
    <button class="btn btn-primary" onclick={onRestart}>
      {$_('typing.actions.restart')} <kbd class="opacity-70">Ctrl+R</kbd>
    </button>
    <button class="btn btn-outline" onclick={onNext}>
      {$_('typing.actions.next')} <kbd class="opacity-70">Ctrl+Entrée</kbd>
    </button>
    {#if back}
      <a href={back.href} class="text-muted px-2 py-2 text-sm underline">{back.label}</a>
    {/if}
  </div>
</section>
