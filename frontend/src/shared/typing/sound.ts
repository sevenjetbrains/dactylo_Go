// Sons de frappe (F-51) : courts bruits synthétisés par Web Audio, sans
// fichier ni réseau. Désactivés par défaut.

export type Tone = 'key' | 'error';

let ctx: AudioContext | null = null;

function audio(): AudioContext | null {
  if (ctx) return ctx;
  try {
    ctx = new AudioContext();
  } catch {
    return null; // pas de sortie audio (tests, WebView restreinte)
  }
  return ctx;
}

/** Joue le son d'une frappe juste (« key ») ou fausse (« error »). */
export function playTone(tone: Tone): void {
  const a = audio();
  if (!a) return;
  if (a.state === 'suspended') void a.resume();
  const t = a.currentTime;
  const osc = a.createOscillator();
  const gain = a.createGain();
  osc.type = tone === 'key' ? 'triangle' : 'square';
  osc.frequency.setValueAtTime(tone === 'key' ? 1800 : 220, t);
  const length = tone === 'key' ? 0.03 : 0.09;
  gain.gain.setValueAtTime(tone === 'key' ? 0.05 : 0.04, t);
  gain.gain.exponentialRampToValueAtTime(0.0001, t + length);
  osc.connect(gain).connect(a.destination);
  osc.start(t);
  osc.stop(t + length);
}
