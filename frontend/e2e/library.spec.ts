import { expect, test, type Page } from '@playwright/test';
import { installMockBackend } from './mock-backend';
import { BOOK, SCANNED } from './fixtures/make-pdf';

type Calls = Record<string, { pages: { number: number; lines: string[] }[] }[]>;

async function openLibrary(page: Page, pdf: string) {
  await page.addInitScript(installMockBackend, { texts: [], pdf });
  await page.goto('/#/library');
  await page.getByRole('button', { name: 'Importer un PDF' }).click();
  await page.getByRole('button', { name: 'Choisir un fichier PDF…' }).click();
}

test('import PDF : pages 3 à 5, sélection de lignes, nettoyage, exercice (F-21, F-22)', async ({
  page,
}) => {
  await openLibrary(page, BOOK);

  // Étape « pages » : aperçu et plage
  await expect(page.getByText('Page 1 / 5')).toBeVisible();
  await page.getByRole('button', { name: 'Page suivante' }).click();
  await expect(page.getByText('Page 2 / 5')).toBeVisible();
  const range = page.getByLabel(/Pages à importer/);
  await range.fill('9');
  await expect(page.getByText('Plage invalide')).toBeVisible();
  await range.fill('3-5');
  await page.screenshot({ path: 'test-results/screens/pdf-pages.png' });
  await page.getByRole('button', { name: 'Extraire le texte' }).click();

  // Étape « lignes » : texte réellement extrait par pdfjs
  const lines = page.getByTestId('pdf-lines');
  await expect(lines).toContainText('— page 3 —');
  await expect(lines).toContainText('Voici un exem-');
  await expect(lines).toContainText('été, à, ça');
  await expect(lines).not.toContainText('— page 2 —');
  await expect(page.getByText('15 ligne(s) sélectionnée(s) sur 15')).toBeVisible();
  // ne garder que les lignes 2 à 14 (sans le premier en-tête ni le dernier pied de page)
  await page.getByLabel('Première ligne').fill('2');
  await page.getByLabel('Nombre de lignes').fill('13');
  await page.getByLabel('Nombre de lignes').press('Tab');
  await expect(page.getByText('13 ligne(s) sélectionnée(s) sur 15')).toBeVisible();
  await page.screenshot({ path: 'test-results/screens/pdf-lines.png' });
  await page.getByRole('button', { name: 'Nettoyer et prévisualiser' }).click();

  const calls = await page.evaluate(() => (window as unknown as { __calls: Calls }).__calls);
  const sent = calls.CleanImport![0]!.pages;
  expect(sent.map((p) => p.number)).toEqual([3, 4, 5]);
  expect(sent[0]!.lines[0]).toBe('Page 3 : début du chapitre.');
  expect(sent.at(-1)!.lines.at(-1)).toBe('Fin de la page.');
  expect(sent.flatMap((p) => p.lines)).toHaveLength(13);

  // Étape « vérifier » : aperçu modifiable puis enregistrement
  await expect(page.getByLabel('Titre (facultatif)')).toHaveValue('livre (p. 3-5)');
  await expect(page.getByLabel('Texte', { exact: true })).toHaveValue(/Voici un exem- ple/);
  await page.screenshot({ path: 'test-results/screens/pdf-review.png' });
  await page.getByRole('button', { name: 'Enregistrer' }).click();

  await expect(page.getByText('PDF : livre.pdf, p. 3-5')).toBeVisible();
  const created = await page.evaluate(
    () =>
      (window as unknown as { __calls: Record<string, { source: string; pages: string }[]> })
        .__calls.CreateText![0],
  );
  expect(created).toMatchObject({ source: 'pdf', pages: '3-5', sourceFile: 'livre.pdf' });

  await page.getByRole('link', { name: "S'entraîner" }).click();
  await expect(page.getByRole('heading', { name: 'livre (p. 3-5)' })).toBeVisible();
});

test('PDF scanné : message clair', async ({ page }) => {
  await openLibrary(page, SCANNED);
  await page.getByRole('button', { name: 'Extraire le texte' }).click();
  await expect(page.getByRole('alert')).toContainText('probablement scanné');
});

test('bibliothèque : modifier, dupliquer, rechercher, supprimer (F-23)', async ({ page }) => {
  await page.addInitScript(installMockBackend, {
    texts: [
      { title: 'Alpha', content: 'Premier texte.' },
      { title: 'Bravo', content: 'Second texte avec un chat.' },
    ],
  });
  await page.goto('/#/library');
  await expect(page.getByText('Mes textes (2)')).toBeVisible();

  await page.getByRole('button', { name: 'Modifier « Alpha »' }).click();
  await page.getByLabel('Titre (facultatif)').fill('Alpha modifié');
  await page.getByRole('button', { name: 'Enregistrer les modifications' }).click();
  await expect(page.getByText('Alpha modifié')).toBeVisible();

  await page.getByRole('button', { name: 'Dupliquer « Bravo »' }).click();
  await expect(page.getByText('Bravo (copie)')).toBeVisible();
  await expect(page.getByText('Mes textes (3)')).toBeVisible();

  await page.getByLabel('Rechercher').fill('chat');
  await expect(page.getByText('2 texte(s) sur 3')).toBeVisible();
  await page.getByLabel('Rechercher').fill('introuvable');
  await expect(page.getByText('Aucun texte ne correspond')).toBeVisible();
  await page.getByLabel('Rechercher').fill('');

  await page.getByRole('button', { name: 'Supprimer « Bravo (copie) »' }).click();
  await page.getByRole('button', { name: 'Confirmer la suppression de « Bravo (copie) »' }).click();
  await expect(page.getByText('Mes textes (2)')).toBeVisible();
  await page.screenshot({ path: 'test-results/screens/library-list.png' });
});
