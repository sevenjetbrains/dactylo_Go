import { afterEach, describe, expect, it } from 'vitest';
import { router } from './router.svelte';

/** Change l'adresse comme un lien du menu, et attend l'événement. */
async function go(to: string) {
  window.location.hash = to;
  await new Promise((r) => setTimeout(r, 0));
}

describe('router.guard', () => {
  let unguard = () => {};
  afterEach(async () => {
    unguard();
    router.stay();
    await go('/');
  });

  it('navigue librement sans garde', async () => {
    await go('/stats');
    expect(router.path).toBe('/stats');
    expect(router.pending).toBeNull();
  });

  it('retient la navigation tant que le garde renvoie un message', async () => {
    await go('/practice');
    unguard = router.guard(() => 'session en cours');
    await go('/dictation');
    expect(router.path).toBe('/practice');
    expect(window.location.hash).toBe('#/practice');
    expect(router.pending).toEqual({ to: '#/dictation', message: 'session en cours' });

    router.stay();
    expect(router.pending).toBeNull();
    expect(router.path).toBe('/practice');
  });

  it('navigue après confirmation', async () => {
    await go('/practice');
    unguard = router.guard(() => 'session en cours');
    await go('/stats');
    router.leave();
    expect(router.pending).toBeNull();
    expect(router.path).toBe('/stats');
    expect(window.location.hash).toBe('#/stats');
  });

  it('laisse passer quand le garde ne retient rien, ou une fois retiré', async () => {
    unguard = router.guard(() => null);
    await go('/stats');
    expect(router.path).toBe('/stats');

    unguard();
    unguard = router.guard(() => 'session en cours');
    unguard();
    await go('/library');
    expect(router.path).toBe('/library');
  });
});
