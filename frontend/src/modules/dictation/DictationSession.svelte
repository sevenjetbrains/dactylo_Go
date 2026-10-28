<!--
  Session de dictée : lecture segment par segment, saisie libre, puis
  correction (F-10 à F-16).
-->
<script lang="ts">
  import { onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { Finish, Prepare, Start } from '$wails/go/app/DictationService';
  import type { app } from '$wails/go/models';
  import Icon from '$shared/Icon.svelte';
  import { speech } from '$shared/speech/engine';
  import { DictationPlayer } from './player.svelte';
  import { dictationSettings as s } from '$shared/dictation/settings.svelte';
  import DictationResult from './DictationResult.svelte';
  import DictationSettingsCard from './DictationSettingsCard.svelte';

  let {
    textId,
    title,
    start = 0,
    end = 0,
    onRedo,
  }: { textId: number; title: string; start?: number; end?: number; onRedo: () => void } = $props();

  const ICON = {
    play: 'M7 4l13 8-13 8z',
    pause: 'M9 5v14M15 5v14',
    prev: 'M6 5v14M19 5l-9 7 9 7z',
    repeat: 'M17 2l4 4-4 4M3 11V9a3 3 0 0 1 3-3h15M7 22l-4-4 4-4M21 13v2a3 3 0 0 1-3 3H3',
    eyeOff:
      'M3 3l18 18M10.6 10.6a2 2 0 0 0 2.8 2.8M9.9 5.1A10 10 0 0 1 12 5c5 0 9 4.5 10 7a13 13 0 0 1-3 4.2M6.2 6.2C4.2 7.6 2.8 9.6 2 12c1 2.5 5 7 10 7a9.8 9.8 0 0 0 4.3-1',
  };

  let plan = $state.raw<app.DictationPlan | null>(null);
  let player = $state.raw<DictationPlayer | null>(null);
  let sessionId: number | null = null;
  let started = $state(false);
  let typed = $state('');
  let finishing = $state(false);
  let result = $state.raw<app.DictationResult | null>(null);
  let error = $state('');

  const typedWords = $derived(typed.trim() ? typed.trim().split(/\s+/).length : 0);

  // Découpage préparé par le cœur Go ; recalculé tant que la dictée n'a pas commencé.
  $effect(() => {
    const req = {
      textId,
      start,
      end,
      chunking: s.chunking,
      groupSize: s.groupSize,
      speakPunctuation: s.speakPunctuation,
    };
    if (started) return;
    Prepare(req)
      .then((p) => {
        plan = p;
        player = new DictationPlayer(p.chunks, speech, {
          speak: () => ({ voiceId: s.voice, rate: s.rate, pitch: s.pitch, volume: s.volume }),
          advance: () => s.advance,
          pauseMs: () => s.pauseMs,
          maxRepeats: () => s.maxRepeats,
        });
      })
      .catch((e: unknown) => (error = String(e)));
  });

  $effect(() => {
    player?.notifyTyped(typedWords);
  });

  /** Crée la session côté backend à la première lecture. */
  async function ensureSession(): Promise<number> {
    sessionId ??= await Start({ textId, start, end, content: '', settings: JSON.stringify(s) });
    return sessionId;
  }

  /** Lecture ou pause de la voix. */
  async function toggle() {
    if (!player) return;
    if (!started) {
      started = true;
      try {
        await ensureSession();
      } catch (e) {
        error = String(e);
        return;
      }
    }
    player.toggle();
  }

  /** Envoie la dictée tapée et affiche la correction. */
  async function finish() {
    if (!player || finishing) return;
    finishing = true;
    player.stop();
    try {
      const id = await ensureSession();
      result = await Finish({
        sessionId: id,
        typed,
        durationMs: Math.round(player.elapsed()),
        repetitions: player.totalRepeats,
      });
    } catch (e) {
      error = String(e);
    } finally {
      finishing = false;
    }
  }

  // F-14 : Ctrl+Espace répète le segment
  function onKeydown(e: KeyboardEvent) {
    if (e.ctrlKey && e.code === 'Space') {
      e.preventDefault();
      player?.repeat();
    }
  }

  onDestroy(() => player?.stop());

  const statusKey = $derived(
    !player || player.status === 'idle'
      ? 'ready'
      : player.status === 'waiting' && player.index >= player.count - 1
        ? 'allRead'
        : player.status,
  );
</script>

<svelte:window onkeydown={onKeydown} />

<div class="flex flex-col gap-6">
  {#if result}
    <DictationResult {result} {title} {onRedo} />
  {:else}
    <header>
      <p class="eyebrow">
        {$_('dictation.eyebrow', { values: { n: plan?.words ?? 0 } })}
      </p>
      <h1 class="mt-1 text-3xl font-bold tracking-tight">{title}</h1>
    </header>

    {#if error}<p role="alert" class="text-error text-sm">{error}</p>{/if}

    <div class="grid items-start gap-6 lg:grid-cols-[1fr_22rem]">
      <div class="flex flex-col gap-6">
        <!-- Lecteur -->
        <section
          class="bg-nav rounded-2xl p-7 text-white"
          aria-label={$_('dictation.player.label')}
        >
          <div class="flex items-center gap-6">
            <button
              class="bg-warm text-warm-ink grid h-18 w-18 shrink-0 place-items-center rounded-full shadow-lg disabled:opacity-50"
              onclick={toggle}
              disabled={!player || finishing}
              aria-label={player?.status === 'speaking' || player?.status === 'waiting'
                ? $_('dictation.player.pause')
                : $_('dictation.player.play')}
            >
              <Icon
                d={player?.status === 'speaking' || player?.status === 'waiting'
                  ? ICON.pause
                  : ICON.play}
                size={28}
              />
            </button>
            <div class="min-w-0 flex-1">
              <div class="text-nav-ink flex flex-wrap justify-between gap-2 text-sm">
                <span data-testid="segment-status">
                  {$_('dictation.player.segment', {
                    values: { n: (player?.index ?? 0) + 1, total: player?.count ?? 0 },
                  })} · {$_(`dictation.player.status.${statusKey}`)}
                </span>
                <span>
                  {$_('dictation.player.repeats', {
                    values: {
                      n: player?.repeats ?? 0,
                      max: s.maxRepeats || $_('dictation.player.unlimited'),
                    },
                  })}
                </span>
              </div>
              <ol class="mt-3 flex gap-1" aria-hidden="true">
                {#each Array.from({ length: player?.count ?? 0 }, (_, i) => i) as i (i)}
                  <li
                    class="h-1.5 flex-1 rounded-full"
                    style:background={i < (player?.index ?? 0)
                      ? '#9aa6f5'
                      : i === player?.index && started
                        ? 'var(--color-warm)'
                        : 'var(--color-nav-raised)'}
                  ></li>
                {/each}
              </ol>
            </div>
          </div>
          <div class="mt-6 flex flex-wrap gap-3">
            <button
              class="btn border border-white/20 text-white hover:bg-white/10"
              onclick={() => player?.previous()}
              disabled={!started || (player?.index ?? 0) === 0}
            >
              <Icon d={ICON.prev} size={16} />
              {$_('dictation.player.previous')}
            </button>
            <button
              class="btn border border-white/20 text-white hover:bg-white/10"
              onclick={() => player?.repeat()}
              disabled={!started || !player?.canRepeat}
            >
              <Icon d={ICON.repeat} size={16} />
              {$_('dictation.player.repeat')}
              <kbd class="font-mono text-xs opacity-70">Ctrl Espace</kbd>
            </button>
          </div>
        </section>

        <!-- Copie -->
        <section class="card flex flex-col gap-3 p-6">
          <div class="flex items-center justify-between">
            <label for="dictation-copy" class="font-bold">{$_('dictation.copy.title')}</label>
            <span class="text-muted flex items-center gap-1.5 text-sm">
              <Icon d={ICON.eyeOff} size={16} />
              {$_('dictation.copy.hidden')}
            </span>
          </div>
          <textarea
            id="dictation-copy"
            class="field bg-sunken focus:border-accent min-h-64 resize-y p-5 font-mono text-xl leading-relaxed outline-none"
            bind:value={typed}
            spellcheck="false"
            autocomplete="off"
            autocapitalize="off"></textarea>
          <div class="flex flex-wrap items-center justify-between gap-3">
            <p class="text-muted text-sm" data-testid="copy-status">
              {$_('dictation.copy.typed', { values: { n: typedWords } })} ·
              {s.advance === 'adaptive'
                ? $_('dictation.copy.adaptiveHint', {
                    values: { n: Math.max(0, (player?.wordsThroughCurrent ?? 0) - typedWords) },
                  })
                : $_('dictation.copy.autoHint', { values: { s: s.pauseMs / 1000 } })}
            </p>
            <button class="btn btn-primary" onclick={finish} disabled={!player || finishing}>
              {$_('dictation.copy.finish')}
            </button>
          </div>
        </section>
      </div>

      <DictationSettingsCard {started} />
    </div>
  {/if}
</div>
