// Phase 3 sur l'application réelle (wails dev, base jetable neuve) :
// mots difficiles, révision depuis l'accueil, tableau de bord.
import { expect, test, type Page } from '@playwright/test';

const SHOTS = process.env.SHOTS_DIR ?? 'test-results/real';
const TEXT = "Le maître et l'élève observent le chêne. Le maître explique et l'élève écoute.";

test.use({
  baseURL: 'http://localhost:34115',
  channel: 'msedge',
  viewport: { width: 1100, height: 800 },
});
test.setTimeout(120_000);

async function typeText(page: Page, s: string) {
  await page.keyboard.type(s, { delay: 8 });
}

test('phase 3 : mots difficiles, révision, statistiques', async ({ page }) => {
  page.on('pageerror', (e) => {
    if (!e.stack?.includes('/wails/ipc.js')) console.log('PAGE ERROR', e.stack);
  });

  // 1. Un texte tapé avec des fautes d'accent
  await page.goto('/#/library');
  await page.getByRole('button', { name: 'Coller un texte' }).click();
  await page.getByLabel('Titre (facultatif)').fill('Le maître et l’élève');
  await page.getByLabel('Texte', { exact: true }).fill(TEXT);
  await page.getByRole('button', { name: 'Enregistrer' }).click();
  // attendre la liste rafraîchie et viser la ligne du texte collé
  const item = page.locator('li', { hasText: 'Le maître et l’élève' });
  await item.getByRole('link', { name: 'Ouvrir' }).click();
  await expect(page.getByRole('heading', { name: 'Le maître et l’élève' })).toBeVisible();
  await expect(page.getByLabel('Zone de frappe')).toBeFocused();
  await typeText(
    page,
    TEXT.replace(/maître/g, 'maitre').replace("l'élève observent", "l'eleve observent"),
  );
  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible();
  console.log(
    'MOTS FAUTIFS :',
    JSON.stringify(await page.locator('li:has(.line-through)').allInnerTexts()),
  );

  // 2. Le jour même : rien à réviser, mais les mots sont suivis
  await page.goto('/#/review');
  await expect(page.getByTestId('due-count')).toHaveText("Aucun mot à réviser aujourd'hui");
  await page.getByRole('tab', { name: 'En cours' }).click();
  const rows = page.locator('tbody tr');
  await expect(rows.filter({ hasText: 'maître' })).toContainText('maitre');
  await expect(rows.filter({ hasText: 'maître' })).toContainText('accent');
  await expect(rows.filter({ hasText: 'élève' })).toBeVisible();
  // l'élision ne fait pas partie du mot suivi : « l'élève » → « élève »
  await expect(page.locator('tbody tr td:first-child')).toHaveText(['maître', 'élève']);
  await page.screenshot({ path: `${SHOTS}/p3-1-mots.png`, fullPage: true });

  // 3. Ajout manuel : à réviser aujourd'hui, proposé par l'accueil (F-43, F-45)
  await page.getByLabel('Ajouter un mot').fill('chêne');
  await page.getByRole('button', { name: 'Ajouter', exact: true }).click();
  await expect(page.getByTestId('due-count')).toHaveText("1 mot à réviser aujourd'hui");

  await page.goto('/#/');
  await expect(page.getByTestId('due-card')).toContainText("1 mot à réviser aujourd'hui");
  const banner = page.getByTestId('due-banner');
  await page.screenshot({ path: `${SHOTS}/p3-2-accueil.png` });
  await banner.click();

  // 4. Séance de révision : liste mélangée (« chêne » × 15)
  await expect(page.getByRole('heading', { name: /Révision \(1 mot\)/ })).toBeVisible();
  await expect(page.locator('.typing')).toContainText('chêne chêne');
  await typeText(page, Array(15).fill('chêne').join(' '));
  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible();
  await page.screenshot({ path: `${SHOTS}/p3-3-revision.png`, fullPage: true });
  await page.getByRole('link', { name: 'Retour aux révisions' }).click();
  await expect(page.getByTestId('due-count')).toHaveText("Aucun mot à réviser aujourd'hui");
  await page.getByRole('tab', { name: 'Tous' }).click();
  await expect(rows.filter({ hasText: 'chêne' })).toContainText('33 %');

  // 5. Tableau de bord
  await page.goto('/#/stats');
  await expect(page.getByTestId('totals')).toContainText('2');
  await expect(page.locator('canvas')).toHaveCount(2);
  const colored = page.locator('[data-testid="heatmap"] [data-step]:not([data-step=""])');
  expect(await colored.count()).toBeGreaterThan(5);
  console.log('TOTAUX :', (await page.getByTestId('totals').innerText()).replace(/\n/g, ' | '));
  console.log('BIGRAMMES :', (await page.getByTestId('bigrams').innerText()).replace(/\n/g, ' | '));
  await page.screenshot({ path: `${SHOTS}/p3-4-stats.png`, fullPage: true });

  // Accueil rempli (maquette « Accueil »)
  await page.goto('/#/settings');
  await page.getByLabel('Votre prénom').fill('Boualem');
  await page.getByRole('button', { name: 'Enregistrer' }).click();
  await page.goto('/#/');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Bonjour Boualem.');
  await expect(page.getByText('Sessions récentes')).toBeVisible();
  await page.screenshot({ path: `${SHOTS}/p3-5-accueil.png`, fullPage: true });
  await page.goto('/#/library');
  await expect(page.locator('.eyebrow', { hasText: /\d+ textes/ })).toBeVisible();
  await page.screenshot({ path: `${SHOTS}/p3-6-bibliotheque.png`, fullPage: true });
});
