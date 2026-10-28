import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { fireEvent, waitFor } from '@testing-library/svelte';
import { Overview } from '$wails/go/app/HomeService';
import { Abandon } from '$wails/go/app/PracticeService';
import HomePage from './HomePage.svelte';

vi.mock('$wails/go/app/SystemService', () => ({
  Info: vi.fn().mockResolvedValue({
    version: '9.9.9',
    platform: 'windows/amd64',
    databasePath: 'C:/x.db',
    schemaVersion: 1,
  }),
}));
vi.mock('$wails/go/app/HomeService', () => ({ Overview: vi.fn() }));
vi.mock('$wails/go/app/PracticeService', () => ({ Abandon: vi.fn().mockResolvedValue(undefined) }));
vi.mock('$wails/go/app/ProfileService', () => ({ Get: vi.fn(), Rename: vi.fn() }));

const empty = {
  profile: { id: 1, name: 'Moi', keyboardLayout: 'azerty' },
  due: { count: 0, words: [], minutes: 0 },
  continue: null,
  speed: { value: 0, delta: 0, hasDelta: false },
  accuracy: { value: 0, delta: 0, hasDelta: false },
  streak: 0,
  bestStreak: 0,
  weekMs: 0,
  weekSessions: 0,
  recent: [],
  interrupted: null,
};

describe('HomePage', () => {
  it('accueil vide', async () => {
    vi.mocked(Overview).mockResolvedValue(empty as never);
    render(HomePage);
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Bonjour.');
    expect(await screen.findByText("Aucun mot à réviser aujourd'hui")).toBeInTheDocument();
    expect(screen.getByText('Aucune copie en cours')).toBeInTheDocument();
    expect(await screen.findByText(/Version 9\.9\.9/)).toHaveTextContent('schéma de données v1');
    expect(screen.queryByTestId('due-banner')).toBeNull();
  });

  it('révision du jour, reprise et indicateurs (F-43)', async () => {
    vi.mocked(Overview).mockResolvedValue({
      ...empty,
      due: {
        count: 12,
        words: ['accueillir', 'parallèle', 'néanmoins', 'œuvre', 'vieillard'],
        minutes: 5,
      },
      continue: {
        textId: 7,
        title: 'Le Tour du monde en 80 jours',
        tags: ['Jules Verne'],
        exercise: 3,
        exercises: 8,
        excerpt: 'était habitée par Phileas Fogg',
        progress: 0.38,
      },
      speed: { value: 48, delta: 3, hasDelta: true },
      accuracy: { value: 96.4, delta: 0.8, hasDelta: true },
      streak: 9,
      bestStreak: 14,
      weekMs: 135 * 60000,
      weekSessions: 11,
    } as never);
    render(HomePage);
    const banner = await screen.findByTestId('due-banner');
    expect(banner).toHaveAttribute('href', '#/review?session=shuffled');
    expect(screen.getByTestId('due-card')).toHaveTextContent("12 mots à réviser aujourd'hui");
    expect(screen.getByTestId('due-card')).toHaveTextContent('+ 7');
    expect(screen.getByText('Jules Verne · exercice 3 sur 8')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Continuer' })).toHaveAttribute(
      'href',
      '#/practice?text=7',
    );
    const kpis = screen.getByTestId('kpis');
    expect(kpis).toHaveTextContent('+3 sur 10 sessions');
    expect(kpis).toHaveTextContent('96,4');
    expect(kpis).toHaveTextContent('Record : 14 jours');
    expect(kpis).toHaveTextContent('2 h 15');
    expect(kpis).toHaveTextContent('11 sessions');
  });

  it('propose de reprendre une session interrompue (NF-06)', async () => {
    vi.mocked(Overview).mockResolvedValue({
      ...empty,
      interrupted: {
        id: 42,
        textId: 7,
        textTitle: 'Le Horla',
        startedAt: new Date().toISOString(),
        keystrokes: 120,
      },
    } as never);
    render(HomePage);
    const box = await screen.findByTestId('interrupted');
    expect(box).toHaveTextContent('« Le Horla »');
    expect(box).toHaveTextContent('120 frappes enregistrées');
    expect(screen.getByRole('link', { name: 'Reprendre' })).toHaveAttribute(
      'href',
      '#/practice?resume=42',
    );
    await fireEvent.click(screen.getByRole('button', { name: 'Ignorer' }));
    expect(Abandon).toHaveBeenCalledWith(42);
    await waitFor(() => expect(screen.queryByTestId('interrupted')).toBeNull());
  });
});
