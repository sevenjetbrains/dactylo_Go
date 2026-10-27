import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { Summary } from '$wails/go/app/ReviewService';
import HomePage from './HomePage.svelte';

vi.mock('$wails/go/app/SystemService', () => ({
  Info: vi.fn().mockResolvedValue({
    version: '9.9.9',
    platform: 'windows/amd64',
    databasePath: 'C:/x.db',
    schemaVersion: 1,
  }),
}));

vi.mock('$wails/go/app/ReviewService', () => ({
  Summary: vi.fn(),
}));

describe('HomePage', () => {
  it("affiche l'accueil et les informations système", async () => {
    vi.mocked(Summary).mockResolvedValue({ due: 0, learning: 0, mastered: 0 });
    render(HomePage);
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Bienvenue');
    expect(screen.getAllByRole('listitem')).toHaveLength(3);
    expect(await screen.findByText(/Version 9\.9\.9/)).toHaveTextContent('schéma de données v1');
    expect(screen.queryByTestId('due-banner')).toBeNull();
  });

  it('propose les mots à réviser aujourd’hui (F-43)', async () => {
    vi.mocked(Summary).mockResolvedValue({ due: 12, learning: 15, mastered: 3 });
    render(HomePage);
    const banner = await screen.findByTestId('due-banner');
    expect(banner).toHaveTextContent('12 mots à réviser aujourd’hui'.replace('’', "'"));
    expect(banner).toHaveAttribute('href', '#/review?session=shuffled');
  });
});
