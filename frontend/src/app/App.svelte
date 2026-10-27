<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { registry } from './modules';
  import { router } from './router.svelte';
  import Icon from '$shared/Icon.svelte';

  const active = $derived(registry.resolve(router.path));
</script>

<div class="flex h-full">
  <nav
    aria-label={$_('app.mainNav')}
    class="border-line bg-panel flex w-56 shrink-0 flex-col gap-1 border-e p-3"
  >
    <div class="mb-4 px-3 pt-2">
      <p class="text-lg font-semibold">{$_('app.name')}</p>
      <p class="text-muted text-sm">{$_('app.tagline')}</p>
    </div>
    {#each registry.menu as m (m.id)}
      {@const current = active?.id === m.id}
      <a
        href={'#' + m.route}
        aria-current={current ? 'page' : undefined}
        class={[
          'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
          current ? 'bg-accent text-accent-ink' : 'text-muted hover:bg-line/60 hover:text-ink',
        ]}
      >
        <Icon d={m.icon} />
        {$_(m.labelKey)}
      </a>
    {/each}
  </nav>

  <main class="min-w-0 flex-1 overflow-y-auto">
    {#if active}
      {@const Page = active.page}
      <Page />
    {:else}
      <div class="p-10">
        <p class="mb-4">{$_('app.notFound')}</p>
        <a href="#/" class="text-accent underline">{$_('app.backHome')}</a>
      </div>
    {/if}
  </main>
</div>
