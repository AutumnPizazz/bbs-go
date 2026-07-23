import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { resolve } from "node:path"
import { fileURLToPath } from "node:url"

const webDir = resolve(fileURLToPath(new URL("..", import.meta.url)))
const createForm = readFileSync(
  resolve(webDir, "components/topic/topic-create-form.tsx"),
  "utf8"
)
const editForm = readFileSync(
  resolve(webDir, "components/topic/topic-edit-form.tsx"),
  "utf8"
)
const categoryDashboard = readFileSync(
  resolve(webDir, "app/routes/dashboard.categories.tsx"),
  "utf8"
)
const categoryList = readFileSync(
  resolve(webDir, "components/topic/topic-dynamic-list-client-page.tsx"),
  "utf8"
)

for (const [name, source] of [
  ["create form", createForm],
  ["edit form", editForm],
]) {
  assert.equal(
    source.includes("categoryTypeMatches"),
    false,
    `${name} should offer every writable category for both topic types`
  )
  assert.match(
    source,
    /attachmentIds:\s*attachmentList\.map\(\(item\) => item\.id\)/,
    `${name} should submit attachments for questions as well as discussions`
  )
  assert.equal(
    /form\.type === 0 && effectiveAttachmentConfig\?\.enabled/.test(source),
    false,
    `${name} should not hide attachments from questions`
  )
}

assert.equal(
  /name:\s*"type"/.test(categoryDashboard),
  false,
  "category management should not expose a category type field"
)
assert.equal(
  /key:\s*"type"/.test(categoryDashboard),
  false,
  "category management should not expose a category type column"
)
assert.equal(
  /currentNode\??\.type|isQaNode|isNormalNode/.test(categoryList),
  false,
  "category lists should not infer their content mode from a retired category type"
)

console.log("unified category and question attachment tests passed")
