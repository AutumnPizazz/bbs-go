import assert from "node:assert/strict"

import { findAnchorIndex, ratioMap } from "../lib/editor/scroll-anchor.ts"

// findAnchorIndex: 升序 data-line 数组中找最后一个 <= target 的索引

// 空数组
assert.equal(findAnchorIndex([], 5), -1)

// 常规命中
assert.equal(findAnchorIndex([0, 3, 5, 8, 12], 5), 2)
assert.equal(findAnchorIndex([0, 3, 5, 8, 12], 6), 2)
assert.equal(findAnchorIndex([0, 3, 5, 8, 12], 4), 1)

// 小于所有锚点
assert.equal(findAnchorIndex([3, 5, 8], 0), -1)
assert.equal(findAnchorIndex([3, 5, 8], 2), -1)

// 大于所有锚点
assert.equal(findAnchorIndex([0, 3, 5], 99), 2)

// 重复值（多个块起始于同一行，取最后一个）
assert.equal(findAnchorIndex([2, 2, 2, 7], 2), 2)

// 长数组（二分正确性）
const big = Array.from({ length: 10000 }, (_, i) => i * 3)
assert.equal(findAnchorIndex(big, 14999), 4999)
assert.equal(findAnchorIndex(big, 15000), 5000)
assert.equal(findAnchorIndex(big, 0), 0)
assert.equal(findAnchorIndex(big, -1), -1)

// ratioMap: 比例兜底映射，越界钳制
assert.equal(ratioMap(50, 100, 200), 100)
assert.equal(ratioMap(0, 100, 200), 0)
assert.equal(ratioMap(200, 100, 200), 200) // 超界钳制到最大值
assert.equal(ratioMap(-10, 100, 200), 0) // 负值钳制到 0
assert.equal(ratioMap(50, 0, 200), 0) // 无滚动范围
assert.equal(ratioMap(50, 100, 0), 0) // 目标不可滚

console.log("scroll-anchor tests passed")
