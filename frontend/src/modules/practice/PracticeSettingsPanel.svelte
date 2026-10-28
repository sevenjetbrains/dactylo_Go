<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { typingSettings as s } from '$shared/typing/settings.svelte';
</script>

<fieldset class="card p-5">
  <legend class="px-1 font-semibold">{$_('practice.settings.title')}</legend>

  <div class="mt-2 grid gap-5 sm:grid-cols-2">
    <label class="flex flex-col gap-2 text-sm">
      <span class="text-muted">
        {$_('practice.settings.fontSize', { values: { n: s.fontSize } })}
      </span>
      <input type="range" min="14" max="32" step="1" bind:value={s.fontSize} />
    </label>

    <label class="flex flex-col gap-2 text-sm">
      <span class="text-muted">
        {$_('practice.settings.visibleLines', { values: { n: s.visibleLines } })}
      </span>
      <input type="range" min="1" max="10" step="1" bind:value={s.visibleLines} />
    </label>

    <label class="flex flex-col gap-2 text-sm">
      <span class="text-muted">{$_('practice.settings.linePosition')}</span>
      <select class="field" bind:value={s.linePosition}>
        <option value="top">{$_('practice.settings.top')}</option>
        <option value="center">{$_('practice.settings.center')}</option>
      </select>
    </label>

    <label class="flex flex-col gap-2 text-sm">
      <span class="text-muted">{$_('practice.settings.correction')}</span>
      <select class="field" bind:value={s.correction}>
        <option value="free">{$_('practice.settings.free')}</option>
        <option value="blocking">{$_('practice.settings.blocking')}</option>
      </select>
    </label>

    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" bind:checked={s.allowBackspace} />
      {$_('practice.settings.allowBackspace')}
    </label>

    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" bind:checked={s.showKeyboard} />
      {$_('practice.settings.showKeyboard')}
    </label>

    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" bind:checked={s.showLiveBar} />
      {$_('practice.settings.showLiveBar')}
    </label>

    <label class="flex flex-col gap-2 text-sm sm:col-span-2">
      <span class="text-muted">
        {$_('practice.settings.slowThreshold', {
          values: { n: s.slowThreshold.toLocaleString($locale ?? 'fr') },
        })}
      </span>
      <input type="range" min="1.2" max="3" step="0.1" bind:value={s.slowThreshold} />
    </label>

    <label class="flex items-center gap-2 text-sm sm:col-span-2">
      <input type="checkbox" bind:checked={s.foldTypography} />
      {$_('practice.settings.foldTypography')}
    </label>
  </div>
</fieldset>
