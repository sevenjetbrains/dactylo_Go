<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { FinishSession, RecordKeystrokes, StartSession } from '$wails/go/app/PracticeService';
  import type { app } from '$wails/go/models';
  import { Clock } from './clock';
  import { TypingEngine, wordEnds } from './engine';
  import { attachTypingInput } from './input';
  import { typingSettings } from './settings.svelte';
  import LiveBar from './LiveBar.svelte';
  import ResultScreen from './ResultScreen.svelte';
  import TypingArea from './TypingArea.svelte';
  import Icon from '$shared/Icon.svelte';
  import VirtualKeyboard from '$shared/keyboard/VirtualKeyboard.svelte';
  import { DURATIONS } from './settings.svelte';

  let {
    textId = 0,
    title,
    eyebrow = '',
    start = 0,
    end = 0,
    content,
    onNext,
    back,
  }: {
    /** Texte de la bibliothèque ; 0 = exercice généré (révision), tapé en entier. */
    textId?: number;
    title: string;
    /** Sur-titre (« Copie · exercice 3 sur 8 »). */
    eyebrow?: string;
    /** Bornes de l'exercice dans le texte (en caractères) ; content = texte[start, end). */
    start?: number;
    end?: number;
    content: string;
    onNext: () => void;
    back?: { href: string; label: string };
  } = $props();

  const BATCH_MS = 2000; // envoi des frappes (section 3.2)
  const TICK_MS = 200;

  type Status = 'ready' | 'typing' | 'paused' | 'finishing' | 'done';

  // Réglages figés pour la durée de la session (Ctrl+R applique les nouveaux).
  let opts = $state.raw(snapshotSettings());
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
      errors: engine.keystrokes - engine.correctKeystrokes,
      next: engine.done ? null : (engine.chars[engine.position] ?? null),
    };
  });

  // Choix rapide du type de session (maquette « Mode Copie », F-06).
  type Preset = { id: string; apply: () => void; active: boolean };
  const presets = $derived<Preset[]>([
    {
      id: 'text',
      apply: () => (typingSettings.sessionType = 'text'),
      active: opts.sessionType === 'text',
    },
    ...DURATIONS.filter((d) => d !== 10).map((d) => ({
      id: `time${d}`,
      apply: () => {
        typingSettings.sessionType = 'time';
        typingSettings.durationMin = d;
      },
      active: opts.sessionType === 'time' && opts.durationMin === d,
    })),
    {
      id: 'words100',
      apply: () => {
        typingSettings.sessionType = 'words';
        typingSettings.wordCount = 100;
      },
      active: opts.sessionType === 'words' && opts.wordCount === 100,
    },
  ]);

  /** Applique un réglage de la barre d'outils et recommence avec lui. */
  function choose(apply: () => void) {
    apply();
    restart();
  }

  /** Un exercice généré se tape toujours en entier. */
  function snapshotSettings() {
    const s = $state.snapshot(typingSettings);
    return textId === 0 ? { ...s, sessionType: 'text' as const } : s;
  }

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
    sessionId ??= StartSession({
      textId,
      start,
      end,
      content: textId === 0 ? content : '',
      settings: JSON.stringify(opts),
    });
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
    opts = snapshotSettings();
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

