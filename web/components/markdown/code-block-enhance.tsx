"use client"

import * as React from "react"

/**
 * CodeBlockEnhance — 代码块增强组件
 *
 * 在挂载后扫描 .bbs-content 容器中的 <pre><code> 代码块：
 * 1. 用 .code-block-wrapper 包裹 <pre>
 * 2. 提取 class="language-xxx" 作为语言标签
 * 3. 注入一键复制按钮
 * 4. 若 <pre> 有内联 background-color（Chroma 内联样式），交由 CSS !important 覆盖
 *
 * 使用方式：在根布局中挂载一次 <CodeBlockEnhance observe />。
 */

/** 从 <code class="language-go ..."> 中提取语言名称 */
function extractLanguage(codeEl: Element): string {
  const cls = codeEl.className || ""
  const match = cls.match(/language-(\w+)/)
  return match ? match[1] : ""
}

/** 语言标签的友好展示名 */
function languageLabel(lang: string): string {
  const overrides: Record<string, string> = {
    js: "JavaScript",
    ts: "TypeScript",
    tsx: "TSX",
    jsx: "JSX",
    html: "HTML",
    css: "CSS",
    scss: "SCSS",
    less: "Less",
    json: "JSON",
    yaml: "YAML",
    yml: "YAML",
    xml: "XML",
    md: "Markdown",
    markdown: "Markdown",
    sql: "SQL",
    sh: "Shell",
    bash: "Bash",
    zsh: "Zsh",
    powershell: "PowerShell",
    py: "Python",
    rb: "Ruby",
    rs: "Rust",
    go: "Go",
    java: "Java",
    kt: "Kotlin",
    swift: "Swift",
    c: "C",
    cpp: "C++",
    cs: "C#",
    php: "PHP",
    rust: "Rust",
    scala: "Scala",
    elixir: "Elixir",
    dart: "Dart",
    lua: "Lua",
    r: "R",
    dockerfile: "Dockerfile",
    nginx: "Nginx",
    graphql: "GraphQL",
    protobuf: "Protobuf",
    toml: "TOML",
    ini: "INI",
    diff: "Diff",
    makefile: "Makefile",
    cmake: "CMake",
  }
  const lower = lang.toLowerCase()
  return overrides[lower] || lang
}

function enhanceCodeBlock(preEl: HTMLPreElement) {
  // 避免重复包裹
  if (preEl.parentElement?.classList.contains("code-block-wrapper")) return

  const codeEl = preEl.querySelector<HTMLElement>("code")
  if (!codeEl) return

  const lang = extractLanguage(codeEl)

  // ---- 构建包裹容器 ----
  const wrapper = document.createElement("div")
  wrapper.className = "code-block-wrapper"

  // ---- 头部工具栏 ----
  const header = document.createElement("div")
  header.className = "code-block-header"

  // 语言标签
  const langSpan = document.createElement("span")
  langSpan.className = "code-block-lang"
  langSpan.textContent = lang ? languageLabel(lang) : ""

  // 复制按钮
  const copyBtn = document.createElement("button")
  copyBtn.type = "button"
  copyBtn.className = "code-block-copy-btn"
  copyBtn.innerHTML =
    `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>` +
    `<span>复制</span>`

  copyBtn.addEventListener("click", (e) => {
    e.stopPropagation()
    const text = codeEl.textContent || ""
    navigator.clipboard.writeText(text).then(
      () => {
        copyBtn.classList.add("copied")
        copyBtn.innerHTML =
          `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>` +
          `<span>已复制</span>`
        setTimeout(() => {
          copyBtn.classList.remove("copied")
          copyBtn.innerHTML =
            `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>` +
            `<span>复制</span>`
        }, 2000)
      },
      () => {
        copyBtn.textContent = "失败"
        setTimeout(() => {
          copyBtn.textContent = "复制"
        }, 1500)
      }
    )
  })

  header.appendChild(langSpan)
  header.appendChild(copyBtn)

  // ---- DOM 重组 ----
  preEl.parentNode?.insertBefore(wrapper, preEl)
  wrapper.appendChild(header)
  wrapper.appendChild(preEl)
}

function enhanceAllIn(root: Element) {
  const pres = root.querySelectorAll<HTMLPreElement>("pre")
  for (const pre of pres) {
    // 只处理 .bbs-content 内的代码块
    if (!pre.closest(".bbs-content")) continue
    enhanceCodeBlock(pre)
  }
}

export function CodeBlockEnhance({ observe = false }: { observe?: boolean }) {
  React.useEffect(() => {
    // 初始化
    const containers = document.querySelectorAll(".bbs-content")
    for (const c of containers) enhanceAllIn(c)

    if (!observe) return

    const observer = new MutationObserver((mutations) => {
      for (const m of mutations) {
        for (const node of m.addedNodes) {
          if (node instanceof Element) {
            if (node.classList.contains("bbs-content")) enhanceAllIn(node)
            const nested = node.querySelectorAll(".bbs-content")
            for (const n of nested) enhanceAllIn(n)
          }
        }
      }
    })
    observer.observe(document.body, { childList: true, subtree: true })
    return () => observer.disconnect()
  }, [observe])

  return null
}
