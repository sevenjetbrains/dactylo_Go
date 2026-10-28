// Phase 4 sur l'application réelle (wails dev, vrai backend Go) : dictée
// complète. La synthèse vocale est remplacée par une voix factice silencieuse
// qui enregistre ce qu'elle lit (window.__spoken).
import { expect, test, type Page } from '@playwright/test';

const SHOTS = process.env.SHOTS_DIR ?? 'test-results/real';
const TEXT =
  "En 1815, M. Myriel était évêque de Digne.\n\nC'était un vieillard d'environ soixante-quinze ans.";

test.use({
  baseURL: 'http://localhost:34115',
  channel: 'msedge',
  viewport: { width: 1440, height: 1000 },
});
test.setTimeout(120_000);

function fakeSpeech() {
  const spoken: string[] = [];
  (window as unknown as { __spoken: string[] }).__spoken = spoken;
  const voices = [
    { voiceURI: 'fr-denise', name: 'Microsoft Denise', lang: 'fr-FR', localService: true },
  ];
  class Utterance {
    voice: unknown = null;
    lang = '';
    rate = 1;
    pitch = 1;
    volume = 1;
    onend: (() => void) | null = null;
    onerror: ((e: { error: string }) => void) | null = null;
    constructor(public text: string) {}
  }
  let current: Utterance | null = null;
  const synth = {
    getVoices: () => voices,
    addEventListener: () => {},
    speak: (u: Utterance) => {
      spoken.push(u.text);
      current = u;
      setTimeout(() => current === u && u.onend?.(), 150);
    },
    cancel: () => {
      const u = current;
      current = null;
      u?.onerror?.({ error: 'interrupted' });
    },
  };
  Object.defineProperty(window, 'speechSynthesis', { value: synth, configurable: true });
  Object.defineProperty(window, 'SpeechSynthesisUtterance', {
    value: Utterance,
    configurable: true,
  });
}

const spoken = (page: Page) =>
  page.evaluate(() => (window as unknown as { __spoken: string[] }).__spoken);

test('phase 4 : dictée lue, adaptative, corrigée et reliée aux révisions', async ({ page }) => {
  page.on('pageerror', (e) => {
    if (!e.stack?.includes('/wails/ipc.js')) console.log('PAGE ERROR', e.stack);
  });
  await page.addInitScript(fakeSpeech);

  await page.goto('/#/library?new=paste');
  await page.getByLabel('Titre (facultatif)').fill('Les Misérables — l’évêque de Digne');
  await page.getByLabel('Texte', { exact: true }).fill(TEXT);
  await page.getByRole('button', { name: 'Enregistrer' }).click();

  await page.goto('/#/dictation');
  await page
    .locator('li', { hasText: 'Les Misérables — l’évêque de Digne' })
    .getByRole('link', { name: 'Lancer la dictée' })
    .first() // le plus récent, si un essai précédent a laissé un exemplaire
    .click();
  await expect(page.getByText('Dictée · 14 mots')).toBeVisible();
  await expect(page.getByRole('combobox', { name: 'Voix' })).toHaveValue('fr-denise');

  // lecture du premier segment, ponctuation dite (F-13), abréviation intacte
  await page.getByRole('button', { name: 'Lire' }).click();
  await expect.poll(() => spoken(page)).toHaveLength(1);
  expect((await spoken(page))[0]).toBe('En 1815 virgule M. Myriel');
  await expect(page.getByTestId('segment-status')).toContainText('Segment 1 sur');

  // répéter (Ctrl+Espace)
  const copy = page.getByLabel('Votre copie');
  await copy.click();
  await page.keyboard.press('Control+Space');
  await expect.poll(() => spoken(page)).toHaveLength(2);
  await expect(page.getByText('Répétitions : 1 / illimitées')).toBeVisible();

  // avancement adaptatif : le segment suivant attend les mots tapés (F-15)
  await page.waitForTimeout(2500);
  expect(await spoken(page)).toHaveLength(2);
  await copy.pressSequentially('En 1815, M. Myriel ', { delay: 10 });
  await expect.poll(() => spoken(page), { timeout: 8000 }).toHaveLength(3);
  await page.screenshot({ path: `${SHOTS}/p4-1-dictee.png` });

  // la suite, avec 5 fautes connues
  await copy.pressSequentially(
    "était évèque de digne. C'était un viellard d'environs soixante quinze ans.",
    { delay: 5 },
  );
  await page.getByRole('button', { name: 'Terminer et corriger' }).click();

  // correction (F-16)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(
    '5 fautes sur 14 mots. Continuez !',
  );
  const kpis = page.getByTestId('dictation-kpis');
  await expect(kpis).toContainText('64,3 %');
  await expect(kpis).toContainText('1 répétition');
  const byType = page.getByTestId('by-type');
  for (const t of ['Accent', 'Majuscule', 'Lettre manquante', 'Lettre en trop', 'Ponctuation']) {
    await expect(byType.locator('li', { hasText: new RegExp(`^${t}`, 'i') })).toContainText('1');
  }
  await expect(page.getByTestId('corrected').locator('ins')).toHaveText([
    'évêque',
    'Digne.',
    'vieillard',
    "d'environ",
    'soixante-quinze',
  ]);
  const added = page.locator('tbody tr');
  await expect(added).toHaveCount(5);
  await page.screenshot({ path: `${SHOTS}/p4-2-correction.png`, fullPage: true });

  // « Retirer » puis « S'entraîner sur ces mots »
  await page.getByRole('button', { name: 'Retirer « Digne. » des révisions' }).click();
  await expect(added.filter({ hasText: 'Digne.' })).toContainText('Retiré');
  await page.getByRole('button', { name: "S'entraîner sur ces 5 mots" }).click();
  await expect(page.getByRole('heading', { name: /Révision \(5 mots\)/ })).toBeVisible();
  await expect(page.locator('.typing')).toContainText('vieillard');

  // les fautes sont suivies (sauf « Digne », retiré)
  await page.goto('/#/review');
  await page.getByRole('tab', { name: 'En cours' }).click();
  await expect(page.locator('tbody tr').first()).toBeVisible();
  const tracked = await page.locator('tbody tr td:first-child').allInnerTexts();
  console.log('MOTS SUIVIS :', JSON.stringify(tracked));
  expect(tracked).not.toContain('digne');
  expect(tracked).toEqual(
    expect.arrayContaining(['évêque', 'vieillard', 'environ', 'soixante-quinze']),
  );

  // accueil : carte Dictée
  await page.goto('/#/');
  await expect(page.getByTestId('dictation-card')).toContainText(
    'Les Misérables — l’évêque de Digne',
  );
  await expect(page.getByTestId('dictation-card')).toContainText('Groupes de 4 mots');
  await page.screenshot({ path: `${SHOTS}/p4-3-accueil.png` });
});
