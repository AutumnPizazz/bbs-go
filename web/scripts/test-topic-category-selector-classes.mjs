import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { resolve } from "node:path"

const webRoot = resolve(import.meta.dirname, "..")
const componentSource = readFileSync(
  resolve(webRoot, "components/topic/category-selector.tsx"),
  "utf8"
)
const topicCss = readFileSync(resolve(webRoot, "styles/topic.css"), "utf8")
const createFormSource = readFileSync(
  resolve(webRoot, "components/topic/topic-create-form.tsx"),
  "utf8"
)
const editFormSource = readFileSync(
  resolve(webRoot, "components/topic/topic-edit-form.tsx"),
  "utf8"
)

const requiredClasses = [
  "topic-subcategories",
  "topic-subcategories-body",
  "topic-subcategories-header",
  "topic-subcategories-label",
  "topic-subcategories-list",
]

for (const className of requiredClasses) {
assert.equal(
  componentSource.includes(`"${className}"`),
    true,
    `CategoryQuickSelector should render ${className}`
  )
  assert.match(
    topicCss,
    new RegExp(`\\.publish-form\\s+\\.${className}\\b`),
    `${className} should be styled in topic.css`
  )
}

assert.match(componentSource, /expandedIds/, "selector should track collapsed parent state")
assert.match(componentSource, /aria-expanded/, "parent controls should expose expanded state")
assert.match(componentSource, /ancestorIds/, "selector should preserve parent paths")

for (const [name, source] of [
  ["create form", createFormSource],
  ["edit form", editFormSource],
]) {
  assert.match(
    source,
    /<CategorySelector\b/,
    `${name} should use the collapsible category selector`
  )
  assert.equal(
    source.includes("CategoryQuickSelector"),
    false,
    `${name} should not bypass the collapsible category selector`
  )
}

console.log("topic category selector class names are covered")
