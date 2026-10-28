// Essais de bout en bout de la bibliothèque et de l'import PDF (backend simulé).

import { expect, test, type Page } from '@playwright/test';
import { installMockBackend } from './mock-backend';
import { BOOK, SCANNED } from './fixtures/make-pdf';

type Calls = Record<string, { pages: { number: number; lines: string[] }[] }[]>;

// Ouvre l'import PDF avec le fichier simulé donné.
async function openImport(page: Page, pdf: string) {
  await page.addInitScript(installMockBackend, { texts: [], pdf });
  await page.goto('/#/library?new=pdf'); // bouton « Importer un PDF » de l'accueil
  await page.getByRole('button', { name: 'Choisir un fichier PDF…' }).click();
}

// Dernier appel au nettoyage de l'import (vide s'il n'y en a pas encore).
const lastClean = (page: Page) =>
  page.evaluate(() => {
    // pas encore d'appel (nettoyage différé de 250 ms) : liste vide, expect.poll réessaie
    const calls = (window as unknown as { __calls: Calls }).__calls.CleanImport ?? [];
    return calls.at(-1)?.pages ?? [];
  });

test('import PDF : plage 3-5, miniatures, nettoyage, enregistrement (F-21, F-22)', async ({
  page,
}) => {
  await openImport(page, BOOK);

  await expect(page.getByText('livre.pdf')).toBeVisible();
  await expect(page.getByText('5 pages · texte détecté')).toBeVisible();
  const thumbs = page.getByTestId('thumbs').getByRole('button');
  await expect(thumbs).toHaveCount(5);
  await expect(thumbs.nth(2)).toHaveAttribute('aria-pressed', 'true'); // 1-3 par défaut

  // plage invalide puis 3-5 ; les miniatures suivent
  const range = page.getByLabel(/Pages à importer/);
  await range.fill('9');
  await expect(page.getByText('Plage invalide')).toBeVisible();
  await range.fill('3-5');
  await expect(thumbs.nth(0)).toHaveAttribute('aria-pressed', 'false');
  await expect(thumbs.nth(4)).toHaveAttribute('aria-pressed', 'true');

  // texte réellement extrait par pdfjs, envoyé au nettoyage page par page
  await expect.poll(async () => (await lastClean(page)).map((p) => p.number)).toEqual([3, 4, 5]);
  const sent = await lastClean(page);
  expect(sent[0]!.lines).toContain('Voici un exem-');
  expect(sent[1]!.lines).toContain('Texte de la page 4, avec des accents : été, à, ça.');

  // compteurs et aperçu surligné
  await expect(page.getByText('Recoller les mots coupés (exem-ple) (1)')).toBeVisible();
  await expect(page.getByText('Supprimer en-têtes et n° de page (4)')).toBeVisible();
  await expect(page.getByTestId('pdf-preview').locator('mark')).toHaveText('exemple');
  await page.screenshot({ path: 'test-results/screens/pdf-panel.png', fullPage: true });

  // choix des lignes (F-21) : lignes 2 à 14
  await page.getByText('Choisir les lignes (début et fin)').click();
  await page.getByLabel('Première ligne').fill('2');
  await page.getByLabel('Dernière ligne').fill('14');
  await page.getByLabel('Dernière ligne').press('Tab');
  await expect.poll(async () => (await lastClean(page)).flatMap((p) => p.lines).length).toBe(13);

  // aperçu modifiable
  await page.getByRole('button', { name: 'Modifier le texte' }).click();
  await page.getByLabel('Aperçu modifiable').fill('Texte corrigé à la main.');
  await page.getByLabel('Titre (facultatif)').fill('Mon livre');

  await page.getByRole('button', { name: 'Enregistrer 3 pages · 3 exercices' }).click();
  const created = await page.evaluate(
    () =>
      (
        window as unknown as {
          __calls: Record<string, { source: string; pages: string; content: string }[]>;
        }
      ).__calls.CreateText![0],
  );
  expect(created).toMatchObject({
    source: 'pdf',
    pages: '3-5',
    sourceFile: 'livre.pdf',
    content: 'Texte corrigé à la main.',
    split: 'paragraph',
  });
  await expect(page.getByText('Mon livre')).toBeVisible();
  await page.screenshot({ path: 'test-results/screens/library-list.png', fullPage: true });
});

test('PDF scanné : message clair', async ({ page }) => {
  await openImport(page, SCANNED);
  await expect(page.getByText('aucun texte détecté (PDF scanné ?)')).toBeVisible();
  await expect(page.getByRole('alert')).toContainText('probablement scanné');
});

test('bibliothèque : filtres, modifier, dupliquer, supprimer (F-23)', async ({ page }) => {
  await page.addInitScript(installMockBackend, {
    texts: [
      { title: 'Alpha', content: 'Premier texte.' },
      { title: 'Bravo', content: 'Second texte avec un chat.' },
    ],
  });
  await page.goto('/#/library');
  await expect(page.getByText('2 textes', { exact: true })).toBeVisible();

  const openMenu = (title: string) => page.getByLabel(`Plus d'actions pour « ${title} »`).click();

  await openMenu('Alpha');
  await page.getByRole('button', { name: 'Modifier « Alpha »' }).click();
  await page.getByLabel('Titre (facultatif)').fill('Alpha modifié');
  await page.getByRole('button', { name: 'Enregistrer les modifications' }).click();
  await expect(page.getByText('Alpha modifié')).toBeVisible();

  await openMenu('Bravo');
  await page.getByRole('button', { name: 'Dupliquer « Bravo »' }).click();
  await expect(page.getByText('Bravo (copie)')).toBeVisible();
  await expect(page.getByText('3 textes', { exact: true })).toBeVisible();

  await page.getByLabel('Rechercher').fill('chat');
  await expect(page.getByText('2 texte(s) sur 3')).toBeVisible();
  await page.getByLabel('Rechercher').fill('introuvable');
  await expect(page.getByText('Aucun texte ne correspond')).toBeVisible();
  await page.getByLabel('Rechercher').fill('');

  await openMenu('Bravo (copie)');
  await page.getByRole('button', { name: 'Supprimer « Bravo (copie) »' }).click();
  await page.getByRole('button', { name: 'Confirmer la suppression de « Bravo (copie) »' }).click();
  await expect(page.getByText('2 textes', { exact: true })).toBeVisible();
});
