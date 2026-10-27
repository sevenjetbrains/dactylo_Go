// Capture de la frappe sur une zone de saisie cachée (section 2.3) :
// on lit le texte produit par le navigateur (input / compositionend), jamais
// les touches (keydown), pour que les touches mortes AZERTY (^ + e = ê)
// donnent un seul caractère.

export interface TypingInputHandlers {
  onChar(ch: string): void;
  onBackspace(): void;
}

export function attachTypingInput(el: HTMLTextAreaElement, h: TypingInputHandlers): () => void {
  let composing = false;

  const flush = () => {
    const value = el.value;
    el.value = '';
    for (const ch of value.normalize('NFC')) {
      h.onChar(ch === '\r' ? '\n' : ch);
    }
  };

  const onBeforeInput = (e: InputEvent) => {
    const t = e.inputType;
    if (t === 'deleteContentBackward' || t === 'deleteWordBackward') {
      e.preventDefault();
      h.onBackspace();
    } else if (t.startsWith('delete')) {
      e.preventDefault();
    } else if (t === 'insertLineBreak' || t === 'insertParagraph') {
      e.preventDefault();
      h.onChar('\n');
    } else if (t === 'insertFromPaste' || t === 'insertFromDrop') {
      e.preventDefault(); // pas de copier-coller dans un exercice de frappe
    }
  };

  const onInput = (e: Event) => {
    if (composing || (e as InputEvent).isComposing) return;
    flush();
  };
  const onCompositionStart = () => {
    composing = true;
  };
  const onCompositionEnd = () => {
    composing = false;
    flush();
  };

  el.addEventListener('beforeinput', onBeforeInput);
  el.addEventListener('input', onInput);
  el.addEventListener('compositionstart', onCompositionStart);
  el.addEventListener('compositionend', onCompositionEnd);
  return () => {
    el.removeEventListener('beforeinput', onBeforeInput);
    el.removeEventListener('input', onInput);
    el.removeEventListener('compositionstart', onCompositionStart);
    el.removeEventListener('compositionend', onCompositionEnd);
  };
}
