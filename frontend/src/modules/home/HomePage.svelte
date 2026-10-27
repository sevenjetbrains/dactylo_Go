<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { Info } from '$wails/go/app/SystemService';

  // href absent = mode pas encore disponible
  const modes = [
    { id: 'copy', href: '#/practice' },
    { id: 'dictation', href: undefined },
    { id: 'review', href: undefined },
  ] as const;
  const info = Info();
</script>

<section class="mx-auto max-w-3xl px-8 py-12">
  <h1 class="text-3xl font-semibold">{$_('home.title')}</h1>
  <p class="text-muted mt-2">{$_('home.subtitle')}</p>

  <ul class="mt-10 grid gap-4 sm:grid-cols-3">
    {#each modes as mode (mode.id)}
      <li class="relative">
        <svelte:element
          this={mode.href ? 'a' : 'div'}
          href={mode.href}
          class={[
            'border-line bg-panel block h-full rounded-lg border p-5',
            mode.href && 'hover:border-accent',
          ]}
        >
          <div class="flex items-start justify-between gap-2">
            <h2 class="font-semibold">{$_(`home.modes.${mode.id}.title`)}</h2>
            {#if !mode.href}
              <span class="bg-line text-muted rounded-full px-2 py-0.5 text-xs">
                {$_('home.soon')}
              </span>
            {/if}
          </div>
          <p class="text-muted mt-2 text-sm">{$_(`home.modes.${mode.id}.desc`)}</p>
        </svelte:element>
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
