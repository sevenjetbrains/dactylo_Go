import { describe, expect, it } from 'vitest';
import type { Component } from 'svelte';
import { createRegistry, normalizePath } from './registry';
import type { ModuleDefinition } from './types';

const Dummy = (() => {}) as unknown as Component;

function def(over: Partial<ModuleDefinition>): ModuleDefinition {
  return { id: 'x', route: '/x', labelKey: 'x', icon: '', order: 0, page: Dummy, ...over };
}

describe('createRegistry', () => {
  it('trie par ordre et résout les routes', () => {
    const r = createRegistry([
      def({ id: 'b', route: '/b', order: 2 }),
      def({ id: 'a', route: '/', order: 1 }),
    ]);
    expect(r.modules.map((m) => m.id)).toEqual(['a', 'b']);
    expect(r.resolve('#/b')?.id).toBe('b');
    expect(r.resolve('')?.id).toBe('a');
    expect(r.resolve('/inconnue')).toBeUndefined();
  });

  it('exclut les modules masqués du menu', () => {
    const r = createRegistry([
      def({ id: 'a', route: '/a' }),
      def({ id: 'h', route: '/h', hidden: true }),
    ]);
    expect(r.menu.map((m) => m.id)).toEqual(['a']);
    expect(r.resolve('/h')?.id).toBe('h');
  });

  it('liste les panneaux de réglages', () => {
    const r = createRegistry([
      def({ id: 'a', route: '/a' }),
      def({ id: 'b', route: '/b', settingsPanel: Dummy }),
    ]);
    expect(r.settingsPanels.map((m) => m.id)).toEqual(['b']);
  });

  it('refuse les doublons et les déclarations invalides', () => {
    expect(() =>
      createRegistry([def({ id: 'a', route: '/a' }), def({ id: 'a', route: '/b' })]),
    ).toThrow(/double/);
    expect(() =>
      createRegistry([def({ id: 'a', route: '/a' }), def({ id: 'b', route: '/a' })]),
    ).toThrow(/double/);
    expect(() => createRegistry([def({ id: 'Bad_Id' })])).toThrow(/invalide/);
    expect(() => createRegistry([def({ route: 'x' })])).toThrow(/invalide/);
  });
});

describe('normalizePath', () => {
  it.each([
    ['', '/'],
    ['#', '/'],
    ['#/', '/'],
    ['#/settings', '/settings'],
    ['#/settings/', '/settings'],
    ['settings?x=1', '/settings'],
    ['//a//', '/a'],
  ])('%s -> %s', (input, expected) => {
    expect(normalizePath(input)).toBe(expected);
  });
});
