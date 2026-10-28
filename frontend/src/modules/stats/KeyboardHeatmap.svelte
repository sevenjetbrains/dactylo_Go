<!--
  Carte de chaleur du clavier : erreurs ou lenteur par touche, dans la
  disposition du profil (F-34).
-->
<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import type { stats } from '$wails/go/models';
  import { formatNumber } from '$shared/format';
  import {
    layoutRows,
    heatByKey,
    MIN_COUNT,
    scaleRange,
    stepFor,
    STEPS,
    type HeatMetric,
  } from './heat';

  let { keys, layout = 'azerty' }: { keys: stats.KeyStat[]; layout?: string } = $props();

  let metric: HeatMetric = $state('errors');
  const lang = $derived($locale ?? 'fr');
  const heat = $derived(heatByKey(keys, layout));
  const range = $derived(scaleRange(heat, metric));

  const fmt = (n: number) =>
    metric === 'errors' ? `${formatNumber(n, lang, 1)} %` : `${formatNumber(n, lang)} ms`;
</script>

<figure class="border-line bg-panel flex flex-col gap-4 rounded-lg border p-4">
  <figcaption class="flex flex-wrap items-center justify-between gap-3">
    <span class="text-sm font-semibold">{$_('stats.heatmap.title')}</span>
    <div role="radiogroup" aria-label={$_('stats.heatmap.metric')} class="flex gap-1 text-xs">
      {#each ['errors', 'speed'] as const as m (m)}
        <button
          role="radio"
          aria-checked={metric === m}
          class={[
            'rounded px-2 py-1',
            metric === m ? 'bg-accent text-accent-ink' : 'text-muted hover:bg-line/60',
          ]}
          onclick={() => (metric = m)}
        >
          {$_(`stats.heatmap.metrics.${m}`)}
        </button>
      {/each}
    </div>
  </figcaption>

  <div class="flex flex-col gap-1 font-mono text-sm" data-testid="heatmap">
    {#each layoutRows(layout) as row, r (r)}
      <div class="flex gap-1" style:padding-inline-start="{r * 0.6}rem">
        {#each row as key (key.label)}
          {@const h = heat.get(key)}
          {@const step = stepFor(h, metric, range)}
          <div
            class="flex h-10 items-center justify-center rounded"
            style:flex="{key.width ?? 1} 1 0"
            style:max-width="{(key.width ?? 1) * 2.75}rem"
            style:background={step === null ? 'var(--color-line)' : `var(--viz-seq-${step})`}
            style:color={step === null
              ? 'var(--color-muted)'
              : step < 3
                ? 'var(--viz-seq-ink-light)'
                : 'var(--viz-seq-ink-dark)'}
            title={h
              ? $_('stats.heatmap.tooltip', {
                  values: {
                    key: key.label,
                    count: h.count,
                    rate: formatNumber(h.errorRate, lang, 1),
                    ms: formatNumber(h.avgMs, lang),
                  },
                })
              : key.label}
            data-step={step ?? ''}
          >
            {key.label}
          </div>
        {/each}
      </div>
    {/each}
  </div>

  <div class="text-muted flex flex-wrap items-center gap-3 text-xs">
    {#if range[1] > range[0] || range[1] > 0}
      <span>{fmt(range[0])}</span>
      <div class="flex overflow-hidden rounded" aria-hidden="true">
        {#each Array.from({ length: STEPS }, (_, i) => i) as i (i)}
          <span class="h-3 w-6" style:background="var(--viz-seq-{i})"></span>
        {/each}
      </div>
      <span>{fmt(range[1])}</span>
    {/if}
    <span class="flex items-center gap-1">
      <span class="bg-line inline-block h-3 w-4 rounded"></span>
      {$_('stats.heatmap.noData', { values: { n: MIN_COUNT } })}
    </span>
  </div>
</figure>
