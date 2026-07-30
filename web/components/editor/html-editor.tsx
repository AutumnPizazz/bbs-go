"use client"

import * as React from "react"
import { Maximize, Minimize, Eye, Code2, Columns2 } from "lucide-react"

import { useI18n } from "@/lib/i18n/provider"
import { cn } from "@/lib/utils"

type ViewMode = "split" | "edit" | "preview"

export function HtmlEditor({
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
  const { t } = useI18n()
  const containerRef = React.useRef<HTMLDivElement>(null)
  const textareaRef = React.useRef<HTMLTextAreaElement>(null)
  const [viewMode, setViewMode] = React.useState<ViewMode>("split")
  const [isFullscreen, setIsFullscreen] = React.useState(false)

  // 同步滚动
  const syncScrollRef = React.useRef(false)
  function handleEditorScroll() {
    if (syncScrollRef.current) return
    syncScrollRef.current = true
    const editor = textareaRef.current
    const preview = containerRef.current?.querySelector(
      ".html-preview-pane"
    ) as HTMLElement | null
    if (!editor || !preview) {
      syncScrollRef.current = false
      return
    }
    const ratio =
      editor.scrollTop / (editor.scrollHeight - editor.clientHeight || 1)
    preview.scrollTop =
      ratio * (preview.scrollHeight - preview.clientHeight || 1)
    requestAnimationFrame(() => {
      syncScrollRef.current = false
    })
  }

  React.useEffect(() => {
    const onFullscreenChange = () =>
      setIsFullscreen(Boolean(document.fullscreenElement))
    document.addEventListener("fullscreenchange", onFullscreenChange)
    return () => document.removeEventListener("fullscreenchange", onFullscreenChange)
  }, [])

  function toggleFullscreen() {
    const el = containerRef.current
    if (!el) return
    if (!document.fullscreenElement) {
      void el.requestFullscreen()
    } else {
      void document.exitFullscreen()
    }
  }

  const tu = (key: string) => t(`component.htmlEditor.toolbar.${key}`)

  return (
    <div
      ref={containerRef}
      className="html-editor-container"
      style={{
        display: "flex",
        flexDirection: "column",
        border: "1px solid var(--border)",
        borderRadius: "8px",
        overflow: "hidden",
        background: "var(--background)",
      }}
    >
      {/* 工具栏 */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "4px 8px",
          borderBottom: "1px solid var(--border)",
          background: "var(--muted)",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 4 }}>
          <span
            style={{
              fontSize: 11,
              color: "var(--muted-foreground)",
              marginRight: 8,
              fontWeight: 600,
              letterSpacing: "0.05em",
            }}
          >
            HTML
          </span>
          <ToolBtn
            title={tu("edit")}
            active={viewMode === "edit"}
            onClick={() => setViewMode("edit")}
          >
            <Code2 size={14} />
          </ToolBtn>
          <ToolBtn
            title={tu("split")}
            active={viewMode === "split"}
            onClick={() => setViewMode("split")}
          >
            <Columns2 size={14} />
          </ToolBtn>
          <ToolBtn
            title={tu("preview")}
            active={viewMode === "preview"}
            onClick={() => setViewMode("preview")}
          >
            <Eye size={14} />
          </ToolBtn>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 4 }}>
          <ToolBtn
            title={
              isFullscreen ? tu("exitFullscreen") : tu("fullscreen")
            }
            active={isFullscreen}
            onClick={toggleFullscreen}
          >
            {isFullscreen ? <Minimize size={14} /> : <Maximize size={14} />}
          </ToolBtn>
        </div>
      </div>

      {/* 编辑区 */}
      <div
        style={{
          display: "flex",
          flex: 1,
          minHeight: 0,
          height,
        }}
      >
        {/* 编辑面板 */}
        {(viewMode === "split" || viewMode === "edit") && (
          <div
            style={{
              flex: viewMode === "split" ? 1 : undefined,
              width: viewMode === "edit" ? "100%" : undefined,
              display: "flex",
              flexDirection: "column",
              borderRight:
                viewMode === "split" ? "1px solid var(--border)" : undefined,
            }}
          >
            <textarea
              ref={textareaRef}
              value={value}
              placeholder={placeholder}
              spellCheck={false}
              onScroll={handleEditorScroll}
              style={{
                flex: 1,
                display: "block",
                width: "100%",
                padding: "16px",
                border: "none",
                background: "var(--background)",
                color: "var(--foreground)",
                fontFamily:
                  '"JetBrains Mono", "Fira Code", "Cascadia Code", "SF Mono", "Menlo", "Consolas", monospace',
                fontSize: "13px",
                lineHeight: "1.7",
                resize: "none",
                outline: "none",
                tabSize: 2,
                overflow: "auto",
              }}
              onChange={(event) => onChange(event.currentTarget.value)}
            />
          </div>
        )}

        {/* 预览面板 */}
        {(viewMode === "split" || viewMode === "preview") && (
          <div
            className="html-preview-pane bbs-content"
            style={{
              flex: viewMode === "split" ? 1 : undefined,
              width: viewMode === "preview" ? "100%" : undefined,
              padding: "16px",
              overflow: "auto",
              wordBreak: "break-word",
            }}
            dangerouslySetInnerHTML={{ __html: value || "<p></p>" }}
          />
        )}
      </div>
    </div>
  )
}

function ToolBtn({
  title,
  active,
  children,
  onClick,
}: {
  title: string
  active?: boolean
  children: React.ReactNode
  onClick: () => void
}) {
  return (
    <button
      type="button"
      title={title}
      className={cn(
        "html-editor-tool-btn",
        active && "html-editor-tool-btn-active"
      )}
      style={{
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        width: 28,
        height: 28,
        border: "none",
        borderRadius: 4,
        background: active ? "var(--accent)" : "transparent",
        color: active ? "var(--accent-foreground)" : "var(--muted-foreground)",
        cursor: "pointer",
        transition: "all 0.15s ease",
      }}
      onClick={onClick}
    >
      {children}
    </button>
  )
}
