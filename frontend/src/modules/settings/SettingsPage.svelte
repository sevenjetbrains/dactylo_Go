<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { registry } from '$app/modules';
</script>

<section class="mx-auto max-w-2xl px-8 py-12">
  <h1 class="text-3xl font-semibold">{$_('settings.title')}</h1>

  <!-- Chaque module peut fournir un panneau via settingsPanel. -->
  <div class="mt-8 flex flex-col gap-6">
    {#each registry.settingsPanels as m (m.id)}
      {@const Panel = m.settingsPanel!}
      <Panel />
    {/each}
  </div>
</section>
