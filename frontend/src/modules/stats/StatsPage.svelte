<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { Dashboard, ExportCSV } from '$wails/go/app/StatsService';
  import type { app } from '$wails/go/models';
  import { formatDuration, formatNumber, parseDay } from '$shared/format';
  import { profile } from '$shared/profile.svelte';
  import KeyboardHeatmap from './KeyboardHeatmap.svelte';
  import LineChart from './LineChart.svelte';

  let data = $state<app.Dashboard | null>(null);
  let error = $state('');
  let exportMsg = $state('');

  const lang = $derived($locale ?? 'fr');

  async function load() {
    try {
      data = await Dashboard();
    } catch (e) {
      error = String(e);
    }
  }
  void load();

  const labels = $derived(
    (data?.points ?? []).map((p) =>
      parseDay(p.endedAt).toLocaleDateString(lang, { day: 'numeric', month: 'short' }),
    ),
  );
  const worstKeys = $derived(
    [...(data?.keys ?? [])].filter((k) => k.count >= 5).sort((a, b) => b.errorRate - a.errorRate),
  );

  async function exportCsv() {
    exportMsg = '';
    try {
      const res = await ExportCSV($_('stats.export.dialogTitle'));
      if (res.path)
        exportMsg = $_('stats.export.done', { values: { n: res.sessions, path: res.path } });
    } catch (e) {
      exportMsg = String(e);
    }
  }

  const keyLabel = (k: string) => (k === ' ' ? '␣' : k);
</script>

<section class="mx-auto max-w-6xl px-10 py-10">
  <div class="flex flex-wrap items-center justify-between gap-4">
    <h1 class="text-4xl font-bold tracking-tight">{$_('stats.title')}</h1>
    <button class="btn btn-outline" onclick={exportCsv}>
      {$_('stats.export.button')}
    </button>
  </div>
  {#if exportMsg}<p class="text-muted mt-2 text-sm" role="status">{exportMsg}</p>{/if}
  {#if error}<p role="alert" class="text-error mt-4">{error}</p>{/if}

  {#if data}
    <dl class="mt-8 grid grid-cols-2 gap-4 sm:grid-cols-4" data-testid="totals">
      {#each [{ k: 'sessions', v: formatNumber(data.totals.sessions, lang) }, { k: 'time', v: formatDuration(data.totals.totalMs) }, { k: 'best', v: formatNumber(data.totals.bestWpm, lang, 1) }, { k: 'streak', v: $_( 'stats.days', { values: { n: data.streak } } ) }] as tile (tile.k)}
        <div class="card p-4">
          <dt class="text-muted text-sm">{$_(`stats.tiles.${tile.k}`)}</dt>
          <dd class="mt-1 text-2xl font-semibold tabular-nums">{tile.v}</dd>
        </div>
      {/each}
    </dl>

    {#if data.points.length === 0}
      <p class="text-muted mt-8">{$_('stats.empty')}</p>
    {:else}
      <div class="mt-6 grid gap-4 md:grid-cols-2">
        <LineChart
          title={$_('stats.charts.wpm')}
          {labels}
          values={data.points.map((p) => p.wpmNet)}
          format={(v) => formatNumber(v, lang)}
          unit=" MPM"
          min={0}
        />
        <LineChart
          title={$_('stats.charts.accuracy')}
          {labels}
          values={data.points.map((p) => p.accuracy)}
          format={(v) => formatNumber(v, lang)}
          unit=" %"
          max={100}
        />
      </div>

      <div class="mt-6 grid gap-4 lg:grid-cols-[2fr_1fr]">
        <div class="flex flex-col gap-2">
          <KeyboardHeatmap keys={data.keys} layout={profile.keyboardLayout} />
          <details class="text-sm">
            <summary class="text-muted cursor-pointer">{$_('stats.heatmap.asTable')}</summary>
            <table class="mt-2 w-full text-left text-xs">
              <thead class="text-muted">
                <tr>
                  <th class="py-1 font-medium">{$_('stats.cols.key')}</th>
                  <th class="py-1 text-right font-medium">{$_('stats.cols.count')}</th>
                  <th class="py-1 text-right font-medium">{$_('stats.cols.errorRate')}</th>
                  <th class="py-1 text-right font-medium">{$_('stats.cols.avgMs')}</th>
                </tr>
              </thead>
              <tbody>
                {#each worstKeys as k (k.key)}
                  <tr>
                    <td class="py-0.5 font-mono">{keyLabel(k.key)}</td>
                    <td class="py-0.5 text-right tabular-nums">{k.count}</td>
                    <td class="py-0.5 text-right tabular-nums"
                      >{formatNumber(k.errorRate, lang, 1)} %</td
                    >
                    <td class="py-0.5 text-right tabular-nums">{formatNumber(k.avgMs, lang)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </details>
        </div>

        <figure class="card p-4">
          <figcaption class="text-sm font-semibold">{$_('stats.bigrams.title')}</figcaption>
          {#if data.bigrams.length === 0}
            <p class="text-muted mt-2 text-sm">{$_('stats.bigrams.empty')}</p>
          {:else}
            {@const slowest = data.bigrams[0]?.avgMs ?? 1}
            <ol class="mt-3 flex flex-col gap-2" data-testid="bigrams">
              {#each data.bigrams as b (b.bigram)}
                <li class="flex items-center gap-3 text-sm">
                  <span class="w-8 font-mono font-semibold">{b.bigram}</span>
                  <span class="bg-line h-2 flex-1 overflow-hidden rounded-full" aria-hidden="true">
                    <span
                      class="block h-full rounded-full"
                      style:width="{(100 * b.avgMs) / slowest}%"
                      style:background="var(--viz-series-1)"
                    ></span>
                  </span>
                  <span class="text-muted w-24 text-right text-xs tabular-nums">
                    {$_('stats.bigrams.value', {
                      values: { ms: formatNumber(b.avgMs, lang), n: b.count },
                    })}
                  </span>
                </li>
              {/each}
            </ol>
          {/if}
        </figure>
      </div>
    {/if}
  {/if}
</section>
