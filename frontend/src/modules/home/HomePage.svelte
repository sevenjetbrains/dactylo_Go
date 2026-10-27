<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { Info } from '$wails/go/app/SystemService';
  import { Summary } from '$wails/go/app/ReviewService';

  // href absent = mode pas encore disponible
  const modes = [
    { id: 'copy', href: '#/practice' },
    { id: 'dictation', href: undefined },
    { id: 'review', href: '#/review' },
  ] as const;
  const info = Info();
  const review = Summary();
</script>

<section class="mx-auto max-w-3xl px-8 py-12">
  <h1 class="text-3xl font-semibold">{$_('home.title')}</h1>
  <p class="text-muted mt-2">{$_('home.subtitle')}</p>

  <!-- F-43 : les mots à réviser, en un clic -->
  {#await review then r}
    {#if r.due > 0}
      <a
        href="#/review?session=shuffled"
        class="border-accent bg-panel mt-8 flex items-center justify-between gap-4 rounded-lg border-2 p-5"
        data-testid="due-banner"
      >
        <span class="text-lg font-semibold">{$_('home.due', { values: { n: r.due } })}</span>
        <span class="bg-accent text-accent-ink rounded-md px-4 py-2 text-sm font-medium">
          {$_('home.reviewNow')}
        </span>
      </a>
    {/if}
  {/await}

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
