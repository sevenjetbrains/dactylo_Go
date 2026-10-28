<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { DeleteProfile, ListProfiles, SaveProfile } from '$wails/go/app/DictationService';
  import type { store } from '$wails/go/models';
  import { speech, type Voice } from '$shared/speech/engine';
  import { formatNumber } from '$shared/format';
  import { dictationSettings as s, type Chunking } from '$shared/dictation/settings.svelte';

  // Réglages de lecture (maquette « Mode Dictée », colonne de droite).
  let { started = false }: { started?: boolean } = $props();

  const CHUNKS: Chunking[] = ['word', 'group', 'sentence', 'paragraph'];
  const lang = $derived($locale ?? 'fr');

  let voices: Voice[] = $state([]);
  let profiles: store.DictationProfile[] = $state([]);
  let naming = $state(false);
  let newName = $state('');
  let error = $state('');

  void speech.voices().then((v) => {
    voices = v;
    if (!s.voice && v[0]) s.voice = v[0].id;
  });

  async function loadProfiles() {
    try {
      profiles = await ListProfiles();
    } catch (e) {
      error = String(e);
    }
  }
  void loadProfiles();

  // F-17 : un profil applique ses réglages
  function applyProfile(id: number) {
    s.profileId = id;
    const p = profiles.find((x) => x.id === id);
    if (!p) return;
    if (p.voice) s.voice = p.voice;
    s.rate = p.rate;
    s.pitch = p.pitch;
    s.volume = p.volume;
    s.chunking = p.chunking as Chunking;
    s.groupSize = p.groupSize;
    s.pauseMs = p.pauseMs;
    s.speakPunctuation = p.speakPunctuation;
    s.advance = p.advance as 'adaptive' | 'auto';
    s.maxRepeats = p.maxRepeats;
  }

  async function saveProfile(e: SubmitEvent) {
    e.preventDefault();
    try {
      const id = await SaveProfile({
        id: 0,
        name: newName,
        voice: s.voice,
        rate: s.rate,
        pitch: s.pitch,
        volume: s.volume,
        chunking: s.chunking,
        groupSize: s.groupSize,
        pauseMs: s.pauseMs,
        speakPunctuation: s.speakPunctuation,
        advance: s.advance,
        maxRepeats: s.maxRepeats,
      });
      await loadProfiles();
      s.profileId = id;
      naming = false;
      newName = '';
      error = '';
    } catch (err) {
      error = String(err);
    }
  }

  async function deleteProfile() {
    await DeleteProfile(s.profileId);
    s.profileId = 0;
    await loadProfiles();
  }

  const voiceLabel = (v: Voice) => {
    const region = v.lang.split('-')[1];
    return `${v.name}${region ? ` — ${v.lang}` : ''}`;
  };
</script>

