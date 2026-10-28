<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { registry } from './modules';
  import { router } from './router.svelte';
  import Icon from '$shared/Icon.svelte';
  import { displayName, loadProfile, profile } from '$shared/profile.svelte';

  const active = $derived(registry.resolve(router.path));
  let badges: Record<string, number> = $state({});

  void loadProfile();

  // Pastilles du menu, rafraîchies à chaque changement de page.
  $effect(() => {
    void router.path;
    for (const m of registry.menu) {
      m.badge?.()
        .then((n) => (badges[m.id] = n))
        .catch(() => {});
    }
  });

  const name = $derived(displayName(profile.name));
  const initial = $derived((name || $_('app.you')).charAt(0).toUpperCase());
</script>

<div class="flex h-full">
  <nav
    aria-label={$_('app.mainNav')}
    class="bg-nav text-nav-ink flex w-60 shrink-0 flex-col gap-1 px-4 py-6"
  >
    <a href="#/" class="mb-6 flex items-center gap-3 px-2" aria-label={$_('app.name')}>
      <span
        class="bg-accent grid h-9 w-9 place-items-center rounded-lg font-mono text-sm font-bold text-white"
        aria-hidden="true">D_</span
      >
      <span class="text-nav-strong text-xl font-bold tracking-tight">{$_('app.name')}</span>
    </a>

    {#each registry.menu as m (m.id)}
      {@const current = active?.id === m.id}
      <a
        href={'#' + m.route}
        aria-current={current ? 'page' : undefined}
        class={[
          'flex items-center gap-3 rounded-xl px-3 py-2.5 text-[0.95rem] font-semibold transition-colors',
          current ? 'bg-nav-raised text-nav-strong' : 'hover:text-nav-strong hover:bg-white/5',
        ]}
      >
        <Icon d={m.icon} />
        <span class="flex-1">{$_(m.labelKey)}</span>
        {#if badges[m.id]}
          <span
            class="bg-warm text-warm-ink min-w-6 rounded-full px-1.5 text-center text-xs font-bold"
            aria-label={$_('app.badge', { values: { n: badges[m.id] } })}>{badges[m.id]}</span
          >
        {/if}
      </a>
    {/each}

    <a
      href="#/settings"
      class="bg-nav-raised/60 mt-auto flex items-center gap-3 rounded-xl px-3 py-3"
      aria-label={$_('app.profile')}
    >
      <span
        class="bg-warm text-warm-ink grid h-9 w-9 shrink-0 place-items-center rounded-full font-bold"
        aria-hidden="true">{initial}</span
      >
      <span class="min-w-0">
        <span class="text-nav-strong block truncate text-sm font-bold">
          {name || $_('app.you')}
        </span>
        <span class="block truncate text-xs">
          {$_(`app.layouts.${profile.keyboardLayout}`)} · {$_(`app.languages.${$locale ?? 'fr'}`)}
        </span>
      </span>
    </a>
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
