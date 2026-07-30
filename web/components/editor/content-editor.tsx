"use client"

import { MarkdownEditor } from "@/components/editor/markdown-editor"
import { RichTextEditor } from "@/components/editor/rich-text-editor"
import { HtmlEditor } from "@/components/editor/html-editor"

export type ContentEditorType = "html" | "html-source" | "markdown"

export function ContentEditor({
  contentType,
  value,
  placeholder,
  height,
  onChange,
}: {
  contentType: ContentEditorType
  value: string
  placeholder?: string
  height?: string
  onChange: (value: string) => void
}) {
  if (contentType === "markdown") {
    return (
      <MarkdownEditor
        value={value}
        placeholder={placeholder}
        height={height}
        onChange={onChange}
      />
    )
  }

  if (contentType === "html-source") {
    return (
      <HtmlEditor
        value={value}
        placeholder={placeholder}
        height={height}
        onChange={onChange}
      />
    )
  }

  return (
    <RichTextEditor
      value={value}
      placeholder={placeholder}
      height={height}
      onChange={onChange}
    />
  )
}
