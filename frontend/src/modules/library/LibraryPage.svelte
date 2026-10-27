<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { CreateText, DeleteText, ListTexts } from '$wails/go/app/TextService';
  import type { store } from '$wails/go/models';

  let texts: store.TextSummary[] = $state([]);
  let loadError = $state('');

  let title = $state('');
  let content = $state('');
  let tags = $state('');
  let language = $state('fr');
  let saving = $state(false);
  let formError = $state('');
  let confirmDelete: number | null = $state(null);

  async function refresh() {
    try {
      texts = await ListTexts();
      loadError = '';
    } catch (e) {
      loadError = String(e);
    }
  }
  void refresh();

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    if (!content.trim()) {
      formError = $_('library.form.empty');
      return;
    }
    saving = true;
    formError = '';
    try {
      await CreateText({
        title,
        content,
        language,
        tags: tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
      });
      title = content = tags = '';
      await refresh();
    } catch (err) {
      formError = String(err);
    } finally {
      saving = false;
    }
  }

  async function remove(id: number) {
    if (confirmDelete !== id) {
      confirmDelete = id;
      return;
    }
    confirmDelete = null;
    await DeleteText(id);
    await refresh();
  }
</script>

<section class="mx-auto max-w-4xl px-8 py-10">
  <h1 class="text-3xl font-semibold">{$_('library.title')}</h1>

  <form
    class="border-line bg-panel mt-8 flex flex-col gap-4 rounded-lg border p-5"
    onsubmit={submit}
  >
    <h2 class="font-semibold">{$_('library.form.heading')}</h2>
    <label class="flex flex-col gap-1 text-sm">
      <span class="text-muted">{$_('library.form.title')}</span>
      <input
        class="border-line bg-surface rounded-md border px-3 py-1.5"
        bind:value={title}
        placeholder={$_('library.form.titlePlaceholder')}
      />
    </label>
    <label class="flex flex-col gap-1 text-sm">
      <span class="text-muted">{$_('library.form.content')}</span>
      <textarea
        class="border-line bg-surface min-h-40 rounded-md border px-3 py-2 font-mono"
        bind:value={content}
        required></textarea>
    </label>
    <div class="flex flex-wrap gap-4">
      <label class="flex flex-1 flex-col gap-1 text-sm">
        <span class="text-muted">{$_('library.form.tags')}</span>
        <input
          class="border-line bg-surface rounded-md border px-3 py-1.5"
          bind:value={tags}
          placeholder={$_('library.form.tagsPlaceholder')}
        />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span class="text-muted">{$_('library.form.language')}</span>
        <select class="border-line bg-surface rounded-md border px-3 py-1.5" bind:value={language}>
          <option value="fr">Français</option>
          <option value="en">English</option>
        </select>
      </label>
    </div>
    {#if formError}<p role="alert" class="text-error text-sm">{formError}</p>{/if}
    <div>
      <button
        type="submit"
        class="bg-accent text-accent-ink rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
        disabled={saving}
      >
        {$_('library.form.save')}
      </button>
    </div>
  </form>

  <h2 class="mt-10 mb-3 font-semibold">{$_('library.list', { values: { n: texts.length } })}</h2>
  {#if loadError}
    <p role="alert" class="text-error">{loadError}</p>
  {:else if texts.length === 0}
    <p class="text-muted">{$_('library.empty')}</p>
  {:else}
    <ul class="flex flex-col gap-2">
      {#each texts as t (t.id)}
        <li class="border-line bg-panel flex items-center gap-4 rounded-lg border px-4 py-3">
          <div class="min-w-0 flex-1">
            <p class="font-medium">
              {t.title}
              {#each t.tags as tag (tag)}
                <span class="bg-line text-muted ms-1 rounded-full px-2 py-0.5 text-xs">{tag}</span>
              {/each}
            </p>
            <p class="text-muted truncate text-sm">{t.preview}</p>
          </div>
          <a
            href={`#/practice?text=${t.id}`}
            class="bg-accent text-accent-ink shrink-0 rounded-md px-3 py-1.5 text-sm font-medium"
          >
            {$_('library.practice')}
          </a>
          <button
            class={[
              'shrink-0 rounded-md border px-3 py-1.5 text-sm',
              confirmDelete === t.id ? 'border-error text-error' : 'border-line text-muted',
            ]}
            onclick={() => remove(t.id)}
            onblur={() => (confirmDelete = null)}
          >
            {confirmDelete === t.id ? $_('library.confirmDelete') : $_('library.delete')}
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</section>
