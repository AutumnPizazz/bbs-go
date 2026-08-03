"use client"

import * as React from "react"
import dynamic from "@/lib/router/dynamic"
import type { PreviewRendererProps, ToolbarNames } from "md-editor-rt"

import "md-editor-rt/lib/style.css"

import { EditorView } from "@codemirror/view"

import { useTheme } from "@/components/theme-provider"
import { uploadEditorImage } from "@/components/editor/upload"
import { ServerRenderedPreview } from "@/components/markdown/server-preview"
import { useI18n } from "@/lib/i18n/provider"
import { findAnchorIndex, ratioMap } from "@/lib/editor/scroll-anchor"

const MdEditor = dynamic(
  () => import("md-editor-rt").then((mod) => mod.MdEditor),
  { ssr: false }
)

const TOOLBARS = [
  "bold",
  "underline",
  "italic",
  "strikeThrough",
  "-",
  "title",
  "sub",
  "sup",
  "quote",
  "unorderedList",
  "orderedList",
  "task",
  "-",
  "codeRow",
  "code",
  "link",
  "image",
  "table",
  "-",
  "revoke",
  "next",
  "-",
  "preview",
  "catalog",
  "=",
  "fullscreen",
] satisfies ToolbarNames[]

// 用户手动滚动预览后，这段时间内编辑区滚动不再强拉预览
const PREVIEW_BROWSE_PAUSE_MS = 3000

/**
 * 双向行级锚点滚动同步，替代 md-editor-rt 内置的 scrollAuto 双向强绑定：
 *
 * - 编辑区滚动时，通过 CodeMirror view 计算视口顶部行号，预览滚动到
 *   对应 data-line 锚点块的位置（块内按编辑区偏移比例微调）。图片、
 *   代码块等高度差异大的块也能精确对应，不再依赖两侧总高度比例；
 * - 预览区滚动时反向同步编辑区：找到视口顶部对应的锚点块，换算回
 *   源码行并滚动 CodeMirror；
 * - 预览区始终可以自由滚动（不会被强行拉回），用户手动滚动预览后
 *   编辑区跟随，期间短暂暂停编辑区对预览的驱动，避免回环抖动；
 * - 预览异步渲染完成（MutationObserver）与图片加载完成（捕获阶段
 *   load 事件）后自动重新校准位置；
 * - 预览缺少 data-line 锚点（两侧渲染块数不一致）时退化为比例同步。
 */
