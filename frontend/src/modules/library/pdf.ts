// Lecture des PDF avec pdfjs-dist, entièrement hors-ligne : le worker est
// embarqué dans l'application (NF-03).
import * as pdfjs from 'pdfjs-dist';
import workerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url';
import { linesFromItems, type PdfTextItem } from './pdfLines';

pdfjs.GlobalWorkerOptions.workerSrc = workerUrl;

export type PdfDocument = pdfjs.PDFDocumentProxy;

/** Ouvre un PDF reçu du backend (octets en base64). `close` libère le worker. */
export async function loadPdf(
  base64: string,
): Promise<{ doc: PdfDocument; close: () => Promise<void> }> {
  const res = await fetch(`data:application/pdf;base64,${base64}`);
  const data = new Uint8Array(await res.arrayBuffer());
  const task = pdfjs.getDocument({ data });
  return { doc: await task.promise, close: () => task.destroy() };
}

/** Lignes de texte d'une page (vide pour une page scannée). */
export async function pageLines(doc: PdfDocument, pageNumber: number): Promise<string[]> {
  const page = await doc.getPage(pageNumber);
  const content = await page.getTextContent();
  const items = content.items.filter((i): i is PdfTextItem & typeof i => 'str' in i);
  return linesFromItems(items);
}

/** Dessine l'aperçu d'une page dans le canvas, à la largeur donnée. */
export async function renderPage(
  doc: PdfDocument,
  pageNumber: number,
  canvas: HTMLCanvasElement,
  width: number,
): Promise<void> {
  const page = await doc.getPage(pageNumber);
  const base = page.getViewport({ scale: 1 });
  const ratio = window.devicePixelRatio || 1;
  const viewport = page.getViewport({ scale: (width / base.width) * ratio });
  canvas.width = viewport.width;
  canvas.height = viewport.height;
  canvas.style.width = `${viewport.width / ratio}px`;
  canvas.style.height = `${viewport.height / ratio}px`;
  await page.render({ canvas, viewport }).promise;
}
