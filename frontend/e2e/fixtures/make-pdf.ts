// Génère un PDF minimal (police Helvetica, encodage WinAnsi) pour les tests.
// Chaque page est une liste de lignes ; une page vide simule un scan.

// Échappe une chaîne pour un objet texte PDF.
function pdfString(s: string): string {
  let out = '';
  for (const ch of s) {
    const code = ch.charCodeAt(0);
    if (ch === '(' || ch === ')' || ch === '\\') out += '\\' + ch;
    else if (code > 126)
      out += '\\' + code.toString(8).padStart(3, '0'); // Latin-1 = WinAnsi ici
    else out += ch;
  }
  return out;
}

// Fabrique un PDF minimal (une ligne de texte par entrée), en base 64.
export function makePdf(pages: string[][]): string {
  const objects: string[] = [];
  const pageIds: number[] = [];
  // 1 catalogue, 2 arbre des pages, 3 police ; puis page + contenu par page
  pages.forEach((lines, i) => {
    const pageId = 4 + i * 2;
    pageIds.push(pageId);
    const body = lines
      .map((l, j) => `BT /F1 12 Tf 72 ${760 - j * 18} Td (${pdfString(l)}) Tj ET`)
      .join('\n');
    objects[pageId] =
      `<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] ` +
      `/Resources << /Font << /F1 3 0 R >> >> /Contents ${pageId + 1} 0 R >>`;
    objects[pageId + 1] = `<< /Length ${body.length} >>\nstream\n${body}\nendstream`;
  });
  objects[1] = '<< /Type /Catalog /Pages 2 0 R >>';
  objects[2] = `<< /Type /Pages /Kids [${pageIds.map((id) => `${id} 0 R`).join(' ')}] /Count ${pages.length} >>`;
  objects[3] = '<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>';

  let pdf = '%PDF-1.4\n';
  const offsets: number[] = [];
  for (let id = 1; id < objects.length; id++) {
    offsets[id] = pdf.length;
    pdf += `${id} 0 obj\n${objects[id]}\nendobj\n`;
  }
  const xref = pdf.length;
  pdf += `xref\n0 ${objects.length}\n0000000000 65535 f \n`;
  for (let id = 1; id < objects.length; id++) {
    pdf += `${String(offsets[id]).padStart(10, '0')} 00000 n \n`;
  }
  pdf += `trailer\n<< /Size ${objects.length} /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`;
  return Buffer.from(pdf, 'latin1').toString('base64');
}

/** Livre de test : en-tête répété, numéros de page, césure en page 3. */
export const BOOK = makePdf(
  [1, 2, 3, 4, 5].map((n) => [
    'LE PETIT LIVRE',
    `Page ${n} : début du chapitre.`,
    n === 3 ? 'Voici un exem-' : `Texte de la page ${n}, avec des accents : été, à, ça.`,
    n === 3 ? 'ple de mot coupé.' : 'Fin de la page.',
    `${n}`,
  ]),
);

/** PDF « scanné » : une page sans texte. */
export const SCANNED = makePdf([[]]);
