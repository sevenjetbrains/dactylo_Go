<!--
  Liste longue avec recherche (sans accents ni majuscules) et pagination.
  `fields` donne le texte cherché de chaque élément ; `row` affiche un élément.
-->
<script lang="ts" generics="T">
  import type { Snippet } from 'svelte';
  import { _ } from 'svelte-i18n';
  import { matches, paginate } from './list';
  import Pagination from './Pagination.svelte';

  let {
    items,
    fields,
    key,
    row,
    size = 20,
    label,
  }: {
    items: readonly T[];
    fields: (item: T) => string[];
    key: (item: T) => string | number;
    row: Snippet<[T]>;
    size?: number;
    /** Libellé du champ de recherche. */
    label?: string;
  } = $props();

  let query = $state('');
  let page = $state(0);

  const found = $derived(items.filter((it) => matches(query, ...fields(it))));
  const shown = $derived(paginate(found, page, size));

  /** Nouvelle recherche : retour à la première page. */
  function onSearch() {
    page = 0;
  }
</script>

{#if items.length > size}
  <label class="mb-3 block">
    <span class="sr-only">{label ?? $_('app.search')}</span>
    <input
      type="search"
      class="field w-full max-w-sm"
      placeholder={label ?? $_('app.search')}
      bind:value={query}
      oninput={onSearch}
    />
  </label>
{/if}

{#if shown.total === 0}
  <p class="text-muted">{$_('app.noMatch', { values: { query } })}</p>
{:else}
  <ul class="flex flex-col gap-3">
    {#each shown.items as it (key(it))}
      {@render row(it)}
    {/each}
  </ul>
  <Pagination
    bind:page={() => shown.page, (p) => (page = p)}
    pages={shown.pages}
    total={shown.total}
  />
{/if}
