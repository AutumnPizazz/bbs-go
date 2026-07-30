"use client"

import * as React from "react"

declare global {
  interface Window {
    __bbsCopyCode?: (btn: HTMLButtonElement) => void
  }
}

/**
 * CopyCodeEnhance — 注册全局复制按钮回调
 *
 * 代码块的包裹容器和语言标签已在 Go 端（misc_render.go）通过 goquery 直接注入 HTML，
 * 不再依赖 JS DOM 操作，避免了时序 / React 重渲染 / MutationObserver 漏检等问题。
 *
 * 本组件仅负责注册 window.__bbsCopyCode 全局函数，供服务端渲染的
 * onclick="__bbsCopyCode(this)" 调用。
 */
export function CopyCodeEnhance() {
  React.useEffect(() => {
    if (window.__bbsCopyCode) return // 避免重复注册

    window.__bbsCopyCode = (btn: HTMLButtonElement) => {
      // 找到同层 wrapper 内的 <code> 元素
      const wrapper = btn.closest(".code-block-wrapper")
      const code = wrapper?.querySelector("code")
      const text = code?.textContent || ""

      navigator.clipboard.writeText(text).then(
        () => {
          btn.classList.add("copied")
          btn.innerHTML =
            `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>` +
            `<span>已复制</span>`
          setTimeout(() => {
            btn.classList.remove("copied")
            btn.innerHTML =
              `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>` +
              `<span>复制</span>`
          }, 2000)
        },
        () => {
          const span = btn.querySelector("span")
          if (span) span.textContent = "失败"
          setTimeout(() => {
            const sp = btn.querySelector("span")
            if (sp) sp.textContent = "复制"
          }, 1500)
        }
      )
    }
  }, [])

  return null
}
