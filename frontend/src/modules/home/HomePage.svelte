<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { Info } from '$wails/go/app/SystemService';

  const modes = ['copy', 'dictation', 'review'] as const;
  const info = Info();
</script>

<section class="mx-auto max-w-3xl px-8 py-12">
  <h1 class="text-3xl font-semibold">{$_('home.title')}</h1>
  <p class="text-muted mt-2">{$_('home.subtitle')}</p>

  <ul class="mt-10 grid gap-4 sm:grid-cols-3">
    {#each modes as mode (mode)}
      <li class="border-line bg-panel rounded-lg border p-5">
        <div class="flex items-start justify-between gap-2">
          <h2 class="font-semibold">{$_(`home.modes.${mode}.title`)}</h2>
          <span class="bg-line text-muted rounded-full px-2 py-0.5 text-xs">{$_('home.soon')}</span>
        </div>
        <p class="text-muted mt-2 text-sm">{$_(`home.modes.${mode}.desc`)}</p>
      </li>
    {/each}
  </ul>

  <p class="text-muted mt-12 text-xs" data-testid="system-info">
    {#await info then i}
      {$_('home.system', {
        values: { version: i.version, platform: i.platform, schema: i.schemaVersion },
      })}
    {:catch e}
      <span class="text-error">{$_('home.systemError', { values: { error: String(e) } })}</span>
    {/await}
  </p>
</section>
