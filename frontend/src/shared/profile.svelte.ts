import { Get, Rename } from '$wails/go/app/ProfileService';

// Profil affiché (barre latérale, accueil). Partagé entre le shell et les
// réglages ; les profils multiples arriveront avec F-50.
export const profile = $state({ name: '', keyboardLayout: 'azerty', loaded: false });

export async function loadProfile(): Promise<void> {
  try {
    const p = await Get();
    profile.name = p.name;
    profile.keyboardLayout = p.keyboardLayout;
    profile.loaded = true;
  } catch {
    // backend indisponible (tests) : on garde les valeurs par défaut
  }
}

export async function renameProfile(name: string): Promise<void> {
  const p = await Rename(name);
  profile.name = p.name;
}

/** Nom à afficher : « Moi » (nom par défaut) ne sert pas de prénom. */
export function displayName(name: string): string {
  return name === 'Moi' ? '' : name;
}
