// Essai sur l'application réelle (wails dev, vrai backend Go), base jetable.
// La base peut contenir des textes ajoutés à la main : rien n'y est compté
// en dur.
import { expect, test, type Page } from '@playwright/test';

const SHOTS = process.env.SHOTS_DIR ?? 'test-results/real';

test.use({
  baseURL: 'http://localhost:34115',
  channel: 'msedge',
  viewport: { width: 1100, height: 760 },
});
test.setTimeout(120_000);

function watchErrors(page: Page) {
  page.on('pageerror', (e) => {
    if (!e.stack?.includes('/wails/ipc.js')) console.log('PAGE ERROR', e.stack);
  });
}

test('phase 2 : pack intégré, recherche, exercices et reprise', async ({ page }) => {
  watchErrors(page);
  await page.goto('/#/library');

  // F-25 : pack intégré
  const heading = page.locator('.eyebrow', { hasText: /\d+ textes/i });
  await expect(heading).toBeVisible();
  const total = Number(/(\d+)/.exec(await heading.innerText())![1]);
  expect(total).toBeGreaterThanOrEqual(8);
  await expect(page.getByText('Le Dormeur du val').first()).toBeVisible();
  await page.screenshot({ path: `${SHOTS}/p2-1-pack.png`, fullPage: true });
  const filtered = page.getByText(new RegExp(`\\d+ texte\\(s\\) sur ${total}`));

  // F-23 : recherche sans accents, filtre par étiquette, tri
  await page.getByLabel('Rechercher').fill('maitre corbeau');
  await expect(filtered).toBeVisible();
  await expect(page.getByText('Le Corbeau et le Renard').first()).toBeVisible();
  await expect(page.getByText('Le Dormeur du val')).toBeHidden();
  await page.getByLabel('Rechercher').fill('');
  await page.getByRole('radio', { name: 'poésie' }).click();
  await expect(filtered).toBeVisible();
  await expect(page.getByText('Le Corbeau et le Renard')).toBeHidden();
  await page.getByLabel('Trier par').selectOption('title');
  await expect(page.locator('li p.font-bold').first()).toContainText('Chanson');
  const titles = await page.locator('li p.font-bold').allInnerTexts();
  console.log('POÉSIE TRIÉE :', JSON.stringify(titles.map((t) => t.split('\n')[0])));
  await page.getByRole('radio', { name: 'Tous' }).click();

  // F-24 : « Demain, dès l'aube » découpé par paragraphe (3 strophes)
  await page.goto('/#/practice');
  await page.getByLabel('Découper le texte').selectOption('paragraph');
  await page.getByLabel('Paragraphes par exercice').fill('1');
  await page
    .getByRole('link', { name: /Demain, dès l'aube/ })
    .first()
    .click();
  const restart = page.getByRole('button', { name: 'Reprendre le texte au début' });
  await expect(restart).toBeVisible();
  await restart.click(); // repartir du début si un essai précédent a laissé une progression
  await expect(page.getByRole('combobox', { name: /^Exercice/ })).toHaveValue('0');
  const area = page.locator('.typing');
  await expect(area).toContainText('Demain, dès l');
  await expect(area).not.toContainText('Je marcherai');

  // Exercice 1 tapé en entier (Entrée pour les fins de vers)
  const strophe1 = [
    "Demain, dès l'aube, à l'heure où blanchit la campagne,",
    "Je partirai. Vois-tu, je sais que tu m'attends.",
    "J'irai par la forêt, j'irai par la montagne.",
    'Je ne puis demeurer loin de toi plus longtemps.',
  ];
  for (const [i, line] of strophe1.entries()) {
    await page.keyboard.type(line, { delay: 5 });
    if (i < strophe1.length - 1) await page.keyboard.press('Enter');
  }
  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible();
  await page.screenshot({ path: `${SHOTS}/p2-2-fin-exercice1.png`, fullPage: true });
  console.log(
    'RÉSULTAT EX1 :',
    (await page.locator('dl').first().innerText()).replace(/\n/g, ' | '),
  );

  // Ctrl+Entrée : exercice suivant
  await page.keyboard.press('Control+Enter');
  await expect(page.getByRole('combobox', { name: /^Exercice/ })).toHaveValue('1');
  await expect(area).toContainText('Je marcherai');

  // Exercice 2 commencé puis interrompu
  await page.keyboard.type('Je marcherai les yeux fix', { delay: 5 });
  await page.getByRole('button', { name: 'Terminer maintenant' }).click();
  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible();

  // Retour par la bibliothèque : progression et reprise
  await page.goto('/#/library');
  await page.getByLabel('Rechercher').fill('demain, des');
  const progress = page.getByRole('progressbar', { name: /Parcouru : [1-9][0-9]* %/ }).first();
  await expect(progress).toBeVisible();
  console.log('PROGRESSION :', await progress.getAttribute('aria-label'));
  await page.getByRole('link', { name: 'Reprendre' }).first().click();
  await expect(page.getByTestId('resumed')).toBeVisible();
  await expect(page.getByRole('combobox', { name: /^Exercice/ })).toHaveValue('1');
  // reprise au début du mot interrompu « fixés »
  await expect(page.locator('.typing .ch-cursor')).toHaveText('f');
  await page.screenshot({ path: `${SHOTS}/p2-3-reprise.png`, fullPage: true });
});
