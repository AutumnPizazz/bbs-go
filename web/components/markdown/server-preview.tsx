"use client"

import * as React from "react"
import { apiFetch, toFormData } from "@/lib/api/client"
import { toast } from "@/lib/toast"

/**
 * ServerRenderedPreview
 *
 * 将 markdown 文本通过后端 API 渲染为 HTML 后展示。
 * 与发布后视图使用完全相同的渲染管线（Lute + bluemonday + goquery 后处理）。
 * 内置滚动同步：与编辑器左侧编辑区保持滚动位置一致。
 */
export function ServerRenderedPreview({
  markdown,
  id,
  className: _className,
}: {
  markdown: string
  id?: string
  className?: string
}) {
  const [html, setHtml] = React.useState("")
  const [loading, setLoading] = React.useState(false)
  const abortRef = React.useRef<AbortController | null>(null)
  const lastRenderedRef = React.useRef("")
  const timerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null)
  const rootRef = React.useRef<HTMLDivElement>(null)

  // ========== Markdown → HTML 渲染 ==========
  React.useEffect(() => {
    if (abortRef.current) abortRef.current.abort()
    if (!markdown || markdown.trim() === "") {
      setHtml("")
      setLoading(false)
      return
    }
    if (markdown === lastRenderedRef.current) return

    setLoading(true)
    if (timerRef.current) clearTimeout(timerRef.current)

    timerRef.current = setTimeout(async () => {
      const controller = new AbortController()
      abortRef.current = controller
      try {
        const result = await apiFetch<{ html: string }>(
          "/api/markdown/render",
          {
            method: "POST",
            body: toFormData({ content: markdown }),
            signal: controller.signal,
          }
        )
        if (controller.signal.aborted) return
        if (result?.html) {
          setHtml(result.html)
          lastRenderedRef.current = markdown
        }
      } catch (err: unknown) {
        if (controller.signal.aborted) return
        if (!(err instanceof TypeError && err.message === "Failed to fetch")) {
          console.warn("[ServerPreview] 渲染失败:", err)
          toast.error("Markdown 预览渲染失败")
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false)
      }
    }, 400)

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
      if (abortRef.current) abortRef.current.abort()
    }
  }, [markdown])

  // ========== 滚动同步 ==========
  // 不接管滚动——让 md-editor-rt 的 .md-editor-preview-wrapper 负责滚动，
  // 这样其内置的滚动同步逻辑才能正常工作。
  // 我们只需确保自己不设 overflow，内容高度自然撑开。
  React.useEffect(() => {
    const root = rootRef.current
    if (!root) return

    // 将父级 .md-editor-preview-wrapper 设为可滚动
    const wrapper = root.closest<HTMLElement>(".md-editor-preview-wrapper")
    if (wrapper) {
      wrapper.style.overflow = "auto"
      return () => {
        wrapper.style.overflow = ""
      }
    }
  }, [])

  return (
    <div
      ref={rootRef}
      id={id}
      style={{
        opacity: loading && !html ? 0.6 : 1,
        transition: "opacity 0.15s ease",
        minHeight: "100%",
      }}
    >
      {html ? (
        <div
          className="bbs-content"
          style={{ padding: "10px 20px" }}
          dangerouslySetInnerHTML={{ __html: html }}
        />
      ) : loading ? (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            padding: "2rem",
            color: "var(--muted-foreground)",
            fontSize: 14,
          }}
        >
          渲染中…
        </div>
      ) : null}
    </div>
  )
}
