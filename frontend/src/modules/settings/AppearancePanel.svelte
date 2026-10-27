<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { LOCALES, setLocale, type Locale } from '$i18n/index';
  import { getTheme, setTheme, type Theme } from '$shared/theme';

  const themes: Theme[] = ['system', 'light', 'dark'];
  let theme = $state(getTheme());
</script>

<fieldset class="border-line bg-panel rounded-lg border p-5">
  <legend class="px-1 font-semibold">{$_('settings.appearance')}</legend>

  <div class="mt-2 flex flex-col gap-5">
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
        class="border-line bg-panel w-48 rounded-md border px-3 py-1.5"
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
