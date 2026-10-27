<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { formatDuration, formatNumber } from '$shared/format';

  let {
    wpm,
    accuracy,
    elapsedMs,
    progress,
    remainingMs,
  }: {
    wpm: number;
    accuracy: number;
    elapsedMs: number;
    progress: number;
    /** Temps restant pour une session à durée fixe. */
    remainingMs?: number;
  } = $props();

  const lang = $derived($locale ?? 'fr');
</script>

<dl class="flex flex-wrap gap-x-8 gap-y-2 text-sm" aria-live="off">
  <div>
    <dt class="text-muted">{$_('typing.live.wpm')}</dt>
    <dd class="text-xl font-semibold tabular-nums">{formatNumber(wpm, lang)}</dd>
  </div>
  <div>
    <dt class="text-muted">{$_('typing.live.accuracy')}</dt>
    <dd class="text-xl font-semibold tabular-nums">{formatNumber(accuracy, lang)} %</dd>
  </div>
  <div>
    <dt class="text-muted">
      {remainingMs === undefined ? $_('typing.live.elapsed') : $_('typing.live.remaining')}
    </dt>
    <dd class="text-xl font-semibold tabular-nums">
      {formatDuration(remainingMs ?? elapsedMs)}
    </dd>
  </div>
  <div class="min-w-32 flex-1">
    <dt class="text-muted">{$_('typing.live.progress')}</dt>
    <dd class="mt-2 flex items-center gap-3">
      <div
        class="bg-line h-2 flex-1 overflow-hidden rounded-full"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(progress * 100)}
      >
        <div class="bg-accent h-full" style:width="{progress * 100}%"></div>
      </div>
      <span class="tabular-nums">{Math.round(progress * 100)} %</span>
    </dd>
  </div>
</dl>