<aside class="card flex flex-col gap-5 p-6" aria-label={$_('dictation.settings.title')}>
  <div class="flex items-center justify-between gap-3">
    <h2 class="text-lg font-bold">{$_('dictation.settings.title')}</h2>
    <select
      class="field py-1.5 text-sm"
      value={s.profileId}
      onchange={(e) => applyProfile(Number(e.currentTarget.value))}
      aria-label={$_('dictation.settings.profile')}
    >
      <option value={0}>{$_('dictation.settings.custom')}</option>
      {#each profiles as p (p.id)}
        <option value={p.id}>{p.name}</option>
      {/each}
    </select>
  </div>

  <label class="flex flex-col gap-1.5 text-sm">
    <span class="font-semibold">{$_('dictation.settings.voice')}</span>
    {#if voices.length > 0}
      <select class="field" bind:value={s.voice}>
        {#each voices as v (v.id)}
          <option value={v.id}>{voiceLabel(v)}</option>
        {/each}
      </select>
    {:else}
      <span class="text-error text-sm" role="alert">{$_('dictation.settings.noVoice')}</span>
    {/if}
  </label>

  <label class="flex flex-col gap-1.5 text-sm">
    <span class="flex justify-between font-semibold">
      {$_('dictation.settings.rate')}
      <span class="text-accent font-mono">{formatNumber(s.rate, lang, 1)}×</span>
    </span>
    <input
      type="range"
      min="0.5"
      max="2"
      step="0.1"
      bind:value={s.rate}
      class="accent-[var(--color-accent)]"
    />
  </label>

  <label class="flex flex-col gap-1.5 text-sm">
    <span class="flex justify-between font-semibold">
      {$_('dictation.settings.pause')}
      <span class="text-accent font-mono">{formatNumber(s.pauseMs / 1000, lang, 1)} s</span>
    </span>
    <input
      type="range"
      min="0"
      max="10000"
      step="500"
      bind:value={s.pauseMs}
      class="accent-[var(--color-accent)]"
    />
  </label>

  <fieldset class="flex flex-col gap-2 text-sm" disabled={started}>
    <legend class="mb-1.5 font-semibold">{$_('dictation.settings.chunking')}</legend>
    <div class="grid grid-cols-2 gap-2">
      {#each CHUNKS as c (c)}
        <button
          type="button"
          aria-pressed={s.chunking === c}
          class={[
            'rounded-xl border px-3 py-2.5 text-sm font-medium disabled:opacity-60',
            s.chunking === c ? 'border-accent bg-accent-soft text-accent' : 'border-line bg-panel',
          ]}
          onclick={() => (s.chunking = c)}
        >
          {c === 'group'
            ? $_('dictation.settings.chunks.group', { values: { n: s.groupSize } })
            : $_(`dictation.settings.chunks.${c}`)}
        </button>
      {/each}
    </div>
    {#if s.chunking === 'group'}
      <label class="text-muted flex items-center gap-2">
        {$_('dictation.settings.groupSize')}
        <input class="field w-20 py-1" type="number" min="1" max="20" bind:value={s.groupSize} />
      </label>
    {/if}
    {#if started}<p class="text-muted text-xs">{$_('dictation.settings.lockedHint')}</p>{/if}
  </fieldset>

  <fieldset class="flex flex-col gap-2 text-sm">
    <legend class="mb-1.5 font-semibold">{$_('dictation.settings.advance')}</legend>
    <label class="flex items-center gap-2">
      <input type="radio" name="advance" value="adaptive" bind:group={s.advance} />
      {$_('dictation.settings.adaptive')}
    </label>
    <label class="flex items-center gap-2">
      <input type="radio" name="advance" value="auto" bind:group={s.advance} />
      {$_('dictation.settings.auto')}
    </label>
  </fieldset>

  <label class="flex flex-col gap-1.5 text-sm">
    <span class="font-semibold">{$_('dictation.settings.repeats')}</span>
    <select class="field" bind:value={s.maxRepeats}>
      <option value={0}>{$_('dictation.settings.unlimited')}</option>
      {#each [1, 2, 3, 5] as n (n)}
        <option value={n}>{n}</option>
      {/each}
    </select>
  </label>

  <label class="border-line flex items-center justify-between gap-3 border-t pt-4 text-sm">
    <span>
      <span class="block font-semibold">{$_('dictation.settings.punctuation')}</span>
      <span class="text-muted text-xs">{$_('dictation.settings.punctuationHint')}</span>
    </span>
    <input
      type="checkbox"
      class="h-5 w-5 accent-[var(--color-accent)]"
      bind:checked={s.speakPunctuation}
      disabled={started}
    />
  </label>

  <details class="text-sm">
    <summary class="text-accent cursor-pointer font-semibold"
      >{$_('dictation.settings.more')}</summary
    >
    <div class="mt-3 flex flex-col gap-3">
      <label class="flex flex-col gap-1">
        <span class="flex justify-between">
          {$_('dictation.settings.pitch')}
          <span class="font-mono">{formatNumber(s.pitch, lang, 1)}</span>
        </span>
        <input type="range" min="0" max="2" step="0.1" bind:value={s.pitch} />
      </label>
      <label class="flex flex-col gap-1">
        <span class="flex justify-between">
          {$_('dictation.settings.volume')}
          <span class="font-mono">{Math.round(s.volume * 100)} %</span>
        </span>
        <input type="range" min="0" max="1" step="0.05" bind:value={s.volume} />
      </label>
    </div>
  </details>

  <div class="border-line flex flex-wrap items-center gap-2 border-t pt-4 text-sm">
    {#if naming}
      <form class="flex w-full gap-2" onsubmit={saveProfile}>
        <input
          class="field min-w-0 flex-1 py-1.5"
          bind:value={newName}
          placeholder={$_('dictation.settings.profileName')}
          aria-label={$_('dictation.settings.profileName')}
          required
        />
        <button type="submit" class="btn btn-primary px-3 py-1.5"
          >{$_('dictation.settings.save')}</button
        >
      </form>
    {:else}
      <button class="text-accent font-semibold" onclick={() => (naming = true)}>
        {$_('dictation.settings.saveAs')}
      </button>
      {#if s.profileId}
        <button class="text-muted ms-auto underline" onclick={deleteProfile}>
          {$_('dictation.settings.deleteProfile')}
        </button>
      {/if}
    {/if}
    {#if error}<p class="text-error w-full" role="alert">{error}</p>{/if}
  </div>
</aside>
