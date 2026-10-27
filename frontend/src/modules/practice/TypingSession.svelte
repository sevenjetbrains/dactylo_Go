<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { FinishSession, RecordKeystrokes, StartSession } from '$wails/go/app/PracticeService';
  import type { app } from '$wails/go/models';
  import { Clock } from './clock';
  import { TypingEngine, wordEnds } from './engine';
  import { attachTypingInput } from './input';
  import { practiceSettings } from './settings.svelte';
  import LiveBar from './LiveBar.svelte';
  import ResultScreen from './ResultScreen.svelte';
  import TypingArea from './TypingArea.svelte';

  let {
    textId,
    title,
    content,
    onNext,
  }: { textId: number; title: string; content: string; onNext: () => void } = $props();

  const BATCH_MS = 2000; // envoi des frappes (section 3.2)
  const TICK_MS = 200;

  type Status = 'ready' | 'typing' | 'paused' | 'finishing' | 'done';

  // Réglages figés pour la durée de la session (Ctrl+R applique les nouveaux).
  let opts = $state.raw($state.snapshot(practiceSettings));
  let engine = $state.raw(makeEngine());
  let tick = $state(0);
  let status: Status = $state('ready');
  let elapsed = $state(0);
  let focused = $state(false);
  let error = $state('');
  let result: app.SessionResult | null = $state(null);

  let clock = new Clock();
  let sessionId: Promise<number> | null = null;
  let chain: Promise<void> = Promise.resolve();
  let textarea: HTMLTextAreaElement | undefined = $state();

  const limitMs = $derived(opts.sessionType === 'time' ? opts.durationMin * 60_000 : undefined);
  // Le moteur n'est pas réactif : `tick` signale chaque frappe.
  const live = $derived.by(() => {
    void tick;
    return {
      wpm: elapsed < 1000 ? 0 : engine.keystrokes / 5 / (elapsed / 60_000),
      accuracy: engine.accuracy,
      progress: engine.progress,
    };
  });

  function makeEngine(): TypingEngine {
    const end =
      opts.sessionType === 'words'
        ? wordEnds(Array.from(content.normalize('NFC')))[opts.wordCount - 1]
        : undefined;
    return new TypingEngine(content, {
      correction: opts.correction,
      allowBackspace: opts.allowBackspace,
      foldTypography: opts.foldTypography,
      end,
    });
  }

  function ensureSession(): Promise<number> {
    sessionId ??= StartSession(textId, JSON.stringify({ ...opts, startedFrom: 'practice' }));
    return sessionId;
  }

  /** Envoie les frappes en attente, dans l'ordre. */
  function flush(): Promise<void> {
    const batch = engine.take();
    if (batch.length > 0) {
      chain = chain
        .then(async () => RecordKeystrokes(await ensureSession(), batch))
        .catch((e: unknown) => {
          error = String(e);
        });
    }
    return chain;
  }

  function beforeKey(): boolean {
    if (status === 'finishing' || status === 'done') return false;
    if (status === 'ready') {
      clock.start();
      status = 'typing';
    } else if (status === 'paused') {
      clock.resume();
      status = 'typing';
    }
    return true;
  }

  function onChar(ch: string) {
    if (!beforeKey()) return;
    engine.type(ch, clock.elapsed());
    tick++;
    if (engine.done) void finish();
  }

  function onBackspace() {
    if (!beforeKey()) return;
    if (engine.backspace(clock.elapsed())) tick++;
  }

  async function finish() {
    if (status === 'finishing' || status === 'done') return;
    status = 'finishing';
    clock.pause();
    elapsed = clock.elapsed();
    await flush();
    if (!sessionId) {
      restart();
      return;
    }
    try {
      result = await FinishSession(await sessionId, Math.round(elapsed));
      status = 'done';
    } catch (e) {
      error = String(e);
      status = 'done';
    }
  }

  /** Recommence : la session en cours reste enregistrée mais non terminée. */
  function restart() {
    void flush();
    opts = $state.snapshot(practiceSettings);
    engine = makeEngine();
    clock = new Clock();
    sessionId = null;
    chain = Promise.resolve();
    result = null;
    error = '';
    elapsed = 0;
    tick++;
    status = 'ready';
    queueMicrotask(() => textarea?.focus());
  }

  function togglePause() {
    if (status === 'typing') {
      clock.pause();
      status = 'paused';
    } else if (status === 'paused') {
      clock.resume();
      status = 'typing';
      textarea?.focus();
    }
  }

  // Raccourcis F-07
  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      togglePause();
    } else if (e.ctrlKey && (e.key === 'r' || e.key === 'R')) {
      e.preventDefault(); // sinon la WebView recharge la page
      restart();
    } else if (e.ctrlKey && e.key === 'Enter') {
      e.preventDefault();
      void flush();
      onNext();
    }
  }

  // Attaché à l'élément lui-même : la zone de saisie est recréée après
  // l'écran de fin (Ctrl+R) et doit recevoir à nouveau les écouteurs.
  function typingInput(el: HTMLTextAreaElement) {
    const detach = attachTypingInput(el, { onChar, onBackspace });
    el.focus();
    return detach;
  }

  onMount(() => {
    const batchTimer = setInterval(() => void flush(), BATCH_MS);
    const tickTimer = setInterval(() => {
      if (status !== 'typing') return;
      elapsed = clock.elapsed();
      if (limitMs !== undefined && elapsed >= limitMs) void finish();
    }, TICK_MS);
    return () => {
      clearInterval(batchTimer);
      clearInterval(tickTimer);
    };
  });

  onDestroy(() => void flush());
