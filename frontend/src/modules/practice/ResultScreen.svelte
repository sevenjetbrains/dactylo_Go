<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import type { app } from '$wails/go/models';
  import { formatDuration, formatNumber } from '$shared/format';

  let {
    result,
    onRestart,
    onNext,
  }: { result: app.SessionResult; onRestart: () => void; onNext: () => void } = $props();

  const lang = $derived($locale ?? 'fr');
  const s = $derived(result.summary);
  const avg = $derived(result.average);

  function delta(value: number, reference: number): string {
    const d = value - reference;
    return (d >= 0 ? '+' : '−') + formatNumber(Math.abs(d), lang, 1);
  }
</script>

<section class="flex flex-col gap-8" aria-labelledby="result-title">
  <h2 id="result-title" class="text-2xl font-semibold">{$_('practice.result.title')}</h2>

  <dl class="grid grid-cols-2 gap-4 sm:grid-cols-3">
    {#each [{ key: 'wpmNet', value: formatNumber(s.wpmNet, lang, 1) }, { key: 'accuracy', value: formatNumber(s.accuracy, lang, 1) + ' %' }, { key: 'wpmGross', value: formatNumber(s.wpmGross, lang, 1) }, { key: 'cpm', value: formatNumber(s.cpm, lang) }, { key: 'duration', value: formatDuration(s.durationMs) }, { key: 'errors', value: String(s.uncorrectedErrors) }] as stat (stat.key)}
      <div class="border-line bg-panel rounded-lg border p-4">
        <dt class="text-muted text-sm">{$_(`practice.result.${stat.key}`)}</dt>
        <dd class="mt-1 text-2xl font-semibold tabular-nums">{stat.value}</dd>
      </div>
    {/each}
  </dl>

  <p class="text-sm" data-testid="comparison">
    {#if avg.sessions > 0}
      {$_('practice.result.comparison', {
        values: {
          n: avg.sessions,
          wpm: delta(s.wpmNet, avg.wpmNet),
          accuracy: delta(s.accuracy, avg.accuracy),
        },
      })}
    {:else}
      {$_('practice.result.firstSession')}
    {/if}
  </p>

  <div>
    <h3 class="mb-3 font-semibold">{$_('practice.result.mistakes')}</h3>
    {#if result.mistakes.length === 0}
      <p class="text-muted text-sm">{$_('practice.result.noMistakes')}</p>
    {:else}
      <ul class="flex flex-wrap gap-2">
        {#each result.mistakes as m (m.position)}
          <li class="border-line bg-panel rounded-md border px-3 py-1.5 text-sm">
            <span class="font-mono font-semibold">{m.expected}</span>
            <span class="text-muted">←</span>
            <span class="text-error font-mono line-through">{m.typed}</span>
            <span class="text-muted ms-1 text-xs">{$_(`practice.errorTypes.${m.type}`)}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  <div class="flex flex-wrap gap-3">
    <button
      class="bg-accent text-accent-ink rounded-md px-4 py-2 text-sm font-medium"
      onclick={onRestart}
    >
      {$_('practice.actions.restart')} <kbd class="opacity-70">Ctrl+R</kbd>
    </button>
    <button class="border-line rounded-md border px-4 py-2 text-sm font-medium" onclick={onNext}>
      {$_('practice.actions.next')} <kbd class="opacity-70">Ctrl+Entrée</kbd>
    </button>
    <a href="#/practice" class="text-muted px-2 py-2 text-sm underline">
      {$_('practice.actions.chooseText')}
    </a>
  </div>
</section>
