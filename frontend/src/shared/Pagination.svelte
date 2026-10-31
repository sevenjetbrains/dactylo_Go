<!--
  Navigation entre les pages d'une liste longue : Précédent, numéros de
  page (1 … 4 5 6 … 15), Suivant et position. Masquée quand tout tient sur
  une page.
-->
<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { pageNumbers } from './list';

  let {
    page = $bindable(0),
    pages,
    total,
  }: { page?: number; pages: number; total: number } = $props();

  const numbers = $derived(pageNumbers(page, pages));
</script>

{#if pages > 1}
  <nav
    class="text-muted mt-4 flex flex-wrap items-center justify-between gap-3 text-sm"
    aria-label={$_('app.pagination.label')}
  >
    <span>
      {$_('app.pagination.position', { values: { page: page + 1, pages, total } })}
    </span>
    <div class="flex flex-wrap items-center gap-1">
      <button class="btn btn-outline px-3 py-1.5" disabled={page === 0} onclick={() => page--}>
        {$_('app.pagination.previous')}
      </button>
      {#each numbers as n, i (n ?? `gap-${i}`)}
        {#if n === null}
          <span class="px-1" aria-hidden="true">…</span>
        {:else}
          <button
            class={[
              'btn min-w-9 px-2 py-1.5 tabular-nums',
              n === page ? 'btn-primary' : 'btn-outline',
            ]}
            aria-current={n === page ? 'page' : undefined}
            aria-label={$_('app.pagination.goTo', { values: { page: n + 1 } })}
            onclick={() => (page = n)}
          >
            {n + 1}
          </button>
        {/if}
      {/each}
      <button
        class="btn btn-outline px-3 py-1.5"
        disabled={page >= pages - 1}
        onclick={() => page++}
      >
        {$_('app.pagination.next')}
      </button>
    </div>
  </nav>
{/if}