function usePreviewFollowScroll(editorId: string) {
  const rafRef = React.useRef<number | null>(null)
  const editorRafRef = React.useRef<number | null>(null)
  // 正向（编辑区→预览）程序化设置预览 scrollTop 的目标值，用于区分用户滚动
  const programmaticTopRef = React.useRef(0)
  // 反向（预览→编辑区）程序化设置编辑区 scrollTop 的目标值，用于跳过回环
  const editorProgrammaticTopRef = React.useRef(-1)
  const browseUntilRef = React.useRef(0)
  const boundRef = React.useRef<{
    scroller: HTMLElement
    preview: HTMLElement
  } | null>(null)
  const cleanupRef = React.useRef<(() => void) | null>(null)

  const sync = React.useCallback(() => {
    const bound = boundRef.current
    if (!bound) return
    if (Date.now() < browseUntilRef.current) return
    const { scroller, preview } = bound

    // 反向同步程序化设置编辑区滚动触发的 scroll 事件，直接跳过避免回环
    if (
      editorProgrammaticTopRef.current >= 0 &&
      Math.abs(scroller.scrollTop - editorProgrammaticTopRef.current) <= 1
    ) {
      editorProgrammaticTopRef.current = -1
      return
    }

    const anchors = [...preview.querySelectorAll<HTMLElement>("[data-line]")]
    const view = EditorView.findFromDOM(scroller)
    let top = 0
    if (anchors.length > 0 && view) {
      top = anchorTop(view, scroller, preview, anchors)
    } else {
      // 锚点不可用：退化为两侧总高度比例映射
      const editorRange = scroller.scrollHeight - scroller.clientHeight
      const previewRange = preview.scrollHeight - preview.clientHeight
      top = ratioMap(scroller.scrollTop, editorRange, previewRange)
    }
    programmaticTopRef.current = top
    preview.scrollTop = top
  }, [])

  // 反向同步：预览滚动 → 编辑区滚动到对应源码行
  const syncEditorFromPreview = React.useCallback(() => {
    const bound = boundRef.current
    if (!bound) return
    const { scroller, preview } = bound
    const anchors = [...preview.querySelectorAll<HTMLElement>("[data-line]")]
    const view = EditorView.findFromDOM(scroller)
    if (anchors.length === 0 || !view) return

    // 找预览视口顶部对应的锚点块（内容位置 <= scrollTop 的最后一个块）
    const previewRect = preview.getBoundingClientRect()
    let idx = -1
    for (let i = 0; i < anchors.length; i++) {
      const elTop =
        anchors[i].getBoundingClientRect().top -
        previewRect.top +
        preview.scrollTop
      if (elTop <= preview.scrollTop + 2) {
        idx = i
      } else {
        break
      }
    }
    if (idx < 0) idx = 0

    const doc = view.state.doc
    const startLine = Math.min(Number(anchors[idx].dataset.line), doc.lines - 1)
    const endLine =
      idx + 1 < anchors.length
        ? Math.min(Number(anchors[idx + 1].dataset.line), doc.lines - 1)
        : doc.lines - 1
    const startPos = doc.line(startLine + 1).from
    const endPos = doc.line(Math.min(endLine + 1, doc.lines)).from
    const startTop = view.lineBlockAt(startPos).top
    const endTop = view.lineBlockAt(endPos).top

    // 块内偏移比例：预览上锚点块到下一块的区间内
    const anchorEl = anchors[idx]
    const anchorTopInPreview =
      anchorEl.getBoundingClientRect().top - previewRect.top + preview.scrollTop
    const nextTopInPreview =
      idx + 1 < anchors.length
        ? anchors[idx + 1].getBoundingClientRect().top -
          previewRect.top +
          preview.scrollTop
        : preview.scrollHeight
    const span = Math.max(nextTopInPreview - anchorTopInPreview, 1)
    const k = (preview.scrollTop - anchorTopInPreview) / span

    let target = startTop + (endTop - startTop) * k
    const maxTop = Math.max(0, scroller.scrollHeight - scroller.clientHeight)
    target = Math.min(Math.max(target, 0), maxTop)

    editorProgrammaticTopRef.current = target
    scroller.scrollTop = target
  }, [])

  React.useEffect(() => {
    const bind = () => {
      if (cleanupRef.current) {
        cleanupRef.current()
        cleanupRef.current = null
        boundRef.current = null
      }
      const editor = document.getElementById(editorId)
      if (!editor) return
      const scroller = editor.querySelector<HTMLElement>(".cm-scroller")
      const preview = document.getElementById(`${editorId}-preview-wrapper`)
      if (!scroller || !preview) return
      boundRef.current = { scroller, preview }

      const scheduleSync = () => {
        if (rafRef.current !== null) cancelAnimationFrame(rafRef.current)
        rafRef.current = requestAnimationFrame(sync)
      }
      const recalibrate = () => {
        // 渲染完成 / 图片加载后预览高度变化，重新映射编辑区当前位置
        if (rafRef.current !== null) cancelAnimationFrame(rafRef.current)
        sync()
      }
      const onPreviewScroll = () => {
        // 程序化同步触发的 scroll 事件忽略，其余视为用户滚动预览：
        // 暂停正向驱动片刻，并反向同步编辑区
        if (Math.abs(preview.scrollTop - programmaticTopRef.current) > 1) {
          browseUntilRef.current = Date.now() + PREVIEW_BROWSE_PAUSE_MS
          if (editorRafRef.current !== null) {
            cancelAnimationFrame(editorRafRef.current)
          }
          editorRafRef.current = requestAnimationFrame(() => {
            editorRafRef.current = null
            syncEditorFromPreview()
          })
        }
      }

      scroller.addEventListener("scroll", scheduleSync, { passive: true })
      preview.addEventListener("scroll", onPreviewScroll, { passive: true })
      // img 的 load 事件不冒泡，需要在捕获阶段监听预览容器
      preview.addEventListener("load", recalibrate, true)
      const observer = new MutationObserver(recalibrate)
      observer.observe(preview, { childList: true, subtree: true })

      cleanupRef.current = () => {
        scroller.removeEventListener("scroll", scheduleSync)
        preview.removeEventListener("scroll", onPreviewScroll)
        preview.removeEventListener("load", recalibrate, true)
        observer.disconnect()
        if (rafRef.current !== null) {
          cancelAnimationFrame(rafRef.current)
          rafRef.current = null
        }
        if (editorRafRef.current !== null) {
          cancelAnimationFrame(editorRafRef.current)
          editorRafRef.current = null
        }
      }
    }

    // MdEditor 是动态加载的子组件，且工具栏可切换预览重建 DOM；
    // 监听整个文档，发现目标元素出现或替换时（重新）绑定。
    const ensureBound = () => {
      const editor = document.getElementById(editorId)
      if (!editor) return
      const scroller = editor.querySelector<HTMLElement>(".cm-scroller")
      const preview = document.getElementById(`${editorId}-preview-wrapper`)
      if (!scroller || !preview) return
      const bound = boundRef.current
      if (!bound || bound.scroller !== scroller || bound.preview !== preview) {
        bind()
      }
    }
    ensureBound()
    const rebindObserver = new MutationObserver(ensureBound)
    rebindObserver.observe(document.body, { childList: true, subtree: true })

    return () => {
      rebindObserver.disconnect()
      if (cleanupRef.current) {
        cleanupRef.current()
        cleanupRef.current = null
        boundRef.current = null
      }
    }
  }, [editorId, sync, syncEditorFromPreview])
}