</script>

<svelte:window onkeydown={onKeydown} />

<div class="flex flex-col gap-6">
  <header class="flex items-baseline justify-between gap-4">
    <h1 class="truncate text-xl font-semibold">{title}</h1>
    <p class="text-muted shrink-0 text-xs">{$_('practice.shortcuts')}</p>
  </header>

  {#if error}
    <p role="alert" class="text-error text-sm">{$_('practice.saveError', { values: { error } })}</p>
  {/if}

  {#if status === 'done' && result}
    <ResultScreen {result} onRestart={restart} {onNext} />
  {:else}
    {#if opts.showLiveBar}
      <LiveBar
        wpm={live.wpm}
        accuracy={live.accuracy}
        elapsedMs={elapsed}
        progress={live.progress}
        remainingMs={limitMs === undefined ? undefined : Math.max(0, limitMs - elapsed)}
      />
    {/if}

    <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
    <div
      class="border-line bg-panel relative cursor-text rounded-lg border px-6 py-5"
      onclick={() => textarea?.focus()}
    >
      <TypingArea
        {engine}
        {tick}
        fontSize={opts.fontSize}
        visibleLines={opts.visibleLines}
        linePosition={opts.linePosition}
      />

      {#if status === 'paused' || !focused}
        <div
          class="bg-panel/85 absolute inset-0 flex items-center justify-center rounded-lg text-sm font-medium"
        >
          {status === 'paused' ? $_('practice.paused') : $_('practice.clickToFocus')}
        </div>
      {/if}

      <textarea
        bind:this={textarea}
        {@attach typingInput}
        class="absolute h-px w-px opacity-0"
        aria-label={$_('practice.inputLabel')}
        autocomplete="off"
        autocapitalize="off"
        spellcheck="false"
        onfocus={() => (focused = true)}
        onblur={() => (focused = false)}></textarea>
    </div>

    {#if status === 'ready'}
      <p class="text-muted text-sm">{$_('practice.startHint')}</p>
    {:else if status !== 'finishing'}
      <div>
        <button class="text-muted text-sm underline" onclick={() => void finish()}>
          {$_('practice.actions.finish')}
        </button>
      </div>
    {/if}
  {/if}
</div>
