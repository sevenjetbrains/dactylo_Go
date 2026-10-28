<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { registry } from '$app/modules';
  import BackupPanel from './BackupPanel.svelte';
  import ProfilesPanel from './ProfilesPanel.svelte';
</script>

<section class="mx-auto max-w-3xl px-10 py-10">
  <h1 class="text-4xl font-bold tracking-tight">{$_('settings.title')}</h1>

  <!-- Chaque module peut fournir un panneau via settingsPanel. -->
  <div class="mt-8 flex flex-col gap-6">
    <ProfilesPanel />
    {#each registry.settingsPanels as m (m.id)}
      {@const Panel = m.settingsPanel!}
      <Panel />
    {/each}
    <BackupPanel />
  </div>
</section>
