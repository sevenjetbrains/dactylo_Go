import { afterEach, describe, expect, it, vi } from 'vitest';
import { flushSync } from 'svelte';
import { Create, Get, SaveSettings, SaveSettingsFor } from '$wails/go/app/ProfileService';
import { createProfile, loadProfile, profile } from './profile.svelte';
import { typingSettings } from './typing/settings.svelte';
import { dictationSettings } from './dictation/settings.svelte';

vi.mock('$wails/go/app/ProfileService', () => ({
  Get: vi.fn(),
  List: vi.fn(),
  Create: vi.fn(),
  Update: vi.fn(),
  Delete: vi.fn(),
  Rename: vi.fn(),
  Switch: vi.fn(),
  SaveSettings: vi.fn().mockResolvedValue(undefined),
  SaveSettingsFor: vi.fn().mockResolvedValue(undefined),
}));

afterEach(() => vi.useRealTimers());

describe('réglages du profil (F-51)', () => {
  it('charge les réglages du profil puis enregistre les modifications', async () => {
    vi.useFakeTimers();
    vi.mocked(Get).mockResolvedValue({
      id: 3,
      name: 'Léa',
      keyboardLayout: 'bepo',
      settings: JSON.stringify({
        typing: { fontSize: 30, sounds: true },
        dictation: { rate: 0.8 },
      }),
      createdAt: '',
    } as never);
    await loadProfile();
    expect(profile).toMatchObject({ id: 3, name: 'Léa', keyboardLayout: 'bepo', loaded: true });
    expect(typingSettings.fontSize).toBe(30);
    expect(typingSettings.sounds).toBe(true);
    expect(dictationSettings.rate).toBe(0.8);

    flushSync();
    expect(SaveSettings).not.toHaveBeenCalled(); // le chargement n'est pas une modification

    typingSettings.fontSize = 18;
    flushSync();
    await vi.advanceTimersByTimeAsync(1000);
    expect(SaveSettings).toHaveBeenCalledTimes(1);
    const saved = JSON.parse(vi.mocked(SaveSettings).mock.calls[0]![0]) as {
      typing: { fontSize: number };
      locale: string;
    };
    expect(saved.typing.fontSize).toBe(18);
    expect(saved.locale).toBe('fr');
  });

  it('un nouveau profil part des réglages par défaut', async () => {
    vi.mocked(Create).mockResolvedValue({ id: 9, name: 'Tom', keyboardLayout: 'qwerty' } as never);
    const p = await createProfile('Tom', 'qwerty');
    expect(p.id).toBe(9);
    const [id, raw] = vi.mocked(SaveSettingsFor).mock.calls[0]!;
    expect(id).toBe(9);
    expect(JSON.parse(raw)).toMatchObject({
      typing: { fontSize: 24, sounds: false },
      theme: 'system',
    });
  });
});
