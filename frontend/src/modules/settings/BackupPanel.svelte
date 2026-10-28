<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { Export, Restore } from '$wails/go/app/BackupService';
  import { persistSettings } from '$shared/profile.svelte';

  // Sauvegarde et restauration de toutes les données dans un seul fichier
  // (F-52). La restauration remplace la base : l'interface est rechargée.
  let message = $state('');
  let error = $state('');
  let confirmRestore = $state(false);
  let busy = $state(false);

  /** Enregistre une sauvegarde dans le fichier choisi. */
  async function exportBackup() {
    message = error = '';
    busy = true;
    try {
      await persistSettings();
      const r = await Export($_('settings.backup.exportTitle'));
      if (r.path) message = $_('settings.backup.exported', { values: { path: r.path } });
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
    }
  }

  /** Restaure une sauvegarde (deux clics : confirmer), puis recharge. */
  async function restoreBackup() {
    if (!confirmRestore) {
      confirmRestore = true;
      return;
    }
    confirmRestore = false;
    message = error = '';
    busy = true;
    try {
      const r = await Restore($_('settings.backup.restoreTitle'));
      if (r.path) window.location.reload();
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
    }
  }
</script>

<fieldset class="card p-5" data-testid="backup">
  <legend class="px-1 font-semibold">{$_('settings.backup.title')}</legend>
  <p class="text-muted mt-1 text-sm">{$_('settings.backup.hint')}</p>

  <div class="mt-4 flex flex-wrap items-center gap-2">
    <button class="btn btn-outline" disabled={busy} onclick={exportBackup}>
      {$_('settings.backup.export')}
    </button>
    <button
      class={['btn', confirmRestore ? 'btn-warm' : 'btn-outline']}
      disabled={busy}
      onclick={restoreBackup}
    >
      {confirmRestore ? $_('settings.backup.confirmRestore') : $_('settings.backup.restore')}
    </button>
    {#if confirmRestore}
      <button class="text-muted text-sm underline" onclick={() => (confirmRestore = false)}>
        {$_('settings.backup.cancel')}
      </button>
    {/if}
  </div>
  {#if confirmRestore}
    <p class="text-muted mt-2 text-sm">{$_('settings.backup.restoreWarning')}</p>
  {/if}
  {#if message}<p class="text-ok mt-3 text-sm break-all" role="status">{message}</p>{/if}
  {#if error}<p role="alert" class="text-error mt-3 text-sm">{error}</p>{/if}
</fieldset>
