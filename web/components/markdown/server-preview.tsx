"use client"

import * as React from "react"
import { apiFetch, toFormData } from "@/lib/api/client"
import { toast } from "@/lib/toast"

/**
 * ServerRenderedPreview
 *
 * 将 markdown 文本通过后端 API 渲染为 HTML 后展示。
 * 与发布后视图使用完全相同的渲染管线（Lute + bluemonday + goquery 后处理）。
 */
export function ServerRenderedPreview({
  markdown,
  id,
  className,
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

  React.useEffect(() => {
    // 取消上一次请求
    if (abortRef.current) {
      abortRef.current.abort()
    }

    // 空内容直接清空
    if (!markdown || markdown.trim() === "") {
      setHtml("")
      setLoading(false)
      return
    }

    // 内容没变化，跳过
    if (markdown === lastRenderedRef.current) return

    setLoading(true)

    // 防抖 400ms
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
        // 网络错误静默处理，不打断用户编辑
        if (err instanceof TypeError && err.message === "Failed to fetch") {
          // 网络不可用，继续使用旧预览
        } else {
          console.warn("[ServerPreview] 渲染失败:", err)
          toast.error("Markdown 预览渲染失败")
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false)
        }
      }
    }, 400)

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
      if (abortRef.current) abortRef.current.abort()
    }
  }, [markdown])

  return (
    <div
      id={id}
      className={className}
      style={{
        opacity: loading && !html ? 0.6 : 1,
        transition: "opacity 0.15s ease",
      }}
    >
      {html ? (
        <div
          className="bbs-content"
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
