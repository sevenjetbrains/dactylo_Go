import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import HomePage from './HomePage.svelte';

vi.mock('$wails/go/app/SystemService', () => ({
  Info: vi.fn().mockResolvedValue({
    version: '9.9.9',
    platform: 'windows/amd64',
    databasePath: 'C:/x.db',
    schemaVersion: 1,
  }),
}));

describe('HomePage', () => {
  it("affiche l'accueil et les informations système", async () => {
    render(HomePage);
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Bienvenue');
    expect(screen.getAllByRole('listitem')).toHaveLength(3);
    expect(await screen.findByText(/Version 9\.9\.9/)).toHaveTextContent('schéma de données v1');
  });
});
