<script lang="ts">
  import { CharState, type TypingEngine } from './engine';
  import { firstVisibleLine, lineAt, wrapLines } from './layout';
  import type { LinePosition } from './settings.svelte';

  let {
    engine,
    tick,
    fontSize,
    visibleLines,
    linePosition,
  }: {
    engine: TypingEngine;
    /** Incrémenté à chaque frappe : déclenche le rafraîchissement. */
    tick: number;
    fontSize: number;
    visibleLines: number;
    linePosition: LinePosition;
  } = $props();

  const LINE_HEIGHT = 1.8; // em

  let container: HTMLDivElement | undefined = $state();
  let probe: HTMLSpanElement | undefined = $state();
  let columns = $state(60);

  // Largeur d'un caractère monospace → nombre de colonnes.
  $effect(() => {
    if (!container || !probe) return;
    void fontSize;
    const measure = () => {
      const charWidth = probe!.getBoundingClientRect().width / 50 || fontSize * 0.6;
      columns = Math.max(10, Math.floor(container!.clientWidth / charWidth) - 1);
    };
    measure();
    const ro = new ResizeObserver(measure);
    ro.observe(container);
    return () => ro.disconnect();
  });

  // Recalculé seulement si la largeur change, pas à chaque frappe.
  const lines = $derived(wrapLines(engine.chars, columns));

  const view = $derived.by(() => {
    void tick;
    const current = lineAt(lines, engine.position);
    const first = firstVisibleLine(current, visibleLines, lines.length, linePosition);
    return lines.slice(first, first + visibleLines).map((l) => {
      const cells: { ch: string; cls: string; key: number }[] = [];
      for (let i = l.start; i < l.end; i++) {
        cells.push({ ch: display(engine.chars[i]!), cls: classFor(i), key: i });
      }
      return { key: l.start, cells };
    });
  });

  function display(ch: string): string {
    if (ch === '\n') return '↵';
    // insécable : une espace seule dans un <span> peut être repliée par le moteur
    if (ch === ' ' || ch === ' ' || ch === ' ') return ' ';
    return ch;
  }

  function classFor(i: number): string {
    const s = engine.state[i];
    const cursor = i === engine.position ? ' ch-cursor' : '';
    // préfixe ch- : « fixed », « block »... sont aussi des utilitaires Tailwind
    const nl = engine.chars[i] === '\n' ? ' ch-nl' : '';
    if (s === CharState.Correct) return 'ch-ok' + nl + cursor;
    if (s === CharState.Corrected) return 'ch-fixed' + nl + cursor;
    if (s === CharState.Error) return 'ch-err' + nl + cursor;
    return 'ch-todo' + nl + cursor;
  }
</script>

<div
  bind:this={container}
  class="typing relative w-full overflow-hidden font-mono"
  style:font-size="{fontSize}px"
  style:line-height={LINE_HEIGHT}
  style:height="{visibleLines * LINE_HEIGHT}em"
>
  <span bind:this={probe} aria-hidden="true" class="invisible absolute whitespace-pre"
    >{'M'.repeat(50)}</span
  >
  {#each view as line (line.key)}
    <div class="whitespace-pre">
      {#each line.cells as c (c.key)}<span class={c.cls}>{c.ch}</span>{/each}
    </div>
  {/each}
</div>

<style>
  .ch-todo {
    color: var(--color-muted);
  }
  .ch-ok {
    color: var(--color-ink);
  }
  .ch-fixed {
    color: var(--color-ink);
    text-decoration: underline dotted var(--color-error);
    text-underline-offset: 0.25em;
  }
  .ch-err {
    color: var(--color-error);
    background: color-mix(in srgb, var(--color-error) 15%, transparent);
    border-radius: 2px;
  }
  .ch-nl {
    opacity: 0.45;
  }
  .ch-cursor {
    box-shadow: inset 0 -0.14em 0 var(--color-accent);
    color: var(--color-ink);
  }
</style>
