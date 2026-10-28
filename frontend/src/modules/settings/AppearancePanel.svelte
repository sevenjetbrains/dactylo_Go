<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { LOCALES, setLocale, type Locale } from '$i18n/index';
  import { getTheme, setTheme, type Theme } from '$shared/theme';
  import { displayName, profile, renameProfile } from '$shared/profile.svelte';

  const themes: Theme[] = ['system', 'light', 'dark'];
  let theme = $state(getTheme());
  let name = $state(displayName(profile.name));
  let saved = $state(false);

  async function saveName(e: SubmitEvent) {
    e.preventDefault();
    await renameProfile(name);
    saved = true;
  }
</script>

<fieldset class="card p-5">
  <legend class="px-1 font-semibold">{$_('settings.appearance')}</legend>

  <div class="mt-2 flex flex-col gap-5">
    <form class="flex flex-wrap items-end gap-2" onsubmit={saveName}>
      <label class="flex flex-col gap-2 text-sm">
        <span class="text-muted">{$_('settings.name')}</span>
        <input
          class="field w-56"
          bind:value={name}
          maxlength="40"
          oninput={() => (saved = false)}
          placeholder={$_('settings.namePlaceholder')}
        />
      </label>
      <button type="submit" class="btn btn-outline">{$_('settings.save')}</button>
      {#if saved}<span class="text-ok text-sm" role="status">{$_('settings.saved')}</span>{/if}
    </form>

    <div role="radiogroup" aria-labelledby="theme-label">
      <p id="theme-label" class="text-muted mb-2 text-sm">{$_('settings.theme')}</p>
      <div class="flex gap-2">
        {#each themes as t (t)}
          <label
            class={[
              'cursor-pointer rounded-md border px-3 py-1.5 text-sm',
              theme === t ? 'border-accent bg-accent text-accent-ink' : 'border-line',
            ]}
          >
            <input
              type="radio"
              name="theme"
              value={t}
              class="sr-only"
              bind:group={theme}
              onchange={() => setTheme(t)}
            />
            {$_(`settings.themes.${t}`)}
          </label>
        {/each}
      </div>
    </div>

    <label class="flex flex-col gap-2 text-sm">
      <span class="text-muted">{$_('settings.language')}</span>
      <select
        class="field w-48"
        value={$locale}
        onchange={(e) => setLocale(e.currentTarget.value as Locale)}
      >
        {#each LOCALES as l (l)}
          <option value={l}>{$_(`settings.languages.${l}`)}</option>
        {/each}
      </select>
    </label>
  </div>
</fieldset>
