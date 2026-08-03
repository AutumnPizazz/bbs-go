/**
 * 编辑器滚动锚点工具。
 *
 * 预览 HTML 中每个块级元素带有 data-line 属性，值为该块在
 * markdown 源码中的起始行号（0-based，来自 md-editor-rt 渲染器的
 * markdown-it token.map）。编辑区滚动时，把 CodeMirror 的当前行号
 * 换算后在这里查找对应的预览块。
 */

/** 在升序排列的 data-line 行号数组中，二分查找最后一个 <= target 的索引。 */
export function findAnchorIndex(lines: number[], target: number): number {
  let lo = 0
  let hi = lines.length - 1
  let ans = -1
  while (lo <= hi) {
    const mid = (lo + hi) >> 1
    if (lines[mid] <= target) {
      ans = mid
      lo = mid + 1
    } else {
      hi = mid - 1
    }
  }
  return ans
}

/** 编辑区 scrollTop → 预览区 scrollTop 的比例兜底（锚点不可用时的退化逻辑）。 */
export function ratioMap(
  scrollTop: number,
  range: number,
  targetRange: number
): number {
  if (range <= 0 || targetRange <= 0) return 0
  const ratio = scrollTop / range
  return Math.min(Math.max(ratio, 0), 1) * targetRange
}
