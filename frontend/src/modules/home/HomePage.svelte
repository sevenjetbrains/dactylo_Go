<script lang="ts">
  import { _, locale } from 'svelte-i18n';
  import { Overview } from '$wails/go/app/HomeService';
  import { Info } from '$wails/go/app/SystemService';
  import type { store } from '$wails/go/models';
  import Icon from '$shared/Icon.svelte';
  import { formatHours, formatNumber, localDay, parseDay } from '$shared/format';
  import { displayName, profile } from '$shared/profile.svelte';
  import { typingSettings } from '$shared/typing/settings.svelte';
  import { dictationSettings } from '$shared/dictation/settings.svelte';

  const ICON_KEYBOARD = 'M3 6h18v12H3zM7 10h.01M11 10h.01M15 10h.01M7 14h10';
  const ICON_SPEAKER = 'M11 5 6 9H2v6h4l5 4zM15.5 8.5a5 5 0 0 1 0 7M19 5a10 10 0 0 1 0 14';
  const ICON_ARROW = 'M5 12h14M13 6l6 6-6 6';

  const overview = Overview(typingSettings.exerciseSplit, typingSettings.exerciseSize);
  const info = Info();
  const lang = $derived($locale ?? 'fr');
  const name = $derived(displayName(profile.name));
  const today = new Date();

  function sessionDate(iso: string): string {
    const d = parseDay(iso);
    const time = d.toLocaleTimeString(lang, { hour: '2-digit', minute: '2-digit' });
    const yesterday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 1);
    if (localDay(d) === localDay(today)) return $_('home.recent.today', { values: { time } });
    if (localDay(d) === localDay(yesterday))
      return $_('home.recent.yesterday', { values: { time } });
    return `${d.toLocaleDateString(lang, { day: 'numeric', month: 'short' })}, ${time}`;
  }

  function sessionTitle(p: store.SessionPoint): string {
    return p.textTitle || $_(`home.modes.${p.mode}`);
  }

  function delta(v: number, digits: number, unit: string): string {
    return (v >= 0 ? '+' : '−') + formatNumber(Math.abs(v), lang, digits) + unit;
  }

  const MODE_BADGE: Record<string, string> = {
    copy: 'bg-accent-soft text-accent',
    dictation: 'bg-amber-soft text-amber-ink',
    review: 'bg-ok-soft text-ok',
  };
</script>

