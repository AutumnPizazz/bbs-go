import assert from "node:assert/strict"
import { existsSync, readFileSync } from "node:fs"
import { resolve } from "node:path"

const webRoot = resolve(import.meta.dirname, "..")
const routesDir = resolve(webRoot, "app/routes")
const dashboardComponentsDir = resolve(webRoot, "components/dashboard")
const dashboardDataDir = resolve(dashboardComponentsDir, "data")

const dedicatedRoutes = {
  "dashboard.users.tsx": {
    expectedDefaultExport: "DashboardUsersRoute",
  },
  "dashboard.settings.tsx": {
    removedComponent: "admin-settings-page.tsx",
    expectedDefaultExport: "DashboardSettingsRoute",
    forbiddenImport: "admin-settings-page",
  },
  "dashboard.topics.tsx": {
    removedComponent: "admin-topic-feed-page.tsx",
    expectedDefaultExport: "DashboardTopicsRoute",
    forbiddenImport: "admin-topic-feed-page",
  },
  "dashboard.user-reports.tsx": {
    expectedDefaultExport: "DashboardUserReportsRoute",
  },
  "dashboard.categories.tsx": {
    expectedDefaultExport: "DashboardCategoriesRoute",
  },
  "dashboard.links.tsx": {
    expectedDefaultExport: "DashboardLinksRoute",
  },
  "dashboard.roles.tsx": {
    expectedDefaultExport: "DashboardRolesRoute",
  },
  "dashboard.content.tsx": {
    expectedDefaultExport: "DashboardContentRoute",
  },
}

for (const [routeFile, routeConfig] of Object.entries(dedicatedRoutes)) {
  const routePath = resolve(routesDir, routeFile)
  assert.equal(
    existsSync(routePath),
    true,
    `${routeFile} should be a dedicated dashboard route`
  )

  const routeSource = readFileSync(routePath, "utf8")
  assert.match(
    routeSource,
    new RegExp(`export default function ${routeConfig.expectedDefaultExport}`),
    `${routeFile} should own its default route component`
  )

  if (routeConfig.forbiddenImport) {
    assert.equal(
      routeSource.includes(routeConfig.forbiddenImport),
      false,
      `${routeFile} should not wrap the old dashboard page component`
    )
  }

  if (routeConfig.removedComponent) {
    assert.equal(
      existsSync(resolve(dashboardComponentsDir, routeConfig.removedComponent)),
      false,
      `${routeConfig.removedComponent} should be folded into its route module`
    )
  }
}

assert.equal(
  existsSync(resolve(routesDir, "dashboard.$.tsx")),
  false,
  "dashboard.$.tsx should be removed after splitting dashboard pages"
)

assert.equal(
  existsSync(resolve(routesDir, "dashboard.comments.tsx")),
  false,
  "dashboard.comments.tsx should be removed because comments are not managed in dashboard"
)

assert.equal(
  existsSync(resolve(dashboardDataDir, "dashboard-data-page-configs.tsx")),
  false,
  "dashboard-data-page-configs.tsx should be removed after moving configs into route modules"
)

for (const sourcePath of [
  resolve(dashboardComponentsDir, "app-sidebar.tsx"),
  resolve(dashboardComponentsDir, "dashboard-overview.tsx"),
]) {
  const source = readFileSync(sourcePath, "utf8")
  assert.equal(
    source.includes("/dashboard/comments"),
    false,
    `${sourcePath} should not link to dashboard comments`
  )
}

const categoriesRoute = readFileSync(resolve(routesDir, "dashboard.categories.tsx"), "utf8")
const usersRoute = readFileSync(resolve(routesDir, "dashboard.users.tsx"), "utf8")
const dashboardSelect = readFileSync(
  resolve(dashboardComponentsDir, "dashboard-select.tsx"),
  "utf8"
)
const dashboardDialog = readFileSync(
  resolve(dashboardComponentsDir, "dashboard-dialog.tsx"),
  "utf8"
)
const dashboardDataUtils = readFileSync(
  resolve(dashboardDataDir, "dashboard-data-utils.tsx"),
  "utf8"
)

assert.equal(
  /(?:name|key):\s*"type"/.test(categoriesRoute),
  false,
  "dashboard.categories.tsx should not expose the retired category type"
)

assert.match(
  dashboardSelect,
  /expandedValues/,
  "dashboard select should track expanded parent nodes"
)
assert.match(
  dashboardSelect,
  /aria-expanded/,
  "dashboard select should expose parent expansion state"
)
assert.match(
  dashboardSelect,
  /option\.ancestorValues/,
  "dashboard select should preserve ancestors while filtering"
)
assert.match(
  dashboardSelect,
  /selectedAncestorValues[\s\S]*?setExpandedValues/,
  "dashboard multi-select should expand ancestors of selected nodes"
)
assert.match(
  dashboardSelect,
  /selected=\{selectedValues\.includes\(optionValue\)\}[\s\S]*?depth=\{option\.depth\}/,
  "dashboard multi-select should render selectable hierarchical options"
)

assert.match(
  usersRoute,
  /name:\s*"categoryIds"[\s\S]*?type:\s*"multiselect"[\s\S]*?optionsEndpoint:\s*"\/api\/admin\/category\/options"[\s\S]*?visibleWhen:\s*\(values\)\s*=>\s*values\.contentAccessMode\s*===\s*"assigned_categories"/,
  "dashboard.users.tsx should use the hierarchical multi-select for accessible nodes"
)
assert.match(
  dashboardSelect,
  /PopoverPrimitive\.Portal\s+container=\{portalContainer \?\? undefined\}/,
  "dashboard selectors should portal inside the active dialog when available"
)
assert.match(
  dashboardDialog,
  /max-h-\[calc\(100vh-2rem\)\][\s\S]*?overflow-visible/,
  "dashboard dialogs should allow selector popovers to overflow the dialog frame"
)
assert.match(
  dashboardDataUtils,
  /childrenByParent[\s\S]*?record\.parentId[\s\S]*?flattenFlat/,
  "dashboard option normalization should rebuild flat parentId category data"
)

assert.equal(
  /name:\s*"parentId"[\s\S]*?type:\s*"tree-select"/.test(categoriesRoute),
  false,
  "dashboard.categories.tsx parent category form field should use DashboardSelect via type select"
)

assert.match(
  categoriesRoute,
  /name:\s*"categoryId"[\s\S]*?label:\s*dashboardData\.label\(t,\s*"category"\)[\s\S]*?optionsEndpoint:\s*"\/api\/admin\/category\/options"/,
  "dashboard.categories.tsx should filter by selected category id instead of parent id"
)

console.log("dashboard route structure tests passed")
