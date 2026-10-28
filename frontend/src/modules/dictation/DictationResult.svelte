<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { RemoveWordByKey } from '$wails/go/app/ReviewService';
  import type { app } from '$wails/go/models';
  import { router } from '$app/router.svelte';
  import Icon from '$shared/Icon.svelte';
  import { formatDuration, formatNumber } from '$shared/format';

  // Maquette « Correction de la dictée » (F-16).
  let {
    result,
    title,
    onRedo,
  }: { result: app.DictationResult; title: string; onRedo: () => void } = $props();

  const TYPES = [
    'accent',
    'case',
    'missing_letter',
    'extra_letter',
    'punctuation',
    'transposition',
    'omitted_word',
    'substitution',
    'extra_word',
  ] as const;
  const PREVIEW_OPS = 40;
  const SPACE = ' '; // séparateur des mots du texte corrigé
  /** Libellé de type de faute, avec majuscule (« Accent »). */
  const typeLabel = (t: string) => {
    const l = $_(`typing.errorTypes.${t}`);
    return l.charAt(0).toUpperCase() + l.slice(1);
  };

  const lang = $derived($locale ?? 'fr');
  const r = $derived(result.report);
  const avg = $derived(result.average);
  let showAll = $state(false);
  let removed: Record<string, boolean> = $state({});

  // les deux derniers types n'apparaissent que s'ils sont présents
  const types = $derived(
    TYPES.filter((t) => (t !== 'substitution' && t !== 'extra_word') || r.byType[t]),
  );
  const maxCount = $derived(Math.max(1, ...Object.values(r.byType)));
  const ops = $derived(showAll ? r.ops : r.ops.slice(0, PREVIEW_OPS));
  // mots à réviser : ceux qui ont un mot attendu
  const words = $derived(result.mistakes.filter((m) => m.expected));

  const headline = $derived(
    r.faults === 0
      ? $_('dictation.result.perfect', { values: { words: r.words } })
      : $_(r.faults <= r.words / 10 ? 'dictation.result.good' : 'dictation.result.keepGoing', {
          values: { n: r.faults, words: r.words },
        }),
  );

  function delta(v: number, unit: string, digits = 1): string {
    return (v >= 0 ? '+' : '−') + formatNumber(Math.abs(v), lang, digits) + unit;
  }

  async function remove(expected: string) {
    await RemoveWordByKey(expected);
    removed[expected] = true;
  }

  function train() {
    const list = words.map((m) => m.expected).join('|');
    router.navigate(`/review?session=shuffled&words=${encodeURIComponent(list)}`);
  }
</script>

