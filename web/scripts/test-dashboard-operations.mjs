import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"
import { resolve } from "node:path"

const repoRoot = resolve(import.meta.dirname, "../..")
const settings = await readFile(resolve(repoRoot, "web/app/routes/dashboard.settings.tsx"), "utf8")
const overview = await readFile(resolve(repoRoot, "web/components/dashboard/dashboard-overview.tsx"), "utf8")
const hook = await readFile(resolve(repoRoot, "web/components/dashboard/data/use-dashboard-data-page.ts"), "utf8")
const permissions = await readFile(resolve(repoRoot, "web/lib/auth/permissions.generated.ts"), "utf8")

for (const endpoint of [
  "/api/admin/announcement/preview",
  "/api/admin/announcement/publish",
  "/api/admin/announcement/history",
]) {
  assert.match(settings, new RegExp(endpoint.replaceAll("/", "\\/")), `${endpoint} should be wired into settings`)
}
assert.match(settings, /announcementHistory/)
assert.match(settings, /dangerouslySetInnerHTML/)
assert.match(hook, /common\/preferences\/views/)
assert.match(overview, /overview\.trend/)
assert.match(overview, /overview\.breakdown/)
assert.match(permissions, /dashboard\.announcement\.publish/)

console.log("dashboard operations contract checks passed")
