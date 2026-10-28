<!--
  Icône SVG au trait à partir d'un tracé (attribut d).
-->
<script lang="ts">
  let {
    d,
    size = 20,
    class: className = '',
  }: { d: string; size?: number; class?: string } = $props();
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="2"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
  class={className}
>
  <path {d} />
</svg>
