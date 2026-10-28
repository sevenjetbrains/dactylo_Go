import { describe, expect, it } from 'vitest';
import { Clock } from './clock';

describe('Clock', () => {
  it('exclut les pauses du temps écoulé', () => {
    let t = 0;
    const c = new Clock(() => t);
    expect(c.elapsed()).toBe(0);
    c.resume(); // sans effet avant start
    expect(c.running).toBe(false);

    c.start();
    t = 1000;
    c.pause();
    t = 5000; // en pause
    expect(c.elapsed()).toBe(1000);
    c.resume();
    t = 5500;
    expect(c.elapsed()).toBe(1500);
    c.start(); // sans effet une fois démarré
    expect(c.elapsed()).toBe(1500);
  });

  it('reprend une session interrompue en pause', () => {
    let t = 100;
    const c = new Clock(() => t);
    c.restore(4000);
    expect(c.started).toBe(true);
    expect(c.running).toBe(false);
    t = 900;
    expect(c.elapsed()).toBe(4000);
    c.resume();
    t = 1400;
    expect(c.elapsed()).toBe(4500);
  });
});
