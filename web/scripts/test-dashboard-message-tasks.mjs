import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"
import { resolve } from "node:path"

const webRoot = resolve(import.meta.dirname, "..")
const root = resolve(webRoot, "..")
const route = await readFile(resolve(webRoot, "app/routes/dashboard.messages.tsx"), "utf8")
const router = await readFile(resolve(root, "internal/server/router.go"), "utf8")
const registry = await readFile(resolve(root, "internal/permissions/admin_permission_registry.go"), "utf8")
const permissions = await readFile(resolve(webRoot, "lib/auth/permissions.generated.ts"), "utf8")

assert.match(route, /\/api\/admin\/message\/task\/preview/)
assert.match(route, /\/api\/admin\/message\/task\/create/)
assert.match(route, /\/api\/admin\/message\/task\/retry/)
assert.match(route, /DASHBOARD_MESSAGE_BROADCAST/)
assert.match(router, /messageGroup\.POST\("\/task\/preview"/)
assert.match(router, /messageGroup\.GET\("\/task\/:id"/)
assert.match(registry, /PermissionMessageBroadcast/)
assert.match(registry, /PermissionMessageTaskView/)
assert.match(permissions, /DASHBOARD_MESSAGE_TASK_VIEW/)

console.log("dashboard message task tests passed")
