<script lang="ts">
  import {
    CategoryScale,
    Chart,
    Filler,
    LinearScale,
    LineController,
    LineElement,
    PointElement,
    Tooltip,
  } from 'chart.js';

  Chart.register(
    LineController,
    LineElement,
    PointElement,
    LinearScale,
    CategoryScale,
    Tooltip,
    Filler,
  );

  // Une seule série et un seul axe par graphique (pas de double axe) :
  // le titre nomme la série, pas de légende.
  let {
    labels,
    values,
    title,
    unit = '',
    min,
    max,
    format = (v: number) => String(v),
  }: {
    labels: string[];
    values: number[];
    title: string;
    unit?: string;
    min?: number;
    max?: number;
    format?: (v: number) => string;
  } = $props();

  let canvas: HTMLCanvasElement | undefined = $state();
  // Incrémenté quand le thème change : les couleurs viennent des variables CSS.
  let themeTick = $state(0);

  $effect(() => {
    const obs = new MutationObserver(() => themeTick++);
    obs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
    return () => obs.disconnect();
  });

  $effect(() => {
    if (!canvas) return;
    void themeTick;
    const css = getComputedStyle(document.documentElement);
    const v = (name: string) => css.getPropertyValue(name).trim();
    const series = v('--viz-series-1');
    const chart = new Chart(canvas, {
      type: 'line',
      data: {
        labels,
        datasets: [
          {
            label: title,
            data: values,
            borderColor: series,
            backgroundColor: series,
            borderWidth: 2,
            pointRadius: values.length > 40 ? 0 : 3,
            pointHoverRadius: 5,
            pointHitRadius: 12,
            tension: 0.25,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        animation: false,
        interaction: { mode: 'index', intersect: false },
        scales: {
          x: {
            grid: { display: false },
            border: { color: v('--color-line') },
            ticks: { color: v('--color-muted'), maxTicksLimit: 8, maxRotation: 0 },
          },
          y: {
            min,
            max,
            grid: { color: v('--color-line') },
            border: { display: false },
            ticks: {
              color: v('--color-muted'),
              maxTicksLimit: 5,
              callback: (t) => format(Number(t)),
            },
          },
        },
        plugins: {
          tooltip: {
            backgroundColor: v('--color-panel'),
            titleColor: v('--color-muted'),
            bodyColor: v('--color-ink'),
            borderColor: v('--color-line'),
            borderWidth: 1,
            displayColors: false,
            callbacks: { label: (ctx) => `${format(ctx.parsed.y ?? 0)}${unit}` },
          },
        },
      },
    });
    return () => chart.destroy();
  });
</script>

<figure class="border-line bg-panel flex flex-col gap-2 rounded-lg border p-4">
  <figcaption class="text-sm font-semibold">{title}</figcaption>
  <div class="relative h-48">
    <canvas bind:this={canvas} aria-label={title}></canvas>
  </div>
</figure>
