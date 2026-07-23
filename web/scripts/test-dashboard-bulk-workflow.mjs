import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { resolve } from "node:path"

const webRoot = resolve(import.meta.dirname, "..")
const dataDir = resolve(webRoot, "components/dashboard/data")
const pageSource = readFileSync(
  resolve(dataDir, "dashboard-data-page.tsx"),
  "utf8"
)
const dialogSource = readFileSync(
  resolve(dataDir, "dashboard-data-bulk-dialog.tsx"),
  "utf8"
)
const topicsSource = readFileSync(
  resolve(webRoot, "app/routes/dashboard.topics.tsx"),
  "utf8"
)

assert.match(
  pageSource,
  /const selectedRecords = Array\.from\(selectedIds\)/,
  "generic dashboard pages should retain selected records outside the current page"
)
assert.doesNotMatch(
  pageSource,
  /availableIds[\s\S]*setSelectedIds\(/,
  "generic dashboard pages should not prune selections when pagination changes"
)
assert.match(
  pageSource,
  /result\.failures[\s\S]*retryBulkFailures/,
  "generic dashboard pages should expose failed batch records for retry"
)
assert.match(
  pageSource,
  /recordsForIds\(current\.ids\)/,
  "bulk payloads should be rebuilt from all selected IDs, including prior pages"
)
assert.match(
  dialogSource,
  /dashboard\.bulk\.retryFailures/,
  "bulk result dialog should provide a retry action"
)
assert.match(
  dialogSource,
  /result\.failures\.map/,
  "bulk result dialog should list individual failed records"
)
assert.doesNotMatch(
  topicsSource,
  /availableIds[\s\S]*setSelectedIds\(/,
  "topic dashboard should retain selections across pages"
)
assert.match(
  topicsSource,
  /currentPageIds[\s\S]*next\.delete\(id\)[\s\S]*next\.add\(id\)/,
  "topic page selection should only toggle the current page"
)
assert.match(
  topicsSource,
  /bulkState\.result\.failures[\s\S]*openBulkAction/,
  "topic dashboard should retry only failed records"
)

console.log("dashboard bulk workflow tests passed")
