"use client"

import * as React from "react"
import { ActivityIcon, CheckCircle2Icon, PlayIcon, RefreshCwIcon, XCircleIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { useCurrentUser } from "@/components/app/app-provider"
import { adminGet, adminPostForm, type AdminRecord } from "@/lib/api/admin"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import { userHasPermission } from "@/lib/auth/roles"
import { useI18n } from "@/lib/i18n/provider"

type HealthState = {
  status?: string
  checkedAt?: number
  components?: Record<string, { status?: string; message?: string }>
}

type TaskState = {
  active?: boolean
  failed?: boolean
  tasks?: Record<string, AdminRecord>
}

function StatusIcon({ status }: { status?: string }) {
  return status === "ok" ? (
    <CheckCircle2Icon className="size-4 text-emerald-600" />
  ) : (
    <XCircleIcon className="size-4 text-destructive" />
  )
}

export default function DashboardHealthRoute() {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const [health, setHealth] = React.useState<HealthState | null>(null)
  const [tasks, setTasks] = React.useState<TaskState | null>(null)
  const [loading, setLoading] = React.useState(false)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const [nextHealth, nextTasks] = await Promise.all([
        adminGet<HealthState>("/api/admin/health"),
        adminGet<TaskState>("/api/admin/tasks/status"),
      ])
      setHealth(nextHealth)
      setTasks(nextTasks)
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void load()
  }, [load])

  async function startTask(endpoint: string) {
    await adminPostForm(endpoint, {})
    await load()
  }

  const canSearch = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_SEARCH_REINDEX)
  const canSitemap = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_SITEMAP_GENERATE)

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 md:p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">{t("dashboard.pages.health.title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t("dashboard.pages.health.description")}</p>
        </div>
        <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
          <RefreshCwIcon className={loading ? "size-4 animate-spin" : "size-4"} />
          {t("dashboard.actions.refresh")}
        </Button>
      </div>

      <section className="rounded-lg border bg-[var(--dashboard-panel)] p-4 shadow-xs">
        <div className="flex items-center gap-2">
          <ActivityIcon className="size-4 text-muted-foreground" />
          <h2 className="font-medium">{t("dashboard.pages.health.components")}</h2>
          <span className="ml-auto text-sm text-muted-foreground">{health?.status || t("common.loading")}</span>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {Object.entries(health?.components || {}).map(([name, component]) => (
            <div key={name} className="rounded-md border p-3">
              <div className="flex items-center gap-2 text-sm font-medium">
                <StatusIcon status={component.status} />
                {t(`dashboard.pages.health.componentNames.${name}`)}
              </div>
              {component.message ? <p className="mt-2 break-all text-xs text-muted-foreground">{component.message}</p> : null}
            </div>
          ))}
        </div>
      </section>

      <section className="rounded-lg border bg-[var(--dashboard-panel)] p-4 shadow-xs">
        <div className="flex items-center gap-2">
          <ActivityIcon className="size-4 text-muted-foreground" />
          <h2 className="font-medium">{t("dashboard.pages.health.tasks")}</h2>
          <span className="ml-auto text-sm text-muted-foreground">
            {tasks?.failed ? t("dashboard.pages.health.taskFailed") : tasks?.active ? t("dashboard.pages.health.taskRunning") : t("dashboard.pages.health.taskIdle")}
          </span>
        </div>
        <div className="mt-4 grid gap-3 md:grid-cols-2">
          <div className="rounded-md border p-3">
            <p className="text-sm font-medium">{t("dashboard.pages.health.taskNames.searchReindex")}</p>
            <p className="mt-1 text-xs text-muted-foreground">{String(tasks?.tasks?.searchReindex?.processed || 0)} / {String(tasks?.tasks?.searchReindex?.total || 0)}</p>
            {canSearch ? <Button className="mt-3" size="sm" onClick={() => void startTask("/api/admin/search/reindex")}><PlayIcon className="size-4" />{t("dashboard.pages.health.start")}</Button> : null}
          </div>
          <div className="rounded-md border p-3">
            <p className="text-sm font-medium">{t("dashboard.pages.health.taskNames.sitemap")}</p>
            <p className="mt-1 text-xs text-muted-foreground">{tasks?.tasks?.sitemap?.error ? String(tasks.tasks.sitemap.error) : t("dashboard.pages.health.taskReady")}</p>
            {canSitemap ? <Button className="mt-3" size="sm" onClick={() => void startTask("/api/admin/seo/sitemap/generate")}><PlayIcon className="size-4" />{t("dashboard.pages.health.start")}</Button> : null}
          </div>
        </div>
      </section>
    </div>
  )
}
