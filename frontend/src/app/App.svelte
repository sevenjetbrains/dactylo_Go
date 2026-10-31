<!--
  Coquille de l'application : menu latéral (modules du registre, pastilles,
  profil actif) et page de la route courante.
-->
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
    class="bg-nav text-nav-ink flex w-60 shrink-0 flex-col gap-1 overflow-y-auto px-4 py-6 max-lg:w-18 max-lg:px-2.5"
  >
    <!-- Fenêtre étroite (< 1024 px) : menu réduit aux icônes, libellés en infobulle. -->
    <a
      href="#/"
      class="mb-6 flex items-center gap-3 px-2 max-lg:justify-center max-lg:px-0"
      aria-label={$_('app.name')}
    >
      <span
        class="bg-accent grid h-9 w-9 place-items-center rounded-lg font-mono text-sm font-bold text-white"
        aria-hidden="true">D_</span
      >
      <span class="text-nav-strong text-xl font-bold tracking-tight max-lg:sr-only"
        >{$_('app.name')}</span
      >
    </a>

    {#each registry.menu as m (m.id)}
      {@const current = active?.id === m.id}
      <a
        href={'#' + m.route}
        aria-current={current ? 'page' : undefined}
        title={$_(m.labelKey)}
        class={[
          'relative flex items-center gap-3 rounded-xl px-3 py-2.5 text-[0.95rem] font-semibold transition-colors max-lg:justify-center',
          current ? 'bg-nav-raised text-nav-strong' : 'hover:text-nav-strong hover:bg-white/5',
        ]}
      >
        <Icon d={m.icon} />
        <span class="flex-1 max-lg:sr-only">{$_(m.labelKey)}</span>
        {#if badges[m.id]}
          <span
            class="bg-warm text-warm-ink min-w-6 rounded-full px-1.5 text-center text-xs font-bold max-lg:absolute max-lg:-top-1 max-lg:-right-1 max-lg:min-w-5 max-lg:px-1 max-lg:text-[0.65rem]"
            aria-label={$_('app.badge', { values: { n: badges[m.id] } })}>{badges[m.id]}</span
          >
        {/if}
      </a>
    {/each}

    <a
      href="#/settings"
      class="bg-nav-raised/60 mt-auto flex items-center gap-3 rounded-xl px-3 py-3 max-lg:justify-center max-lg:px-0"
      aria-label={$_('app.profile')}
      title={name || $_('app.you')}
    >
      <span
        class="bg-warm text-warm-ink grid h-9 w-9 shrink-0 place-items-center rounded-full font-bold"
        aria-hidden="true">{initial}</span
      >
      <span class="min-w-0 max-lg:sr-only">
        <span class="text-nav-strong block truncate text-sm font-bold">
          {name || $_('app.you')}
        </span>
        <span class="block truncate text-xs">
          {$_(`app.layouts.${profile.keyboardLayout}`)} · {$_(`app.languages.${$locale ?? 'fr'}`)}
        </span>
      </span>
    </a>
  </nav>

  <!-- relative : les éléments positionnés restent dans la zone défilante -->
  <main class="relative min-w-0 flex-1 overflow-y-auto">
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

{#if router.pending}
  <!-- Confirmation avant de quitter une session en cours. Les touches ne
       remontent pas jusqu'à la session (Échap y met en pause). -->
  <div class="fixed inset-0 z-50 grid place-items-center bg-black/50 p-6">
    <div
      class="card w-full max-w-md p-6 shadow-xl"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="leave-title"
      aria-describedby="leave-message"
      tabindex="-1"
      onkeydown={(e) => {
        e.stopPropagation();
        if (e.key === 'Escape') router.stay();
      }}
    >
      <h2 id="leave-title" class="text-lg font-bold">{$_('app.leave.title')}</h2>
      <p id="leave-message" class="text-muted mt-2 text-sm">{router.pending.message}</p>
      <div class="mt-6 flex justify-end gap-2">
        <button class="btn btn-outline" onclick={() => router.leave()}>
          {$_('app.leave.leave')}
        </button>
        <button class="btn btn-primary" {@attach (el) => el.focus()} onclick={() => router.stay()}>
          {$_('app.leave.stay')}
        </button>
      </div>
    </div>
  </div>
{/if}
