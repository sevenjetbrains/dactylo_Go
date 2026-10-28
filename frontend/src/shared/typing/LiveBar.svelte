<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { formatDuration, formatNumber } from '$shared/format';

  let {
    wpm,
    accuracy,
    elapsedMs,
    errors,
    progress,
    remainingMs,
  }: {
    wpm: number;
    accuracy: number;
    elapsedMs: number;
    errors: number;
    progress: number;
    /** Temps restant pour une session à durée fixe. */
    remainingMs?: number;
  } = $props();

  const lang = $derived($locale ?? 'fr');
  const time = $derived(formatDuration(remainingMs ?? elapsedMs).padStart(5, '0'));
</script>

<dl class="card flex flex-wrap items-center gap-x-10 gap-y-3 px-7 py-4" aria-live="off">
  <div class="flex items-baseline gap-2">
    <dd class="font-mono text-3xl font-semibold tabular-nums">{formatNumber(wpm, lang)}</dd>
    <dt class="text-muted text-sm">{$_('typing.live.wpm')}</dt>
  </div>
  <div class="flex items-baseline gap-2">
    <dd class="font-mono text-3xl font-semibold tabular-nums">{formatNumber(accuracy, lang, 1)}</dd>
    <dt class="text-muted text-sm">{$_('typing.live.accuracy')}</dt>
  </div>
  <div class="flex items-baseline gap-2">
    <dd class="font-mono text-3xl font-semibold tabular-nums">{time}</dd>
    <dt class="text-muted text-sm">
      {remainingMs === undefined ? $_('typing.live.elapsed') : $_('typing.live.remaining')}
    </dt>
  </div>
  <div class="flex items-baseline gap-2">
    <dd class={['font-mono text-3xl font-semibold tabular-nums', errors > 0 && 'text-error']}>
      {errors}
    </dd>
    <dt class="text-muted text-sm">{$_('typing.live.errors', { values: { n: errors } })}</dt>
  </div>
  <div class="flex min-w-40 flex-1 items-center gap-4">
    <dt class="sr-only">{$_('typing.live.progress')}</dt>
    <dd class="flex flex-1 items-center gap-4">
      <div
        class="bg-line h-2 flex-1 overflow-hidden rounded-full"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(progress * 100)}
        aria-label={$_('typing.live.progress')}
      >
        <div class="bg-accent h-full rounded-full" style:width="{progress * 100}%"></div>
      </div>
      <span class="text-sm font-semibold tabular-nums">{Math.round(progress * 100)} %</span>
    </dd>
  </div>
</dl>
