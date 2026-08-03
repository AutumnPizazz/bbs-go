"use client"

import * as React from "react"
import { apiFetch, toFormData } from "@/lib/api/client"
import { toast } from "@/lib/toast"

/** 取块的行号：data-line 可能在块本身或其后代（如 pre > code）上。 */
function blockLine(el: Element): number {
  const direct = el.getAttribute("data-line")
  if (direct !== null) return Number(direct)
  const inner = el.querySelector("[data-line]")
  return inner ? Number(inner.getAttribute("data-line")) : Number.NaN
}

/**
 * attachLineAnchors 把 md-editor-rt 渲染产物（sourceHtml）中每个顶层块的
 * data-line（markdown 源码起始行号，0-based）按顺序复刻到后端渲染的
 * 预览 HTML（renderedHtml）上，供编辑区滚动做行级锚点同步。
 *
 * 两侧顶层块数量不一致（语法差异、HTML 块被安全过滤等）时放弃锚点，
 * 返回原始渲染结果，滚动同步会退化为比例模式。
 */
function attachLineAnchors(sourceHtml: string, renderedHtml: string): string {
  if (!sourceHtml || !renderedHtml) return renderedHtml
  const srcDoc = new DOMParser().parseFromString(sourceHtml, "text/html")
  const srcBlocks = [...srcDoc.body.children]
  if (srcBlocks.length === 0) return renderedHtml
  const lines = srcBlocks.map(blockLine)
  if (lines.some((n) => Number.isNaN(n))) return renderedHtml

  const doc = new DOMParser().parseFromString(renderedHtml, "text/html")
  const blocks = [...doc.body.children]
  if (blocks.length !== lines.length) return renderedHtml

  blocks.forEach((el, i) => {
    el.setAttribute("data-line", String(lines[i]))
  })
  return doc.body.innerHTML
}

/**
 * ServerRenderedPreview
 *
 * 将 markdown 文本通过后端 API 渲染为 HTML 后展示。
 * 与发布后视图使用完全相同的渲染管线（Lute + bluemonday + goquery 后处理）。
 * html prop 是 md-editor-rt 内部渲染产物（携带 data-line 行号锚点），
 * 仅用于提取行号映射，实际显示内容始终以后端渲染为准。
 */
export function ServerRenderedPreview({
  markdown,
  html,
  id,
  className: _className,
}: {
  markdown: string
  html?: string
  id?: string
  className?: string
}) {
  const [renderedHtml, setRenderedHtml] = React.useState("")
  const [loading, setLoading] = React.useState(false)
  const abortRef = React.useRef<AbortController | null>(null)
  const lastRenderedRef = React.useRef("")
  // 保存后端渲染的纯 HTML（不含行号锚点），html prop 更新时重新附着锚点
  const plainHtmlRef = React.useRef("")
  const timerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null)
  const rootRef = React.useRef<HTMLDivElement>(null)

  // ========== Markdown → HTML 渲染 ==========
  React.useEffect(() => {
    if (abortRef.current) abortRef.current.abort()
    if (!markdown || markdown.trim() === "") {
      plainHtmlRef.current = ""
      setRenderedHtml("")
      setLoading(false)
      return
    }
    // 内容已渲染完成：仅重新附着行号锚点（md-editor-rt 的 html prop 更新晚于后端渲染）
    if (markdown === lastRenderedRef.current) {
      if (plainHtmlRef.current) {
        setRenderedHtml(attachLineAnchors(html ?? "", plainHtmlRef.current))
      }
      return
    }

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
          lastRenderedRef.current = markdown
          plainHtmlRef.current = result.html
          // 渲染完成后把行号锚点附着到预览 DOM 上
          setRenderedHtml(attachLineAnchors(html ?? "", result.html))
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
  }, [markdown, html])

  // ========== 滚动 ==========
  // 不接管滚动——让 .md-editor-preview-wrapper 作为滚动容器，
  // markdown-editor 的自定义滚动跟随会按 data-line 锚点驱动它。
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
        opacity: loading && !renderedHtml ? 0.6 : 1,
        transition: "opacity 0.15s ease",
        minHeight: "100%",
      }}
    >
      {renderedHtml ? (
        <div
          className="bbs-content"
          style={{ padding: "10px 20px" }}
          dangerouslySetInnerHTML={{ __html: renderedHtml }}
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