<section class="mx-auto max-w-6xl px-10 py-10">
  <header class="flex flex-wrap items-end justify-between gap-4">
    <div>
      <p class="text-muted text-sm font-medium first-letter:uppercase">
        {today.toLocaleDateString(lang, { weekday: 'long', day: 'numeric', month: 'long' })}
      </p>
      <h1 class="mt-1 text-4xl font-bold tracking-tight">
        {name ? $_('home.hello', { values: { name } }) : $_('home.helloAnon')}
      </h1>
    </div>
    <div class="flex gap-3">
      <a href="#/library?new=paste" class="btn btn-outline">{$_('home.paste')}</a>
      <a href="#/library?new=pdf" class="btn btn-outline">{$_('home.importPdf')}</a>
    </div>
  </header>

  {#await overview}
    <p class="text-muted mt-10">…</p>
  {:then o}
    <div class="mt-8 grid gap-5 lg:grid-cols-3">
      <!-- Révision du jour (F-43) -->
      <article class="bg-nav flex flex-col rounded-2xl p-7 text-white" data-testid="due-card">
        <div class="flex items-center justify-between">
          <p class="eyebrow" style:color="var(--color-warm)">{$_('home.due.eyebrow')}</p>
          {#if o.due.count > 0}
            <span class="text-nav-ink text-sm">≈ {o.due.minutes} min</span>
          {/if}
        </div>
        {#if o.due.count > 0}
          <h2 class="mt-4 text-3xl leading-tight font-bold">
            {$_('home.due.title', { values: { n: o.due.count } })}
          </h2>
          <ul class="mt-5 flex flex-wrap gap-2">
            {#each o.due.words as w (w)}
              <li class="bg-nav-raised rounded-md px-2.5 py-1 font-mono text-sm">{w}</li>
            {/each}
            {#if o.due.count > o.due.words.length}
              <li class="text-nav-ink px-2 py-1 font-mono text-sm">
                + {o.due.count - o.due.words.length}
              </li>
            {/if}
          </ul>
          <a
            href="#/review?session=shuffled"
            class="btn btn-warm mt-auto self-start"
            data-testid="due-banner"
          >
            {$_('home.due.cta')}
            <Icon d={ICON_ARROW} size={16} />
          </a>
        {:else}
          <h2 class="mt-4 text-2xl leading-tight font-bold">{$_('home.due.none')}</h2>
          <p class="text-nav-ink mt-2 text-sm">{$_('home.due.noneHint')}</p>
          <a href="#/review" class="text-nav-ink mt-auto pt-6 text-sm underline">
            {$_('home.due.list')}
          </a>
        {/if}
      </article>

      <!-- Reprendre la copie -->
      <article class="card flex flex-col p-7">
        <p class="eyebrow flex items-center gap-2">
          <Icon d={ICON_KEYBOARD} size={16} />
          {$_('home.continue.eyebrow')}
        </p>
        {#if o.continue}
          {@const c = o.continue}
          <h2 class="mt-4 text-xl font-bold">{c.title}</h2>
          <p class="text-muted mt-1 text-sm">
            {[
              ...c.tags,
              $_('home.continue.exercise', { values: { n: c.exercise, total: c.exercises } }),
            ].join(' · ')}
          </p>
          <p class="text-muted mt-5 font-mono text-sm leading-relaxed">« {c.excerpt} »</p>
          <div class="mt-auto pt-6">
            <div class="text-muted flex justify-between text-sm">
              <span>{$_('home.continue.progress')}</span>
              <span class="text-ink font-semibold">{Math.round(c.progress * 100)} %</span>
            </div>
            <div class="bg-line mt-2 h-1.5 overflow-hidden rounded-full">
              <div class="bg-accent h-full rounded-full" style:width="{c.progress * 100}%"></div>
            </div>
            <a href={`#/practice?text=${c.textId}`} class="btn btn-primary mt-5">
              {$_('home.continue.cta')}
            </a>
          </div>
        {:else}
          <h2 class="mt-4 text-xl font-bold">{$_('home.continue.none')}</h2>
          <p class="text-muted mt-2 text-sm">{$_('home.continue.noneHint')}</p>
          <a href="#/practice" class="btn btn-primary mt-auto self-start">
            {$_('home.continue.start')}
          </a>
        {/if}
      </article>

      <!-- Dictée : texte proposé et réglages de lecture -->
      <article class="card flex flex-col p-7" data-testid="dictation-card">
        <p class="eyebrow flex items-center gap-2">
          <Icon d={ICON_SPEAKER} size={16} />
          {$_('home.dictation.eyebrow')}
        </p>
        {#if o.dictation}
          {@const d = o.dictation}
          <h2 class="mt-4 text-xl font-bold">{d.title}</h2>
          <p class="text-muted mt-1 text-sm">
            {[...d.tags, $_('home.dictation.words', { values: { n: d.words } })].join(' · ')}
          </p>
          <ul class="mt-4 flex flex-wrap gap-2">
            <li class="badge bg-accent-soft text-accent">
              {dictationSettings.chunking === 'group'
                ? $_('home.dictation.group', { values: { n: dictationSettings.groupSize } })
                : $_(`home.dictation.chunks.${dictationSettings.chunking}`)}
            </li>
            <li class="badge bg-accent-soft text-accent">
              {$_('home.dictation.rate', {
                values: { r: formatNumber(dictationSettings.rate, lang, 1) },
              })}
            </li>
            {#if dictationSettings.speakPunctuation}
              <li class="badge bg-accent-soft text-accent">{$_('home.dictation.punctuation')}</li>
            {/if}
          </ul>
          <a
            href={`#/dictation?text=${d.textId}`}
            class="btn btn-outline text-accent border-accent mt-auto self-start"
          >
            {$_('home.dictation.cta')}
          </a>
        {:else}
          <h2 class="mt-4 text-xl font-bold">{$_('home.dictation.title')}</h2>
          <p class="text-muted mt-2 text-sm">{$_('home.dictation.desc')}</p>
          <a href="#/dictation" class="btn btn-outline mt-auto self-start"
            >{$_('home.dictation.cta')}</a
          >
        {/if}
      </article>
    </div>

    <!-- Indicateurs -->
    <dl class="mt-6 grid grid-cols-2 gap-5 lg:grid-cols-4" data-testid="kpis">
      <div class="card p-6">
        <dt class="text-muted text-sm">{$_('home.kpi.speed')}</dt>
        <dd class="mt-2 text-3xl font-bold tabular-nums">
          {formatNumber(o.speed.value, lang)}
          <span class="text-muted text-base font-medium">MPM</span>
        </dd>
        {#if o.speed.hasDelta}
          <dd class={['mt-2 text-sm font-semibold', o.speed.delta >= 0 ? 'text-ok' : 'text-error']}>
            {$_('home.kpi.speedDelta', { values: { d: delta(o.speed.delta, 0, '') } })}
          </dd>
        {/if}
      </div>
      <div class="card p-6">
        <dt class="text-muted text-sm">{$_('home.kpi.accuracy')}</dt>
        <dd class="mt-2 text-3xl font-bold tabular-nums">
          {formatNumber(o.accuracy.value, lang, 1)}
          <span class="text-muted text-base font-medium">%</span>
        </dd>
        {#if o.accuracy.hasDelta}
          <dd
            class={['mt-2 text-sm font-semibold', o.accuracy.delta >= 0 ? 'text-ok' : 'text-error']}
          >
            {delta(o.accuracy.delta, 1, ' pt')}
          </dd>
        {/if}
      </div>
      <div class="card p-6">
        <dt class="text-muted text-sm">{$_('home.kpi.streak')}</dt>
        <dd class="mt-2 text-3xl font-bold tabular-nums">
          {o.streak}
          <span class="text-muted text-base font-medium"
            >{$_('home.kpi.days', { values: { n: o.streak } })}</span
          >
        </dd>
        <dd class="text-muted mt-2 text-sm">
          {$_('home.kpi.record', { values: { n: o.bestStreak } })}
        </dd>
      </div>
      <div class="card p-6">
        <dt class="text-muted text-sm">{$_('home.kpi.week')}</dt>
        <dd class="mt-2 text-3xl font-bold tabular-nums">{formatHours(o.weekMs)}</dd>
        <dd class="text-muted mt-2 text-sm">
          {$_('home.kpi.sessions', { values: { n: o.weekSessions } })}
        </dd>
      </div>
    </dl>

    <!-- Sessions récentes -->
    <article class="card mt-6 p-7">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-bold">{$_('home.recent.title')}</h2>
        <a href="#/stats" class="text-accent text-sm font-semibold">{$_('home.recent.all')}</a>
      </div>
      {#if o.recent.length === 0}
        <p class="text-muted mt-4 text-sm">{$_('home.recent.none')}</p>
      {:else}
        <table class="mt-4 w-full text-sm">
          <thead class="text-muted text-left text-xs tracking-wide uppercase">
            <tr class="border-line border-b">
              <th class="py-2 font-semibold">{$_('home.recent.date')}</th>
              <th class="py-2 font-semibold">{$_('home.recent.text')}</th>
              <th class="py-2 font-semibold">{$_('home.recent.mode')}</th>
              <th class="py-2 text-right font-semibold">MPM</th>
              <th class="py-2 text-right font-semibold">{$_('home.recent.accuracy')}</th>
            </tr>
          </thead>
          <tbody>
            {#each o.recent as p (p.id)}
              <tr class="border-line border-b last:border-0">
                <td class="text-muted py-3">{sessionDate(p.endedAt)}</td>
                <td class="py-3">{sessionTitle(p)}</td>
                <td class="py-3">
                  <span class={['badge', MODE_BADGE[p.mode] ?? 'bg-line text-muted']}>
                    {$_(`home.modes.${p.mode}`)}
                  </span>
                </td>
                <td class="py-3 text-right font-mono tabular-nums"
                  >{formatNumber(p.wpmNet, lang)}</td
                >
                <td class="py-3 text-right font-mono tabular-nums">
                  {formatNumber(p.accuracy, lang, 1)} %
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </article>
  {:catch e}
    <p role="alert" class="text-error mt-8">{String(e)}</p>
  {/await}

  <p class="text-muted mt-10 text-xs" data-testid="system-info">
    {#await info then i}
      {$_('home.system', {
        values: { version: i.version, platform: i.platform, schema: i.schemaVersion },
      })}
    {:catch e}
      <span class="text-error">{$_('home.systemError', { values: { error: String(e) } })}</span>
    {/await}
  </p>
</section>
