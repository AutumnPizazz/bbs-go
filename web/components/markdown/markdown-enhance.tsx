"use client"

import * as React from "react"

/**
 * MarkdownEnhance — 客户端增强渲染组件
 *
 * 在挂载后扫描页面上的 .bbs-content 容器，自动为其中的：
 * 1. $$...$$ / $...$ 数学公式 → KaTeX 渲染
 * 2. ```mermaid 代码块 → Mermaid 图表
 *
 * 使用方式：在需要增强的页面或根布局中挂载一次即可，
 * 例如 <MarkdownEnhance observe /> 会通过 MutationObserver 自动处理后续 DOM 变化。
 */

interface KaTeXModule {
  renderToString: (
    latex: string,
    options?: { throwOnError?: boolean; displayMode?: boolean }
  ) => string
}

interface MermaidModule {
  run: (options?: { nodes?: ArrayLike<Element> }) => Promise<void>
}

function isKaTeXModule(mod: unknown): mod is KaTeXModule {
  return (
    typeof mod === "object" &&
    mod !== null &&
    typeof (mod as Record<string, unknown>).renderToString === "function"
  )
}

function isMermaidModule(mod: unknown): mod is MermaidModule {
  return (
    typeof mod === "object" &&
    mod !== null &&
    typeof (mod as Record<string, unknown>).run === "function"
  )
}

async function renderMathInElement(element: Element) {
  try {
    const katexMod = await import("katex")
    // KaTeX 提供 default export，也可能有命名导出
    const katex = isKaTeXModule(katexMod)
      ? katexMod
      : (katexMod as { default?: unknown }).default
    if (!katex || typeof (katex as KaTeXModule).renderToString !== "function") {
      console.warn("[MarkdownEnhance] KaTeX 模块加载失败，请检查 katex 版本")
      return
    }
    const render = (katex as KaTeXModule).renderToString

    // 先处理块级公式 $$...$$
    const blockPattern = /\$\$([\s\S]*?)\$\$/g
    const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT, {
      acceptNode(node) {
        // 跳过 code/pre 内的文本节点
        const parent = node.parentElement
        if (parent && (parent.tagName === "CODE" || parent.tagName === "PRE")) {
          return NodeFilter.FILTER_SKIP
        }
        return NodeFilter.FILTER_ACCEPT
      },
    })

    const textNodes: Text[] = []
    while (walker.nextNode()) {
      textNodes.push(walker.currentNode as Text)
    }

    for (const textNode of textNodes) {
      const text = textNode.nodeValue || ""
      if (!blockPattern.test(text) && !/\$[^$]+\$/.test(text)) {
        continue
      }
      blockPattern.lastIndex = 0

      const fragment = document.createDocumentFragment()
      let lastIndex = 0
      let match: RegExpExecArray | null

      // 处理块级公式
      const combinedPattern = /(\$\$[\s\S]*?\$\$|\$[^$\n]+?\$)/g
      while ((match = combinedPattern.exec(text)) !== null) {
        const before = text.slice(lastIndex, match.index)
        if (before) {
          fragment.appendChild(document.createTextNode(before))
        }

        const raw = match[0]
        const isBlock = raw.startsWith("$$")
        const latex = isBlock ? raw.slice(2, -2).trim() : raw.slice(1, -1).trim()

        try {
          const html = render(latex, {
            throwOnError: false,
            displayMode: isBlock,
          })
          const span = document.createElement("span")
          span.innerHTML = html
          if (isBlock) {
            span.className = "katex-block"
            span.style.cssText = "display:block;margin:1rem 0;overflow-x:auto"
          }
          fragment.appendChild(span)
        } catch {
          // 渲染失败，保留原始文本
          fragment.appendChild(document.createTextNode(raw))
        }

        lastIndex = match.index + raw.length
      }

      if (lastIndex < text.length) {
        fragment.appendChild(document.createTextNode(text.slice(lastIndex)))
      }

      if (lastIndex > 0 && textNode.parentNode) {
        textNode.parentNode.replaceChild(fragment, textNode)
      }
    }
  } catch (error) {
    console.warn("[MarkdownEnhance] KaTeX 加载失败:", error)
  }
}

async function renderMermaidInElement(element: Element) {
  const mermaidBlocks = element.querySelectorAll<HTMLElement>(
    "pre code.language-mermaid"
  )
  if (mermaidBlocks.length === 0) return

  try {
    const mermaidMod = await import("mermaid")
    const mermaid = isMermaidModule(mermaidMod)
      ? mermaidMod
      : (mermaidMod as { default?: unknown }).default
    if (!mermaid || typeof (mermaid as MermaidModule).run !== "function") {
      console.warn(
        "[MarkdownEnhance] Mermaid 模块加载失败，请检查 mermaid 版本"
      )
      return
    }

    const mermaidApi = mermaid as MermaidModule

    // 将每个 <pre><code class="language-mermaid"> 转换为 mermaid 渲染容器
    for (const codeEl of mermaidBlocks) {
      const preEl = codeEl.parentElement
      if (!preEl || preEl.tagName !== "PRE") continue

      const graphText = codeEl.textContent?.trim() || ""
      if (!graphText) continue

      // 生成唯一 ID
      const id = `mermaid-${Math.random().toString(36).slice(2, 9)}`

      // 替换 pre 为 div
      const container = document.createElement("div")
      container.className = "mermaid-container"
      container.style.cssText = "margin:1rem 0;overflow-x:auto"
      container.innerHTML = `<pre class="mermaid">${graphText}</pre>`
      preEl.replaceWith(container)
    }

    // 运行 Mermaid
    await mermaidApi.run({ nodes: element.querySelectorAll("pre.mermaid") })
  } catch (error) {
    console.warn("[MarkdownEnhance] Mermaid 加载失败:", error)
  }
}

async function enhanceElement(element: Element) {
  await renderMathInElement(element)
  await renderMermaidInElement(element)
}

export function MarkdownEnhance({ observe = false }: { observe?: boolean }) {
  React.useEffect(() => {
    // 初始化：增强页面上现有的所有 .bbs-content
    const enhanceAll = () => {
      const containers = document.querySelectorAll(".bbs-content")
      for (const container of containers) {
        enhanceElement(container)
      }
    }

    enhanceAll()

    if (!observe) return

    // 使用 MutationObserver 监听后续 DOM 变化
    const observer = new MutationObserver((mutations) => {
      for (const mutation of mutations) {
        for (const node of mutation.addedNodes) {
          if (node instanceof Element) {
            if (node.classList.contains("bbs-content")) {
              enhanceElement(node)
            }
            // 也检查新增节点内部的 .bbs-content
            const nested = node.querySelectorAll(".bbs-content")
            for (const n of nested) {
              enhanceElement(n)
            }
          }
        }
      }
    })

    observer.observe(document.body, { childList: true, subtree: true })
    return () => observer.disconnect()
  }, [observe])

  return null
}
