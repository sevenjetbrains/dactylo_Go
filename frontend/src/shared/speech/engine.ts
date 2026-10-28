// Synthèse vocale derrière une interface (section 2.3) : Web Speech API
// aujourd'hui (voix françaises de Windows dans WebView2), voix Go (Piper,
// espeak-ng) demain sous Linux, sans changer le mode Dictée.

export interface Voice {
  id: string;
  name: string;
  lang: string;
  local: boolean;
}

export interface SpeakOptions {
  voiceId?: string;
  rate: number; // 0,5 à 2 (F-12)
  pitch: number; // 0 à 2
  volume: number; // 0 à 1
}

export interface SpeechEngine {
  available(): boolean;
  /** Voix françaises d'abord, puis les autres. */
  voices(): Promise<Voice[]>;
  /** Se résout à la fin de la lecture, ou dès qu'elle est annulée. */
  speak(text: string, opts: SpeakOptions): Promise<void>;
  cancel(): void;
}

const VOICES_TIMEOUT_MS = 1500;

export class WebSpeechEngine implements SpeechEngine {
  constructor(private readonly synth: SpeechSynthesis | undefined = globalThis.speechSynthesis) {}

  available(): boolean {
    return !!this.synth && typeof SpeechSynthesisUtterance !== 'undefined';
  }

  private async raw(): Promise<SpeechSynthesisVoice[]> {
    if (!this.synth) return [];
    const now = this.synth.getVoices();
    if (now.length > 0) return now;
    // Chromium charge les voix de façon asynchrone
    return new Promise((resolve) => {
      const done = () => resolve(this.synth!.getVoices());
      this.synth!.addEventListener('voiceschanged', done, { once: true });
      setTimeout(done, VOICES_TIMEOUT_MS);
    });
  }

  async voices(): Promise<Voice[]> {
    const all = (await this.raw()).map((v) => ({
      id: v.voiceURI,
      name: v.name,
      lang: v.lang,
      local: v.localService,
    }));
    const fr = (v: Voice) => (v.lang.toLowerCase().startsWith('fr') ? 0 : 1);
    return all.sort((a, b) => fr(a) - fr(b) || a.name.localeCompare(b.name));
  }

  async speak(text: string, opts: SpeakOptions): Promise<void> {
    if (!this.synth) return;
    const u = new SpeechSynthesisUtterance(text);
    u.lang = 'fr-FR';
    const voice = (await this.raw()).find((v) => v.voiceURI === opts.voiceId);
    if (voice) {
      try {
        u.voice = voice;
        u.lang = voice.lang;
      } catch {
        // voix refusée par le navigateur : on garde la voix française par défaut
      }
    }
    u.rate = opts.rate;
    u.pitch = opts.pitch;
    u.volume = opts.volume;
    return new Promise((resolve, reject) => {
      u.onend = () => resolve();
      u.onerror = (e) =>
        e.error === 'interrupted' || e.error === 'canceled'
          ? resolve()
          : reject(new Error(e.error));
      this.synth!.speak(u);
    });
  }

  cancel(): void {
    this.synth?.cancel();
  }
}

/** Moteur par défaut de l'application. */
export const speech: SpeechEngine = new WebSpeechEngine();
