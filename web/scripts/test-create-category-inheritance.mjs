import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { resolve } from "node:path"

const webRoot = resolve(import.meta.dirname, "..")
const headerSource = readFileSync(
  resolve(webRoot, "components/layout/site-header.tsx"),
  "utf8"
)
const createRouteSource = readFileSync(
  resolve(webRoot, "app/routes/topic.create.tsx"),
  "utf8"
)

assert.match(
  headerSource,
  /pathname\.match\(\/\^\\\/topics\\\/category\\\/\(\\d\+\)/,
  "create actions should derive the selected category from a category page"
)
assert.match(
  headerSource,
  /searchParams\.get\("categoryId"\)/,
  "create actions should preserve an explicit category parameter"
)
assert.match(
  headerSource,
  /href:\s*topicCreateHref\(0, categoryId\)/,
  "discussion creation should inherit the selected category"
)
assert.match(
  headerSource,
  /href:\s*topicCreateHref\(2, categoryId\)/,
  "question creation should inherit the selected category"
)
assert.match(
  createRouteSource,
  /searchParams\.get\("categoryId"\)/,
  "the create route should initialize its category from the inherited parameter"
)

console.log("create category inheritance tests passed")
