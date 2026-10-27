import { beforeEach, describe, expect, it } from 'vitest';
import { attachTypingInput } from './input';

let el: HTMLTextAreaElement;
let chars: string[];
let backspaces: number;

beforeEach(() => {
  el = document.createElement('textarea');
  document.body.append(el);
  chars = [];
  backspaces = 0;
  attachTypingInput(el, {
    onChar: (c) => chars.push(c),
    onBackspace: () => backspaces++,
  });
});

function input(value: string, isComposing = false) {
  el.value += value;
  el.dispatchEvent(new InputEvent('input', { data: value, isComposing }));
}

function beforeInput(inputType: string) {
  const e = new InputEvent('beforeinput', { inputType, cancelable: true });
  el.dispatchEvent(e);
  return e.defaultPrevented;
}

describe('attachTypingInput', () => {
  it('transmet chaque caractère tapé et vide la zone', () => {
    input('a');
    input('b');
    expect(chars).toEqual(['a', 'b']);
    expect(el.value).toBe('');
  });

  it('touche morte par composition : ê compte pour un caractère', () => {
    el.dispatchEvent(new CompositionEvent('compositionstart'));
    input('^', true);
    el.value = 'ê';
    input('', true);
    expect(chars).toEqual([]);
    el.dispatchEvent(new CompositionEvent('compositionend', { data: 'ê' }));
    expect(chars).toEqual(['ê']);
  });

  it('normalise en NFC une saisie décomposée', () => {
    input('ê');
    expect(chars).toEqual(['ê']);
  });

  it('Retour arrière, Entrée et collage', () => {
    expect(beforeInput('deleteContentBackward')).toBe(true);
    expect(backspaces).toBe(1);
    expect(beforeInput('deleteContentForward')).toBe(true);
    expect(backspaces).toBe(1);
    expect(beforeInput('insertLineBreak')).toBe(true);
    expect(chars).toEqual(['\n']);
    expect(beforeInput('insertFromPaste')).toBe(true);
    expect(beforeInput('insertText')).toBe(false);
  });
});
