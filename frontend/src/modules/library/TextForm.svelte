<!--
  Formulaire d'un texte (coller ou modifier) : titre, contenu, étiquettes, découpage.
-->
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
  let split = $state(start.split ?? '');
  let splitSize = $state(start.splitSize || 1);
  let saving = $state(false);
  let error = $state('');

  /** Valide le formulaire et transmet la saisie. */
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
        split,
        splitSize: split && split !== 'none' ? splitSize : 0,
      });
    } catch (err) {
      error = String(err);
    } finally {
      saving = false;
    }
  }
</script>

<form class="flex flex-col gap-4" onsubmit={submit}>
  <header class="flex items-center justify-between">
    <h2 class="text-lg font-bold">{heading}</h2>
    {#if onCancel}
      <button
        type="button"
        class="text-muted hover:text-ink text-xl leading-none"
        onclick={onCancel}
        aria-label={$_('library.close')}>×</button
      >
    {/if}
  </header>
  <label class="flex flex-col gap-1.5 text-sm">
    <span class="font-semibold">{$_('library.form.title')}</span>
    <input class="field" bind:value={title} placeholder={$_('library.form.titlePlaceholder')} />
  </label>
  <label class="flex flex-col gap-1.5 text-sm">
    <span class="font-semibold">{$_('library.form.content')}</span>
    <textarea
      class="field bg-sunken min-h-56 font-mono text-sm leading-relaxed"
      bind:value={content}
      required></textarea>
  </label>
  <div class="grid grid-cols-2 gap-3">
    <label class="flex flex-col gap-1.5 text-sm">
      <span class="font-semibold">{$_('library.split.label')}</span>
      <select class="field" bind:value={split}>
        <option value="">{$_('library.split.default')}</option>
        <option value="none">{$_('library.split.none')}</option>
        <option value="paragraph">{$_('library.split.paragraph')}</option>
        <option value="lines">{$_('library.split.lines')}</option>
      </select>
    </label>
    {#if split === 'paragraph' || split === 'lines'}
      <label class="flex flex-col gap-1.5 text-sm">
        <span class="font-semibold">{$_(`library.split.size.${split}`)}</span>
        <input class="field" type="number" min="1" max="100" bind:value={splitSize} />
      </label>
    {/if}
    <label class="flex flex-col gap-1.5 text-sm">
      <span class="font-semibold">{$_('library.form.tags')}</span>
      <input class="field" bind:value={tags} placeholder={$_('library.form.tagsPlaceholder')} />
    </label>
    <label class="flex flex-col gap-1.5 text-sm">
      <span class="font-semibold">{$_('library.form.language')}</span>
      <select class="field" bind:value={language}>
        <option value="fr">Français</option>
        <option value="en">English</option>
      </select>
    </label>
  </div>
  {#if error}<p role="alert" class="text-error text-sm">{error}</p>{/if}
  <button type="submit" class="btn btn-primary w-full py-3" disabled={saving}>{submitLabel}</button>
</form>
