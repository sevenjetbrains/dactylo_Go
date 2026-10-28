// Phase 5 sur l'application réelle (wails dev, vrai backend Go, base
// jetable) : profils et réglages par profil, disposition BÉPO, reprise
// d'une session interrompue.
import { expect, test, type Page } from '@playwright/test';

const SHOTS = process.env.SHOTS_DIR ?? 'test-results/real';

test.use({
  baseURL: 'http://localhost:34115',
  channel: 'msedge',
  viewport: { width: 1440, height: 1000 },
});
test.setTimeout(120_000);

const sizeLabel = (page: Page) => page.getByText(/^Taille du texte : \d+ px$/);
const sizeSlider = (page: Page) =>
  page.locator('label', { has: sizeLabel(page) }).locator('input[type=range]');

test('profils : création, bascule, réglages propres à chacun (F-50, F-51, F-08)', async ({
  page,
}) => {
  await page.goto('/#/settings');
  const rows = page.getByTestId('profile-row');
  await expect(rows).toHaveCount(1);

  await page.getByLabel('Nouveau profil').fill('Léa');
  await page
    .getByTestId('profiles')
    .getByLabel('Disposition du clavier')
    .last()
    .selectOption('bepo');
  await page.getByRole('button', { name: 'Ajouter le profil' }).click();
  await expect(rows).toHaveCount(2);

  // bascule : l'interface est rechargée avec le profil de Léa
  await rows.nth(1).getByRole('button', { name: 'Utiliser ce profil' }).click();
  await expect(page.getByRole('link', { name: 'Profil' })).toContainText('Léa');
  await expect(page.getByRole('link', { name: 'Profil' })).toContainText('BÉPO');
  await expect(sizeLabel(page)).toHaveText('Taille du texte : 24 px');
  await sizeSlider(page).fill('30');
  await expect(sizeLabel(page)).toHaveText('Taille du texte : 30 px');
  await page.waitForTimeout(1200); // enregistrement différé dans le profil
  await page.screenshot({ path: `${SHOTS}/p5-settings.png`, fullPage: true });

  // clavier virtuel BÉPO : touche « ê » dédiée
  await page.goto('/#/practice');
  await page.locator('a[href^="#/practice?text="]').first().click();
  await expect(page.locator('[data-key="ê"]')).toBeVisible();
  await page.screenshot({ path: `${SHOTS}/p5-bepo.png` });

  // retour au premier profil : ses propres réglages, son clavier AZERTY
  await page.goto('/#/settings');
  await rows.nth(0).getByRole('button', { name: 'Utiliser ce profil' }).click();
  await expect(page.getByRole('link', { name: 'Profil' })).toContainText('AZERTY');
  await expect(sizeLabel(page)).toHaveText('Taille du texte : 24 px');

  // et ceux de Léa ont bien été gardés
  await rows.nth(1).getByRole('button', { name: 'Utiliser ce profil' }).click();
  await expect(page.getByRole('link', { name: 'Profil' })).toContainText('Léa');
  await expect(sizeLabel(page)).toHaveText('Taille du texte : 30 px');

  // suppression de Léa (profil actif) : retour sur l'autre
  await rows.nth(1).getByRole('button', { name: 'Supprimer' }).click();
  await rows.nth(1).getByRole('button', { name: 'Confirmer : tout supprimer' }).click();
  await expect(page.getByRole('link', { name: 'Profil' })).toContainText('AZERTY');
  await expect(rows).toHaveCount(1);
});

test('reprise après une fermeture brutale (NF-06)', async ({ page }) => {
  await page.goto('/#/practice');
  await page.locator('a[href^="#/practice?text="]').first().click();
  const input = page.getByLabel('Zone de frappe');
  await expect(input).toBeFocused();
  const chars = await page
    .locator('.typing .ch-todo')
    .evaluateAll((els) => els.slice(0, 12).map((e) => e.textContent ?? ''));
  const first = chars.join('').replace(/\n/g, ' ');
  await page.keyboard.type(first.slice(0, 10));
  await page.waitForTimeout(2500); // lot de frappes envoyé

  // « plantage » : rechargement sans quitter la session
  await page.goto('/#/');
  await page.reload();
  const box = page.getByTestId('interrupted');
  await expect(box).toContainText('10 frappes enregistrées');
  await page.screenshot({ path: `${SHOTS}/p5-interrupted.png` });

  await box.getByRole('link', { name: 'Reprendre' }).click();
  await expect(page.getByTestId('session-resumed')).toBeVisible();
  await expect(page.getByText(/^En pause/)).toBeVisible();
  await expect(page.locator('.typing .ch-ok')).toHaveCount(10);
  await page.getByLabel('Zone de frappe').focus();
  await page.keyboard.type(first.slice(10, 12));
  await expect(page.locator('.typing .ch-ok')).toHaveCount(12);
  await page.screenshot({ path: `${SHOTS}/p5-resumed.png` });

  // Ctrl+R : la session est abandonnée et n'est plus proposée
  await page.keyboard.press('Control+r');
  await page.waitForTimeout(500);
  await page.goto('/#/');
  await page.reload();
  await expect(page.getByTestId('due-card')).toBeVisible();
  await expect(page.getByTestId('interrupted')).toHaveCount(0);
});