/**
 * 根据编辑区视口顶部行号计算预览应滚动到的位置：
 * 找到该行所属的锚点块（data-line <= 行号的最后一个块），
 * 在"块起始行 → 下一块起始行"的编辑区区间内按偏移比例映射到
 * "锚点块 → 下一块"的预览区间。
 */
function anchorTop(
  view: EditorView,
  scroller: HTMLElement,
  preview: HTMLElement,
  anchors: HTMLElement[]
): number {
  const topBlock = view.lineBlockAtHeight(scroller.scrollTop)
  // CodeMirror 行号为 1-based，data-line 为 0-based
  const lineNo = view.state.doc.lineAt(topBlock.from).number - 1
  const lines = anchors.map((el) => Number(el.dataset.line))
  const idx = findAnchorIndex(lines, lineNo)
  if (idx < 0) {
    const editorRange = scroller.scrollHeight - scroller.clientHeight
    const previewRange = preview.scrollHeight - preview.clientHeight
    return ratioMap(scroller.scrollTop, editorRange, previewRange)
  }

  const startEl = anchors[idx]
  const startLine = lines[idx]
  // 编辑区上锚点行到下一个锚点行的偏移区间
  const doc = view.state.doc
  const startPos = doc.line(Math.min(startLine + 1, doc.lines)).from
  const endPos =
    idx + 1 < lines.length
      ? doc.line(Math.min(lines[idx + 1] + 1, doc.lines)).from
      : doc.length
  const startOffset = view.lineBlockAt(startPos).top
  const endOffset = view.lineBlockAt(endPos).top
  const editorSpan = endOffset - startOffset

  // 预览上锚点块到下一个块的偏移区间（用实测位置，图片加载后自动修正）
  const previewRect = preview.getBoundingClientRect()
  const startRect = startEl.getBoundingClientRect()
  const anchorTopInPreview = startRect.top - previewRect.top + preview.scrollTop
  const nextEl = idx + 1 < anchors.length ? anchors[idx + 1] : null
  const previewSpan = nextEl
    ? nextEl.getBoundingClientRect().top - startRect.top
    : preview.scrollHeight - anchorTopInPreview

  const k = editorSpan > 0 ? (scroller.scrollTop - startOffset) / editorSpan : 0
  const maxTop = Math.max(0, preview.scrollHeight - preview.clientHeight)
  return Math.min(Math.max(anchorTopInPreview + previewSpan * k, 0), maxTop)
}

export function MarkdownEditor({
  value,
  placeholder,
  height = "400px",
  onChange,
}: {
  value: string
  placeholder?: string
  height?: string
  onChange: (value: string) => void
}) {
  const { resolvedTheme } = useTheme()
  const { locale } = useI18n()

  // editorId 用于定位编辑区与预览区 DOM，必须去掉 useId 生成的冒号
  const reactId = React.useId()
  const editorId = reactId.replace(/[^a-zA-Z0-9_-]/g, "")
  usePreviewFollowScroll(editorId)

  const currentValueRef = React.useRef(value)
  React.useEffect(() => {
    currentValueRef.current = value
  }, [value])

  const ServerPreview = React.useCallback(
    ({ id, className, html }: PreviewRendererProps) => (
      <ServerRenderedPreview
        markdown={currentValueRef.current}
        html={html}
        id={id}
        className={className}
      />
    ),
    []
  )

  async function uploadImg(files: File[], callback: (urls: string[]) => void) {
    const urls = await Promise.all(files.map((file) => uploadEditorImage(file)))
    callback(urls)
  }

  return (
    <MdEditor
      editorId={editorId}
      scrollAuto={false}
      modelValue={value}
      theme={resolvedTheme === "dark" ? "dark" : "light"}
      toolbars={TOOLBARS}
      style={{ height }}
      placeholder={placeholder}
      preview
      language={locale}
      footers={[]}
      previewComponent={ServerPreview}
      onChange={onChange}
      onUploadImg={uploadImg}
    />
  )
}
