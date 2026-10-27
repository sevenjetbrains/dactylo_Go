import type { Component } from 'svelte';
import type { Locale } from '$i18n/index';

/** Dictionnaire de traductions imbriqué (format svelte-i18n). */
export interface Dictionary {
  [key: string]: string | Dictionary;
}

/**
 * Déclaration d'un module fonctionnel (feature slice).
 * Un module se décrit ici ; le shell se charge de la route, du menu,
 * des traductions et du panneau de réglages.
 */
export interface ModuleDefinition {
  /** Identifiant unique, en kebab-case (ex. "practice"). */
  id: string;
  /** Route hash, commençant par "/" (ex. "/practice"). */
  route: string;
  /** Clé i18n du libellé de menu. */
  labelKey: string;
  /** Tracé SVG (attribut d) de l'icône 24×24. */
  icon: string;
  /** Position dans le menu (croissante). */
  order: number;
  /** Page affichée pour la route. */
  page: Component;
  /** Masquer du menu (route accessible quand même). */
  hidden?: boolean;
  /** Panneau affiché dans l'écran Réglages. */
  settingsPanel?: Component;
  /** Traductions propres au module, fusionnées au démarrage. */
  messages?: Partial<Record<Locale, Dictionary>>;
}
