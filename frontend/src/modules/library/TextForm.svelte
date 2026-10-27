<script lang="ts">
  import { untrack } from 'svelte';
  import { _ } from 'svelte-i18n';
  import type { app } from '$wails/go/models';

  let {
    heading,
    initial,
    submitLabel,
    onSubmit,
    onCancel,
  }: {
    heading: string;
    initial?: Partial<app.TextInput>;
    submitLabel: string;
    onSubmit: (input: app.TextInput) => Promise<void>;
    onCancel?: () => void;
  } = $props();

  // Copie de départ : le formulaire est ensuite indépendant de `initial`.
  const start = untrack(() => $state.snapshot(initial)) ?? {};
  let title = $state(start.title ?? '');
  let content = $state(start.content ?? '');
  let tags = $state((start.tags ?? []).join(', '));
  let language = $state(start.language || 'fr');
  let saving = $state(false);
  let error = $state('');

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    if (!content.trim()) {
      error = $_('library.form.empty');
      return;
    }
    saving = true;
    error = '';
    try {
      await onSubmit({
        title,
        content,
        language,
        tags: tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
        source: start.source ?? 'paste',
        sourceFile: start.sourceFile ?? '',
        pages: start.pages ?? '',
      });
    } catch (err) {
      error = String(err);
    } finally {
      saving = false;
    }
  }
</script>

<form class="border-line bg-panel flex flex-col gap-4 rounded-lg border p-5" onsubmit={submit}>
  <h2 class="font-semibold">{heading}</h2>
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
      class="border-line bg-surface min-h-48 rounded-md border px-3 py-2 font-mono"
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
  {#if error}<p role="alert" class="text-error text-sm">{error}</p>{/if}
  <div class="flex gap-3">
    <button
      type="submit"
      class="bg-accent text-accent-ink rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
      disabled={saving}
    >
      {submitLabel}
    </button>
    {#if onCancel}
      <button
        type="button"
        class="border-line rounded-md border px-4 py-2 text-sm"
        onclick={onCancel}
      >
        {$_('library.cancel')}
      </button>
    {/if}
  </div>
</form>
