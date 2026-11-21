import { get, readonly, writable } from "svelte/store";

export type ColorTheme = "light" | "dark";
export type ThemePreference = "system" | ColorTheme;

const storageKey = "theme-preference";
const mediaQuery = "(prefers-color-scheme: dark)";

function readPreference(): ThemePreference {
  try {
    const value = localStorage.getItem(storageKey);
    return value === "light" || value === "dark" ? value : "system";
  } catch {
    return "system";
  }
}

const preference = writable<ThemePreference>(readPreference());
export const themePreference = readonly(preference);

// Existing chart/color helpers always need a resolved light or dark theme.
export function getColorPreference(): ColorTheme {
  const value = get(preference);
  if (value !== "system") return value;
  return typeof window !== "undefined" && window.matchMedia(mediaQuery).matches ? "dark" : "light";
}

export function setColorPreference(value: ThemePreference) {
  try {
    localStorage.setItem(storageKey, value);
  } catch {
    // Keep the preference working for this session when storage is unavailable.
  }
  preference.set(value);
}

// Owned by the root layout, so page refreshes do not register extra listeners.
export function initializeTheme(onChange: (theme: ColorTheme) => void) {
  const media = window.matchMedia(mediaQuery);
  const apply = () => {
    const theme = getColorPreference();
    document.documentElement.dataset.theme = theme;
    document.documentElement.style.colorScheme = theme;
    onChange(theme);
  };
  const onSystemChange = () => {
    if (get(preference) === "system") apply();
  };
  const onStorageChange = (event: StorageEvent) => {
    if (event.key === storageKey || event.key === null) {
      preference.set(readPreference());
    }
  };

  const unsubscribe = preference.subscribe(apply);
  media.addEventListener("change", onSystemChange);
  window.addEventListener("storage", onStorageChange);

  return () => {
    unsubscribe();
    media.removeEventListener("change", onSystemChange);
    window.removeEventListener("storage", onStorageChange);
  };
}