<div class="flex flex-col gap-5">
  <header>
    {#if eyebrow}<p class="eyebrow">{eyebrow}</p>{/if}
    <h1 class="mt-1 truncate text-3xl font-bold tracking-tight">{title}</h1>
  </header>

  {#if error}
    <p role="alert" class="text-error text-sm">{$_('typing.saveError', { values: { error } })}</p>
  {/if}

  {#if status === 'done' && result}
    <ResultScreen {result} onRestart={restart} {onNext} {back} />
  {:else}
    <!-- Barre d'outils : type de session, correction, pause, recommencer -->
    <div class="flex flex-wrap items-center gap-3">
      {#if textId !== 0}
        <div
          class="bg-line/70 flex rounded-xl p-1"
          role="radiogroup"
          aria-label={$_('typing.toolbar.sessionType')}
        >
          {#each presets as p (p.id)}
            <button
              role="radio"
              aria-checked={p.active}
              class={[
                'rounded-lg px-3.5 py-1.5 text-sm font-semibold',
                p.active ? 'bg-panel text-ink shadow-sm' : 'text-muted hover:text-ink',
              ]}
              onclick={() => choose(p.apply)}
            >
              {$_(`typing.toolbar.presets.${p.id}`)}
            </button>
          {/each}
        </div>
      {/if}
      <label class="text-muted flex items-center gap-2 text-sm">
        {$_('typing.toolbar.correction')}
        <select
          class="field text-ink w-36"
          value={opts.correction}
          onchange={(e) =>
            choose(
              () => (typingSettings.correction = e.currentTarget.value as 'free' | 'blocking'),
            )}
        >
          <option value="free">{$_('typing.toolbar.free')}</option>
          <option value="blocking">{$_('typing.toolbar.blocking')}</option>
        </select>
      </label>
      <button
        class="btn btn-outline h-11 w-11 p-0"
        onclick={togglePause}
        disabled={status !== 'typing' && status !== 'paused'}
        aria-label={status === 'paused' ? $_('typing.toolbar.resume') : $_('typing.toolbar.pause')}
      >
        <Icon d={status === 'paused' ? 'M7 4l13 8-13 8z' : 'M9 5v14M15 5v14'} size={18} />
      </button>
      <button
        class="btn btn-outline h-11 w-11 p-0"
        onclick={restart}
        aria-label={$_('typing.actions.restart')}
      >
        <Icon d="M3 12a9 9 0 1 0 3-6.7L3 8M3 3v5h5" size={18} />
      </button>
    </div>

    {#if opts.showLiveBar}
      <LiveBar
        wpm={live.wpm}
        accuracy={live.accuracy}
        elapsedMs={elapsed}
        errors={live.errors}
        progress={live.progress}
        remainingMs={limitMs === undefined ? undefined : Math.max(0, limitMs - elapsed)}
      />
    {/if}

    <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
    <div
      class={[
        'bg-panel relative cursor-text rounded-2xl border-2 px-10 py-8 transition-shadow',
        focused ? 'border-accent shadow-[0_0_0_4px_var(--color-accent-soft)]' : 'border-line',
      ]}
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
          class="bg-panel/85 absolute inset-0 flex items-center justify-center rounded-2xl text-sm font-semibold"
        >
          {status === 'paused' ? $_('typing.paused') : $_('typing.clickToFocus')}
        </div>
      {/if}

      <textarea
        bind:this={textarea}
        {@attach typingInput}
        class="absolute h-px w-px opacity-0"
        aria-label={$_('typing.inputLabel')}
        autocomplete="off"
        autocapitalize="off"
        spellcheck="false"
        onfocus={() => (focused = true)}
        onblur={() => (focused = false)}></textarea>
    </div>

    <div class="text-muted flex flex-wrap items-center gap-x-5 gap-y-2 text-sm">
      <span><kbd class="kbd">Échap</kbd> {$_('typing.keys.pause')}</span>
      <span><kbd class="kbd">Ctrl R</kbd> {$_('typing.keys.restart')}</span>
      <span><kbd class="kbd">Ctrl ↵</kbd> {$_('typing.keys.next')}</span>
      {#if status === 'ready'}
        <span>{$_('typing.startHint')}</span>
      {:else if status !== 'finishing'}
        <button class="underline" onclick={() => void finish()}>
          {$_('typing.actions.finish')}
        </button>
      {/if}
      <span class="ms-auto text-xs">
        {$_('typing.display', { values: { lines: opts.visibleLines, size: opts.fontSize } })}
      </span>
    </div>

    {#if opts.showKeyboard}
      <VirtualKeyboard next={live.next} />
    {/if}
  {/if}
</div>
