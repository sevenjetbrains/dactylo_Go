// Essais de bout en bout de l'accueil, de la Copie, de la reprise et des
// réglages (backend simulé).

import { expect, test, type Page } from '@playwright/test';
import { installMockBackend } from './mock-backend';

const TEXT = 'Le chat être là.';

// Ouvre l'application à l'adresse donnée, avec des textes simulés.
async function open(page: Page, hash: string, texts = [{ title: 'Essai', content: TEXT }]) {
  await page.addInitScript(installMockBackend, { texts });
  await page.goto('/#' + hash);
}

test('accueil et navigation', async ({ page }) => {
  await open(page, '/');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Bonjour.');
  await expect(page.getByTestId('system-info')).toContainText(/schéma de données v[0-9]+/);
  await page.screenshot({ path: 'test-results/screens/home.png' });

  await page.getByRole('navigation').getByRole('link', { name: 'Copie' }).click();
  await expect(page.getByRole('heading', { name: 'Mode Copie' })).toBeVisible();
  await page.screenshot({ path: 'test-results/screens/practice-picker.png' });
});

test('session complète : frappe, touche morte, correction, écran de fin', async ({ page }) => {
  await open(page, '/practice?text=1');
  const input = page.getByLabel('Zone de frappe');
  await expect(input).toBeFocused();

  await page.keyboard.type('Le chat ');
  // Touche morte AZERTY (^ puis e) : le navigateur produit une composition.
  await input.dispatchEvent('compositionstart');
  await page.keyboard.insertText('ê');
  await input.dispatchEvent('compositionend', { data: 'ê' });
  await page.keyboard.type('tro');
  await page.keyboard.press('Backspace');
  await page.keyboard.type('e l');
  await page.screenshot({ path: 'test-results/screens/practice-typing.png' });
  await page.keyboard.type('à.');

  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible();
  await page.screenshot({ path: 'test-results/screens/practice-result.png' });

  const calls = await page.evaluate(
    () =>
      (
        window as unknown as {
          __calls: Record<string, { keys: { typed: string; position: number }[] }[]>;
        }
      ).__calls,
  );
  const keys = calls.FinishSession![0]!.keys;
  // une seule frappe pour « ê », à la position 8
  expect(keys.filter((k) => k.position === 8 && k.typed !== '\b').map((k) => k.typed)).toEqual([
    'ê',
  ]);
  expect(keys.some((k) => k.typed === '\b')).toBe(true);
  expect(keys.at(-1)).toMatchObject({ typed: '.', position: TEXT.length - 1 });

  // Ctrl+R depuis l'écran de fin : la nouvelle zone de saisie doit réagir
  await page.keyboard.press('Control+r');
  await expect(page.getByLabel('Zone de frappe')).toBeFocused();
  await page.keyboard.type(TEXT);
  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible();
});

test('Échap met en pause, Ctrl+R recommence', async ({ page }) => {
  await open(page, '/practice?text=1');
  await page.keyboard.type('Le');
  await page.keyboard.press('Escape');
  await expect(page.getByText('En pause')).toBeVisible();
  await page.keyboard.type(' ');
  await expect(page.getByText('En pause')).toBeHidden();
  await page.keyboard.press('Control+r');
  await expect(page.getByText('Commencez à taper')).toBeVisible();
  await expect(page.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '0');
});

test('texte de 10 000 mots : frappe fluide (NF-01)', async ({ page }) => {
  const long = Array.from({ length: 10_000 }, (_, i) => ['le', 'chat', 'être', 'là'][i % 4]).join(
    ' ',
  );
  await open(page, '/practice?text=1', [{ title: 'Long', content: long }]);
  await expect(page.getByLabel('Zone de frappe')).toBeFocused();
  // seules les lignes visibles sont rendues
  expect(await page.locator('.typing span').count()).toBeLessThan(600);

  const t0 = Date.now();
  await page.keyboard.type('le chat être là le chat être là ', { delay: 0 });
  const perKey = (Date.now() - t0) / 32;
  expect(perKey).toBeLessThan(50);
  // les caractères restent alignés sur la grille monospace
  const xs = await page
    .locator('.typing div.whitespace-pre')
    .first()
    .locator('span')
    .evaluateAll((ss) => ss.map((s) => s.getBoundingClientRect().x));
  const step = xs[1]! - xs[0]!;
  xs.forEach((x, i) => expect(Math.abs(x - xs[0]! - i * step)).toBeLessThan(1));
  await expect(page.getByRole('progressbar')).not.toHaveAttribute('aria-valuenow', '100');
});

test('bibliothèque : coller un texte', async ({ page }) => {
  await open(page, '/library', []);
  await page.getByRole('button', { name: 'Coller un texte' }).click();
  await page.getByLabel('Texte', { exact: true }).fill('Un texte collé pour l’exercice.');
  await page.getByLabel('Titre (facultatif)').fill('Collé');
  await page.getByRole('button', { name: 'Enregistrer' }).click();
  await expect(page.getByText('1 texte', { exact: true })).toBeVisible();
  await page.screenshot({ path: 'test-results/screens/library.png' });
  await page.getByRole('link', { name: 'Ouvrir' }).click();
  await expect(page.getByRole('heading', { name: 'Collé' })).toBeVisible();
});

test('reprise d’une session interrompue (NF-06)', async ({ page }) => {
  await open(page, '/practice?text=1');
  await expect(page.getByLabel('Zone de frappe')).toBeFocused();
  await page.keyboard.type('Le chat ');
  // les frappes partent par lots (2 s)
  await expect
    .poll(() =>
      page.evaluate(
        () =>
          (
            (window as unknown as { __calls: Record<string, unknown[][]> }).__calls
              .RecordKeystrokes ?? []
          ).flat().length,
      ),
    )
    .toBe(8);

  await page.evaluate(() => (window.location.hash = '#/practice?resume=1'));
  await expect(page.getByTestId('session-resumed')).toBeVisible();
  await expect(page.getByText('En pause')).toBeVisible();
  await page.getByLabel('Zone de frappe').focus();
  await page.keyboard.type('être là.');
  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible();

  const finish = await page.evaluate(
    () =>
      (
        window as unknown as {
          __calls: Record<string, { id: number; keys: { typed: string }[] }[]>;
        }
      ).__calls.FinishSession![0]!,
  );
  expect(finish.id).toBe(1);
  expect(finish.keys.map((k) => k.typed).join('')).toBe(TEXT);
});

test('réglages : profils et sauvegarde (F-50, F-52)', async ({ page }) => {
  await open(page, '/settings');
  await expect(page.getByTestId('profile-row')).toHaveCount(1);
  await expect(page.getByTestId('profile-row')).toContainText('Profil actif');
  await page.getByRole('button', { name: 'Sauvegarder dans un fichier…' }).click();
  await expect(page.getByText('Sauvegarde enregistrée : C:/sauvegarde.db')).toBeVisible();
  await page.getByRole('button', { name: 'Restaurer une sauvegarde…' }).click();
  await expect(page.getByText('Les données actuelles seront remplacées')).toBeVisible();
  await page.screenshot({ path: 'test-results/screens/settings.png', fullPage: true });
});
