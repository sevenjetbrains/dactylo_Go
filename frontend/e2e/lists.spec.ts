// Essais de bout en bout des listes longues : recherche et pagination
// (backend simulé).

import { expect, test } from '@playwright/test';
import { installMockBackend } from './mock-backend';

test('révisions : recherche sans accents et pages de 25 mots', async ({ page }) => {
  const words = Array.from({ length: 60 }, (_, i) => `mot${i + 1}`);
  words[41] = 'Élève';
  await page.addInitScript(installMockBackend, { texts: [], words });
  await page.goto('/#/review');

  const rows = page.locator('tbody tr');
  await expect(rows).toHaveCount(25);
  await expect(page.getByText('Page 1 sur 3 · 60 résultats')).toBeVisible();
  await page.screenshot({ path: 'test-results/screens/review-pages.png' });

  await page.getByRole('button', { name: 'Suivant' }).click();
  await expect(page.getByText('Page 2 sur 3 · 60 résultats')).toBeVisible();
  await expect(rows.first()).toContainText('mot26');
  await page.getByRole('button', { name: 'Précédent' }).click();
  await expect(rows.first()).toContainText('mot1');

  // numéros de page : aller directement à la dernière
  const pager = page.getByRole('navigation', { name: 'Pages de la liste' });
  await pager.getByRole('button', { name: 'Aller à la page 3' }).click();
  await expect(rows).toHaveCount(10);
  await expect(pager.getByRole('button', { name: 'Aller à la page 3' })).toHaveAttribute(
    'aria-current',
    'page',
  );
  await page.screenshot({ path: 'test-results/screens/review-page3.png' });
  await pager.getByRole('button', { name: 'Aller à la page 1' }).click();

  // recherche : sans accent ni majuscule, retour à la première page
  await page.getByRole('button', { name: 'Suivant' }).click();
  await page.getByPlaceholder('Rechercher un mot').fill('ELEVE');
  await expect(rows).toHaveCount(1);
  await expect(rows.first()).toContainText('Élève');
  await expect(page.getByRole('navigation', { name: 'Pages de la liste' })).toBeHidden();

  await page.getByPlaceholder('Rechercher un mot').fill('zzz');
  await expect(page.getByText('Aucun résultat pour « zzz ».')).toBeVisible();
});

test('copie : recherche et pages dans le choix du texte', async ({ page }) => {
  const texts = Array.from({ length: 45 }, (_, i) => ({
    title: `Texte ${i + 1}`,
    content: i === 7 ? 'La fenêtre est ouverte.' : `Contenu numéro ${i + 1}.`,
  }));
  await page.addInitScript(installMockBackend, { texts });
  await page.goto('/#/practice');

  const items = page.locator('main li');
  await expect(items).toHaveCount(20);
  await expect(page.getByText('Page 1 sur 3 · 45 résultats')).toBeVisible();
  await page.getByPlaceholder('Rechercher un texte').fill('fenetre');
  await expect(items).toHaveCount(1);
  await expect(items.first()).toContainText('Texte 8');
});

test('bibliothèque : pages de 20 textes', async ({ page }) => {
  const texts = Array.from({ length: 25 }, (_, i) => ({
    title: `Texte ${i + 1}`,
    content: `Contenu ${i + 1}.`,
  }));
  await page.addInitScript(installMockBackend, { texts });
  await page.goto('/#/library');
  await expect(page.locator('main ul > li')).toHaveCount(20);
  await page.getByRole('button', { name: 'Suivant' }).click();
  await expect(page.locator('main ul > li')).toHaveCount(5);
  await expect(page.getByText('Page 2 sur 2 · 25 résultats')).toBeVisible();
});
