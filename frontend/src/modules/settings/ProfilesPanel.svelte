<!--
  Panneau Profils : liste, création, renommage, disposition, bascule et
  suppression des profils (F-50).
-->
<script lang="ts">
  import { _ } from 'svelte-i18n';
  import type { store } from '$wails/go/models';
  import {
    createProfile,
    deleteProfile,
    listProfiles,
    profile,
    switchProfile,
    updateProfile,
  } from '$shared/profile.svelte';

  // Profils locaux (F-50) : chacun a ses textes, statistiques, mots
  // difficiles et réglages ; la disposition de clavier sert au clavier
  // virtuel et à la carte de chaleur (F-08).
  const LAYOUTS = ['azerty', 'qwerty', 'bepo'] as const;

  let profiles = $state<store.Profile[]>([]);
  let error = $state('');
  let newName = $state('');
  let newLayout = $state<string>('azerty');
  let confirmDelete = $state(0);

  /** Recharge la liste des profils. */
  async function refresh() {
    try {
      profiles = await listProfiles();
    } catch (e) {
      error = String(e);
    }
  }
  void refresh();

  /** Exécute une action, affiche son erreur éventuelle et recharge. */
  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      await refresh();
    } catch (e) {
      error = String(e);
    }
  }

  /** Crée un profil avec le nom et la disposition saisis. */
  function add(e: SubmitEvent) {
    e.preventDefault();
    void run(async () => {
      await createProfile(newName, newLayout);
      newName = '';
    });
  }

  /** Supprime un profil (deux clics : confirmer). */
  function remove(id: number) {
    if (confirmDelete !== id) {
      confirmDelete = id;
      return;
    }
    confirmDelete = 0;
    void run(() => deleteProfile(id));
  }
</script>

<fieldset class="card p-5" data-testid="profiles">
  <legend class="px-1 font-semibold">{$_('settings.profiles.title')}</legend>
  <p class="text-muted mt-1 text-sm">{$_('settings.profiles.hint')}</p>

  <ul class="mt-4 flex flex-col gap-2">
    {#each profiles as p (p.id)}
      {@const current = p.id === profile.id}
      <li
        class={[
          'flex flex-wrap items-center gap-3 rounded-xl border px-4 py-3',
          current ? 'border-accent bg-accent-soft' : 'border-line',
        ]}
        data-testid="profile-row"
      >
        <input
          class="field w-48 py-1"
          value={p.name}
          maxlength="40"
          aria-label={$_('settings.profiles.name')}
          onchange={(e) => run(() => updateProfile(p.id, e.currentTarget.value, p.keyboardLayout))}
        />
        <select
          class="field py-1"
          value={p.keyboardLayout}
          aria-label={$_('settings.profiles.layout')}
          onchange={(e) => run(() => updateProfile(p.id, p.name, e.currentTarget.value))}
        >
          {#each LAYOUTS as l (l)}
            <option value={l}>{$_(`settings.layouts.${l}`)}</option>
          {/each}
        </select>
        <span class="flex-1"></span>
        {#if current}
          <span class="badge bg-accent text-white">{$_('settings.profiles.active')}</span>
        {:else}
          <button class="btn btn-soft py-1" onclick={() => run(() => switchProfile(p.id))}>
            {$_('settings.profiles.switch')}
          </button>
        {/if}
        {#if profiles.length > 1}
          <button
            class={['btn py-1', confirmDelete === p.id ? 'btn-warm' : 'btn-outline']}
            onclick={() => remove(p.id)}
          >
            {confirmDelete === p.id
              ? $_('settings.profiles.confirmDelete')
              : $_('settings.profiles.delete')}
          </button>
        {/if}
      </li>
    {/each}
  </ul>

  <form class="mt-4 flex flex-wrap items-end gap-2" onsubmit={add}>
    <label class="flex flex-col gap-2 text-sm">
      <span class="text-muted">{$_('settings.profiles.newName')}</span>
      <input class="field w-48" bind:value={newName} maxlength="40" required />
    </label>
    <label class="flex flex-col gap-2 text-sm">
      <span class="text-muted">{$_('settings.profiles.layout')}</span>
      <select class="field" bind:value={newLayout}>
        {#each LAYOUTS as l (l)}
          <option value={l}>{$_(`settings.layouts.${l}`)}</option>
        {/each}
      </select>
    </label>
    <button type="submit" class="btn btn-outline">{$_('settings.profiles.add')}</button>
  </form>

  {#if error}<p role="alert" class="text-error mt-3 text-sm">{error}</p>{/if}
</fieldset>
