<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { layoutRows, strokeFor, type Finger, type KeyDef } from './layouts';

  // Clavier virtuel (F-08) : touche suivante et doigt conseillé, dans la
  // disposition du profil (AZERTY, QWERTY ou BÉPO).
  let { next, layout = 'azerty' }: { next: string | null; layout?: string } = $props();

  const stroke = $derived(next === null ? null : strokeFor(next, layout));
  const L = $derived(layoutRows(layout));
  // Maj : celle de la main opposée à la touche.
  const shiftSide = $derived(stroke?.key.finger.startsWith('l') ? 'right' : 'left');

  /** Libellé de touche : lettres en capitales, comme sur un clavier réel. */
  const capLabel = (l: string) => (/^[a-z]$/.test(l) ? l.toUpperCase() : l);

  type Cap = { id: string; label: string; width: number; finger: Finger; key?: KeyDef };
  const mod = (id: string, label: string, width: number): Cap => ({
    id,
    label,
    width,
    finger: 'thumb',
  });
  const caps = (row: KeyDef[]): Cap[] =>
    row.map((k) => ({
      id: k.label,
      label: capLabel(k.label),
      width: 1,
      finger: k.finger,
      key: k,
    }));

  const rows: Cap[][] = $derived([
    [...caps(L[0]!), mod('backspace', '⌫', 1.6)],
    [mod('tab', '⇥', 1.4), ...caps(L[1]!), mod('enter', '↵', 1.2)],
    [mod('caps', '⇪', 1.7), ...caps(L[2]!), mod('enter2', '', 0.9)],
    [mod('lshift', '⇧', 1.2), ...caps(L[3]!), mod('rshift', '⇧', 2.4)],
    [
      mod('lctrl', 'Ctrl', 1.4),
      mod('alt', 'Alt', 1.2),
      {
        id: 'space',
        label: $_('typing.keyboard.space'),
        width: 6.6,
        finger: 'thumb',
        key: L[4]![0],
      },
      mod('altgr', 'AltGr', 1.4),
      mod('rctrl', 'Ctrl', 1.4),
    ],
  ]);

  /** Touche à mettre en évidence : touche, touche morte, Maj, AltGr, Entrée. */
  function active(c: Cap): boolean {
    const s = stroke;
    if (!s) return false;
    if (c.key && (c.key === s.key || c.key === s.dead)) return true;
    if (c.id === 'enter' && s.key.base === '↵') return true;
    if (s.shift && c.id === (shiftSide === 'left' ? 'lshift' : 'rshift')) return true;
    if (s.altgr && c.id === 'altgr') return true;
    return false;
  }
</script>

<figure class="flex flex-col items-center gap-3" aria-label={$_('typing.keyboard.label')}>
  <div class="bg-sunken border-line w-full max-w-[52rem] rounded-2xl border p-3">
    {#each rows as row, r (r)}
      <div class="flex gap-1.5 [&+&]:mt-1.5">
        {#each row as c (c.id)}
          {@const on = active(c)}
          <div
            class={[
              'flex h-11 min-w-0 items-center justify-center rounded-lg border font-mono text-sm',
              on ? 'bg-accent border-accent text-white shadow-md' : 'bg-panel border-line',
            ]}
            style:flex="{c.width} 1 0"
            style:border-top={c.finger === 'thumb' || on
              ? undefined
              : `3px solid var(--finger-${c.finger})`}
            data-key={c.id}
            data-active={on || undefined}
          >
            {c.label}
          </div>
        {/each}
      </div>
    {/each}
  </div>
  {#if stroke}
    <figcaption class="text-muted flex items-center gap-2 text-sm" data-testid="next-key">
      <span
        class="inline-block h-3 w-3 rounded-sm"
        style:background="var(--finger-{stroke.key.finger})"
        aria-hidden="true"
      ></span>
      {$_('typing.keyboard.next')}
      <strong class="text-ink font-mono">
        {[
          stroke.dead ? `${stroke.deadShift ? '⇧ ' : ''}${stroke.dead.label}  →  ` : '',
          stroke.shift ? '⇧ ' : '',
          stroke.altgr ? 'AltGr ' : '',
          stroke.key.base === ' ' ? $_('typing.keyboard.space') : capLabel(stroke.key.label),
        ].join('')}
      </strong>
      — {$_(`typing.keyboard.fingers.${stroke.key.finger}`)}
    </figcaption>
  {/if}
</figure>
