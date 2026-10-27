// Essai sur l'application réelle (wails dev, vrai backend Go), base jetable.
import { expect, test } from '@playwright/test';

const SHOTS = process.env.SHOTS_DIR ?? 'test-results/real';
const TEXT =
  'Maître corbeau, sur un arbre perché, tenait en son bec un fromage. Maître renard, par l’odeur alléché, lui tint à peu près ce langage.';

test.use({
  baseURL: 'http://localhost:34115',
  channel: 'msedge',
  viewport: { width: 1100, height: 760 },
});
test.setTimeout(90_000);

test('parcours réel : coller, taper avec fautes, résultat, seconde session', async ({ page }) => {
  page.on('pageerror', (e) => console.log('PAGE ERROR', page.url(), e.stack));
  page.on(
    'console',
    (m) => m.type() === 'error' && console.log('CONSOLE', m.text(), m.location().url),
  );

  await page.goto('/#/');
  await expect(page.getByTestId('system-info')).toContainText('schéma de données v2');
  await page.screenshot({ path: `${SHOTS}/1-accueil.png` });

  // Bibliothèque : coller un texte
  await page.getByRole('link', { name: 'Bibliothèque' }).click();
  await page.getByLabel('Titre (facultatif)').fill('Le Corbeau et le Renard');
  await page.getByLabel('Texte', { exact: true }).fill(TEXT);
  await page.getByLabel('Étiquettes, séparées par des virgules').fill('fable, La Fontaine');
  await page.getByRole('button', { name: 'Enregistrer' }).click();
  await expect(page.getByText('Le Corbeau et le Renard').first()).toBeVisible();
  await page.screenshot({ path: `${SHOTS}/2-bibliotheque.png` });

  // Session 1 : avec des fautes (accent oublié, lettre erronée) et une correction
  await page.getByRole('link', { name: "S'entraîner" }).first().click();
  await expect(page.getByLabel('Zone de frappe')).toBeFocused();
  const typed = TEXT.replace('Maître corbeau', 'Maitre corbeau') // accent
    .replace('fromage', 'fromaje') // substitution
    .replace('’', "'"); // apostrophe droite acceptée
  const chars = [...typed];
  for (let i = 0; i < chars.length; i++) {
    await page.keyboard.type(chars[i]!, { delay: 15 });
    if (i === 40) {
      // faute corrigée : on tape un mauvais caractère puis Retour arrière
      await page.keyboard.type('x');
      await page.keyboard.press('Backspace');
    }
    if (i === 60) await page.screenshot({ path: `${SHOTS}/3-frappe.png` });
  }
  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible({
    timeout: 15_000,
  });
  await page.screenshot({ path: `${SHOTS}/4-resultat-1.png`, fullPage: true });
  const mistakes = await page.locator('li:has(.line-through)').allInnerTexts();
  console.log('MOTS FAUTIFS 1 :', JSON.stringify(mistakes));
  console.log('COMPARAISON 1 :', await page.getByTestId('comparison').innerText());

  // Session 2 : sans faute, via Ctrl+R
  await page.keyboard.press('Control+r');
  await expect(page.getByText('Commencez à taper')).toBeVisible();
  await page.keyboard.type(TEXT.replace('’', "'"), { delay: 10 });
  await expect(page.getByRole('heading', { name: 'Session terminée' })).toBeVisible({
    timeout: 15_000,
  });
  await page.screenshot({ path: `${SHOTS}/5-resultat-2.png`, fullPage: true });
  console.log('COMPARAISON 2 :', await page.getByTestId('comparison').innerText());
  console.log('STATS 2 :', (await page.locator('dl').first().innerText()).replace(/\n/g, ' | '));
});
