import type { TFunction } from "@/lib/i18n"

export type EditorMode = "html" | "html-source" | "markdown"

export function getEditorModeOptions(t: TFunction) {
  return [
    { value: "html" as const, label: t("component.editorMode.visual") },
    {
      value: "html-source" as const,
      label: t("component.editorMode.htmlSource"),
    },
    { value: "markdown" as const, label: t("component.editorMode.markdown") },
  ]
}

export function getEditorSwitchTarget(contentType: EditorMode): EditorMode {
  if (contentType === "html") return "html-source"
  if (contentType === "html-source") return "markdown"
  return "html"
}

export function getEditorSwitchConfirmMessage(
  contentType: EditorMode,
  t: TFunction
) {
  const target = getEditorSwitchTarget(contentType)

  if (target === "markdown") {
    return t("component.editorMode.switchConfirm", {
      mode: t("component.editorMode.markdown"),
    })
  }
  if (target === "html-source") {
    return t("component.editorMode.switchConfirm", {
      mode: t("component.editorMode.htmlSource"),
    })
  }
  return t("component.editorMode.switchConfirm", {
    mode: t("component.editorMode.visual"),
  })
}
