import type { AppLocale } from "./types"

export function normalizeLocale(value: unknown): AppLocale {
  return value === "zh-CN" ? "zh-CN" : "en-US"
}

export function getBrowserLocale(fallback: AppLocale): AppLocale {
  if (typeof navigator === "undefined") return fallback

  const browserLocales = [navigator.language, ...(navigator.languages || [])]
  if (browserLocales.some((value) => value?.toLowerCase().startsWith("zh"))) {
    return "zh-CN"
  }
  if (browserLocales.some((value) => value?.toLowerCase().startsWith("en"))) {
    return "en-US"
  }

  return fallback
}
