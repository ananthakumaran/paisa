<script lang="ts">
  // Toggle between zh-CN and en. We persist the explicit choice to
  // localStorage (read on next boot by both `+layout.ts` files) and then
  // force a full-page reload — several screens read translations once via
  // `onMount + get(t)('key')` so an in-place locale swap leaves stale
  // labels behind. Reload is the canonical answer (see issue #59).
  import { locale } from "$lib/i18n";
  import { _ as t } from "$lib/i18n";
  import { persistLocale, pickOppositeLocale, isChineseLocale } from "$lib/i18n/toggle";

  function toggle() {
    const next = pickOppositeLocale($locale);
    persistLocale(next);
    locale.set(next);
    location.reload();
  }

  // We render the "other" language label — clicking switches there.
  // Mirrors ThemeSwitcher's "show what you'll get" pattern.
  $: isZh = isChineseLocale($locale);
  $: label = isZh ? "EN" : "中";
  $: tooltip = isZh ? $t("nav.switch_to_english") : $t("nav.switch_to_chinese");
</script>

<button
  class="button is-ghost is-small invertable language-switcher"
  aria-label={tooltip}
  title={tooltip}
  on:click={toggle}
>
  <strong>{label}</strong>
</button>

<style>
  .language-switcher {
    border: none;
    background: transparent;
    padding: 0.25rem 0.5rem;
    line-height: 1;
    min-width: 2rem;
    /* Bulma's .is-small shrinks text to ~10px; we want the EN / 中 label to
       sit at roughly the same visual weight as the theme toggle icon. */
    font-size: 0.95rem;
  }

  .language-switcher strong {
    color: inherit;
    font-size: inherit;
  }

  .language-switcher:hover,
  .language-switcher:focus {
    background: transparent;
    color: var(--bulma-primary, currentColor);
  }
</style>