<section class="flex flex-col gap-6" aria-labelledby="dictation-result-title">
  <header class="flex flex-wrap items-end justify-between gap-4">
    <div>
      <p class="eyebrow" style:color="var(--color-ok)">{$_('dictation.result.eyebrow')}</p>
      <h1 id="dictation-result-title" class="mt-1 text-3xl font-bold tracking-tight">{headline}</h1>
      <p class="text-muted mt-1 text-sm">{title}</p>
    </div>
    <div class="flex gap-3">
      <button class="btn btn-outline" onclick={onRedo}>{$_('dictation.result.redo')}</button>
      {#if words.length > 0}
        <button class="btn btn-primary" onclick={train}>
          {$_('dictation.result.train', { values: { n: words.length } })}
        </button>
      {/if}
    </div>
  </header>

  <dl class="grid grid-cols-2 gap-5 lg:grid-cols-4" data-testid="dictation-kpis">
    <div class="card p-5">
      <dt class="text-muted text-sm">{$_('dictation.result.accuracy')}</dt>
      <dd class="mt-1 text-3xl font-bold tabular-nums">{formatNumber(r.accuracy, lang, 1)} %</dd>
      {#if avg.sessions > 0}
        <dd
          class={[
            'mt-1 text-sm font-semibold',
            r.accuracy >= avg.accuracy ? 'text-ok' : 'text-error',
          ]}
        >
          {$_('dictation.result.vsAverage', {
            values: { d: delta(r.accuracy - avg.accuracy, ' pts') },
          })}
        </dd>
      {/if}
    </div>
    <div class="card p-5">
      <dt class="text-muted text-sm">{$_('dictation.result.faults')}</dt>
      <dd class={['mt-1 text-3xl font-bold tabular-nums', r.faults > 0 && 'text-error']}>
        {r.faults}
      </dd>
      {#if avg.sessions > 0}
        <dd class="text-muted mt-1 text-sm">
          {$_('dictation.result.average', { values: { v: formatNumber(avg.faults, lang) } })}
        </dd>
      {/if}
    </div>
    <div class="card p-5">
      <dt class="text-muted text-sm">{$_('dictation.result.speed')}</dt>
      <dd class="mt-1 text-3xl font-bold tabular-nums">{formatNumber(result.wpm, lang)} MPM</dd>
      {#if avg.sessions > 0}
        <dd class="text-muted mt-1 text-sm">
          {$_('dictation.result.average', { values: { v: formatNumber(avg.wpm, lang) } })}
        </dd>
      {/if}
    </div>
    <div class="card p-5">
      <dt class="text-muted text-sm">{$_('dictation.result.duration')}</dt>
      <dd class="mt-1 text-3xl font-bold tabular-nums">{formatDuration(result.durationMs)}</dd>
      <dd class="text-muted mt-1 text-sm">
        {$_('dictation.result.repetitions', { values: { n: result.repetitions } })}
      </dd>
    </div>
  </dl>

  <div class="grid items-start gap-6 lg:grid-cols-[1fr_19rem]">
    <article class="card p-7">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-bold">{$_('dictation.result.corrected')}</h2>
        <p class="flex gap-4 text-sm">
          <span class="text-error line-through">{$_('dictation.result.typedLegend')}</span>
          <span class="text-ok underline">{$_('dictation.result.expectedLegend')}</span>
        </p>
      </div>
      <p class="mt-5 font-mono text-lg leading-[2.4]" data-testid="corrected">
        {#each ops as o, i (i)}
          {#if o.correct}
            <span>{o.expected}</span>
          {:else}
            {#if o.typed}<del class="text-error me-1 text-base">{o.typed}</del>{/if}
            {#if o.expected}<ins
                class="bg-ok-soft text-ok border-ok/40 rounded border-b-2 px-1 font-semibold no-underline"
                title={$_(`typing.errorTypes.${o.type}`)}>{o.expected}</ins
              >{/if}
          {/if}{SPACE}
        {/each}
      </p>
      {#if r.ops.length > PREVIEW_OPS}
        <button class="text-accent mt-4 text-sm font-semibold" onclick={() => (showAll = !showAll)}>
          {showAll
            ? $_('dictation.result.showLess')
            : $_('dictation.result.showAll', { values: { n: r.words } })}
        </button>
      {/if}
    </article>

    <article class="card p-6" data-testid="by-type">
      <h2 class="text-lg font-bold">{$_('dictation.result.byType')}</h2>
      <ul class="mt-4 flex flex-col gap-3 text-sm">
        {#each types as t (t)}
          {@const n = r.byType[t] ?? 0}
          <li class="flex items-center gap-3">
            <span class={['w-32', n === 0 && 'text-muted']}>{typeLabel(t)}</span>
            <span class="bg-line h-1.5 flex-1 overflow-hidden rounded-full" aria-hidden="true">
              <span
                class="bg-accent block h-full rounded-full"
                style:width="{(100 * n) / maxCount}%"
              ></span>
            </span>
            <span class="w-5 text-right font-mono tabular-nums">{n}</span>
          </li>
        {/each}
      </ul>
    </article>
  </div>

  {#if words.length > 0}
    <article class="card p-7">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-bold">{$_('dictation.result.added')}</h2>
        <span class="text-ok flex items-center gap-1.5 text-sm font-semibold">
          <Icon d="M5 12l5 5L20 7" size={16} />
          {$_('dictation.result.tomorrow')}
        </span>
      </div>
      <table class="mt-4 w-full text-sm">
        <thead class="text-muted text-left text-xs tracking-wide uppercase">
          <tr class="border-line border-b">
            <th class="py-2 font-semibold">{$_('dictation.result.cols.expected')}</th>
            <th class="py-2 font-semibold">{$_('dictation.result.cols.typed')}</th>
            <th class="py-2 font-semibold">{$_('dictation.result.cols.type')}</th>
            <th class="py-2 text-right font-semibold">{$_('dictation.result.cols.action')}</th>
          </tr>
        </thead>
        <tbody>
          {#each words as m, i (i)}
            <tr class="border-line border-b last:border-0">
              <td class="py-3 font-mono font-semibold">{m.expected}</td>
              <td class="text-error py-3 font-mono line-through">{m.typed || '—'}</td>
              <td class="py-3">
                <span class="badge bg-amber-soft text-amber-ink">{typeLabel(m.type)}</span>
              </td>
              <td class="py-3 text-right">
                {#if removed[m.expected]}
                  <span class="text-muted">{$_('dictation.result.removed')}</span>
                {:else}
                  <button
                    class="text-muted hover:text-ink"
                    onclick={() => remove(m.expected)}
                    aria-label={$_('dictation.result.removeLabel', {
                      values: { word: m.expected },
                    })}>{$_('dictation.result.remove')}</button
                  >
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </article>
  {/if}
</section>
