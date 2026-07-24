"use client"

import * as React from "react"
import {
  ActivityIcon,
  CheckCircle2Icon,
  CircleHelpIcon,
  DatabaseBackupIcon,
  DownloadIcon,
  PlayIcon,
  RefreshCwIcon,
  RotateCcwIcon,
  SaveIcon,
  Trash2Icon,
  XCircleIcon,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import {
  ConfirmDialog,
  type ConfirmDialogState,
} from "@/components/common/confirm-dialog"
import { useCurrentUser } from "@/components/app/app-provider"
import { adminGet, adminPostForm, type AdminRecord } from "@/lib/api/admin"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import { userHasPermission } from "@/lib/auth/roles"
import { useI18n } from "@/lib/i18n/provider"
import { msgError, msgSuccess } from "@/lib/toast"

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

type BackupConfig = {
  enabled?: boolean
  schedule?: string
  retention?: number
  directory?: string
}

type BackupStatus = {
  running?: boolean
  lastSuccessAt?: number
  lastFailureAt?: number
  lastError?: string
  retention?: number
  directory?: string
  restore?: {
    running?: boolean
    lastSuccessAt?: number
    lastFailureAt?: number
    lastError?: string
    lastBackupId?: number
  }
}

type BackupRecord = {
  id: number
  fileName?: string
  databaseType?: string
  triggerType?: string
  status?: number
  size?: number
  checksum?: string
  error?: string
  startedAt?: number
  finishedAt?: number
}

type BackupState = {
  config?: BackupConfig
  status?: BackupStatus
  results?: BackupRecord[]
}

const byteCountFormatter = new Intl.NumberFormat("en-US")

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
  const [backup, setBackup] = React.useState<BackupState | null>(null)
  const [backupConfig, setBackupConfig] = React.useState<BackupConfig | null>(null)
  const [confirmState, setConfirmState] = React.useState<ConfirmDialogState>(null)
  const [loading, setLoading] = React.useState(false)

  const canBackupView = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_BACKUP_VIEW)
  const canBackupCreate = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_BACKUP_CREATE)
  const canBackupDownload = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_BACKUP_DOWNLOAD)
  const canBackupDelete = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_BACKUP_DELETE)
  const canBackupConfig = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_BACKUP_CONFIG)
  const canBackupRestore = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_BACKUP_RESTORE)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const [nextHealth, nextTasks, nextBackup, nextBackupList] = await Promise.all([
        adminGet<HealthState>("/api/admin/health"),
        adminGet<TaskState>("/api/admin/tasks/status"),
        canBackupView
          ? adminGet<BackupState>("/api/admin/backup/config")
          : Promise.resolve(null),
        canBackupView
          ? adminGet<BackupState>("/api/admin/backup/list?limit=50")
          : Promise.resolve(null),
      ])
      setHealth(nextHealth)
      setTasks(nextTasks)
      if (nextBackup || nextBackupList) {
        setBackup({
          config: nextBackup?.config,
          status: nextBackupList?.status || nextBackup?.status,
          results: nextBackupList?.results || [],
        })
        if (!backupConfig && nextBackup?.config) {
          setBackupConfig(nextBackup.config)
        }
      }
    } finally {
      setLoading(false)
    }
  }, [backupConfig, canBackupView])

  React.useEffect(() => {
    void load()
  }, [load])

  React.useEffect(() => {
    if (!backup?.status?.restore?.running) {
      return
    }

    const timer = window.setInterval(() => {
      void load()
    }, 2000)

    return () => window.clearInterval(timer)
  }, [backup?.status?.restore?.running, load])

  async function startTask(endpoint: string) {
    try {
      await adminPostForm(endpoint, {})
      await load()
    } catch (error) {
      msgError(error instanceof Error ? error.message : t("dashboard.pages.health.actionFailed"))
    }
  }

  async function createBackup() {
    try {
      await adminPostForm("/api/admin/backup/create", { confirm: "CREATE_BACKUP" })
      msgSuccess(t("dashboard.pages.health.backup.created"))
      await load()
    } catch (error) {
      msgError(error instanceof Error ? error.message : t("dashboard.pages.health.actionFailed"))
    }
  }

  async function saveBackupConfig() {
    if (!backupConfig) return
    try {
      await adminPostForm("/api/admin/backup/config", {
        enabled: backupConfig.enabled || false,
        schedule: backupConfig.schedule || "",
        retention: backupConfig.retention || 0,
        directory: backupConfig.directory || "",
        confirm: "CREATE_BACKUP",
      })
      msgSuccess(t("dashboard.pages.health.backup.saved"))
      await load()
    } catch (error) {
      msgError(error instanceof Error ? error.message : t("dashboard.pages.health.actionFailed"))
    }
  }

  async function deleteBackup(id: number) {
    try {
      await adminPostForm("/api/admin/backup/delete", { id, confirm: "DELETE_BACKUP" })
      msgSuccess(t("dashboard.pages.health.backup.deleted"))
      await load()
    } catch (error) {
      msgError(error instanceof Error ? error.message : t("dashboard.pages.health.actionFailed"))
    }
  }

  async function restoreBackup(id: number) {
    try {
      await adminPostForm("/api/admin/backup/restore", { id, confirm: "RESTORE_BACKUP" })
      msgSuccess(t("dashboard.pages.health.backup.restoreStarted"))
      await load()
    } catch (error) {
      msgError(error instanceof Error ? error.message : t("dashboard.pages.health.actionFailed"))
    }
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
          <div className="rounded-md border p-3">
            <p className="text-sm font-medium">{t("dashboard.pages.health.taskNames.databaseBackup")}</p>
            <p className="mt-1 text-xs text-muted-foreground">
              {tasks?.tasks?.databaseBackup?.lastError
                ? String(tasks.tasks.databaseBackup.lastError)
                : tasks?.tasks?.databaseBackup?.running
                  ? t("dashboard.pages.health.taskRunning")
                  : t("dashboard.pages.health.taskReady")}
            </p>
            {canBackupCreate ? (
              <Button
                className="mt-3"
                size="sm"
                onClick={() =>
                  setConfirmState({
                    title: t("dashboard.pages.health.backup.createTitle"),
                    description: t("dashboard.pages.health.backup.createDescription"),
                    confirmText: t("dashboard.pages.health.backup.create"),
                    onConfirm: () => void createBackup(),
                  })
                }
                disabled={Boolean(backup?.status?.running || backup?.status?.restore?.running)}
              >
                <DatabaseBackupIcon className="size-4" />
                {t("dashboard.pages.health.backup.create")}
              </Button>
            ) : null}
          </div>
        </div>
      </section>

      {canBackupView ? (
        <section className="rounded-lg border bg-[var(--dashboard-panel)] p-4 shadow-xs">
          <div className="flex items-center gap-2">
            <DatabaseBackupIcon className="size-4 text-muted-foreground" />
            <h2 className="font-medium">{t("dashboard.pages.health.backup.title")}</h2>
            <span className="ml-auto text-sm text-muted-foreground">
              {backup?.status?.restore?.running
                ? t("dashboard.pages.health.backup.restoreRunning")
                : backup?.status?.restore?.lastError || backup?.status?.lastError || t("dashboard.pages.health.backup.noErrors")}
            </span>
          </div>

          {canBackupConfig && backupConfig ? (
            <div className="mt-4 grid gap-3 rounded-md border p-3 md:grid-cols-4">
              <label className="flex items-center gap-2 text-sm md:col-span-4">
                <input
                  type="checkbox"
                  checked={Boolean(backupConfig.enabled)}
                  onChange={(event) => setBackupConfig({ ...backupConfig, enabled: event.target.checked })}
                />
                {t("dashboard.pages.health.backup.enabled")}
              </label>
              <div className="space-y-2">
                <div className="flex items-center gap-1">
                  <Label htmlFor="backup-schedule">{t("dashboard.pages.health.backup.schedule")}</Label>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button
                        type="button"
                        aria-label={t("dashboard.pages.health.backup.scheduleHelpLabel")}
                        className="text-muted-foreground transition-colors hover:text-foreground"
                      >
                        <CircleHelpIcon className="size-4" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent className="max-w-sm whitespace-pre-line">
                      {t("dashboard.pages.health.backup.scheduleHelp")}
                    </TooltipContent>
                  </Tooltip>
                </div>
                <Input
                  id="backup-schedule"
                  value={backupConfig.schedule || ""}
                  onChange={(event) => setBackupConfig({ ...backupConfig, schedule: event.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="backup-retention">{t("dashboard.pages.health.backup.retention")}</Label>
                <Input
                  id="backup-retention"
                  type="number"
                  min={1}
                  max={100}
                  value={backupConfig.retention || 1}
                  onChange={(event) => setBackupConfig({ ...backupConfig, retention: Number(event.target.value) })}
                />
              </div>
              <div className="space-y-2 md:col-span-2">
                <div className="flex items-center gap-1">
                  <Label htmlFor="backup-directory">{t("dashboard.pages.health.backup.directory")}</Label>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button
                        type="button"
                        aria-label={t("dashboard.pages.health.backup.directoryHelpLabel")}
                        className="text-muted-foreground transition-colors hover:text-foreground"
                      >
                        <CircleHelpIcon className="size-4" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent className="max-w-sm whitespace-pre-line">
                      {t("dashboard.pages.health.backup.directoryHelp")}
                    </TooltipContent>
                  </Tooltip>
                </div>
                <Input
                  id="backup-directory"
                  value={backupConfig.directory || ""}
                  onChange={(event) => setBackupConfig({ ...backupConfig, directory: event.target.value })}
                />
              </div>
              <div className="md:col-span-4">
                <Button
                  size="sm"
                  onClick={() =>
                    setConfirmState({
                      title: t("dashboard.pages.health.backup.saveTitle"),
                      description: t("dashboard.pages.health.backup.saveDescription"),
                      confirmText: t("dashboard.actions.save"),
                      onConfirm: () => void saveBackupConfig(),
                    })
                  }
                >
                  <SaveIcon className="size-4" />
                  {t("dashboard.actions.save")}
                </Button>
              </div>
            </div>
          ) : null}

          <div className="mt-4 overflow-x-auto rounded-md border">
            <table className="w-full min-w-[760px] text-sm">
              <thead className="border-b bg-muted/30 text-left text-xs text-muted-foreground">
                <tr>
                  <th className="px-3 py-2">{t("dashboard.pages.health.backup.file")}</th>
                  <th className="px-3 py-2">{t("dashboard.pages.health.backup.status")}</th>
                  <th className="px-3 py-2">{t("dashboard.pages.health.backup.size")}</th>
                  <th className="px-3 py-2">{t("dashboard.pages.health.backup.finishedAt")}</th>
                  <th className="px-3 py-2 text-right">{t("dashboard.pages.health.backup.actions")}</th>
                </tr>
              </thead>
              <tbody>
                {(backup?.results || []).map((record) => (
                  <tr key={record.id} className="border-b last:border-0">
                    <td className="max-w-[280px] truncate px-3 py-2 font-mono text-xs">{record.fileName}</td>
                    <td className="px-3 py-2">
                      {record.status === 2
                        ? t("dashboard.pages.health.backup.success")
                        : record.status === 1
                          ? t("dashboard.pages.health.backup.running")
                          : t("dashboard.pages.health.backup.failed")}
                      {record.error ? <p className="mt-1 max-w-[260px] text-xs text-destructive">{record.error}</p> : null}
                    </td>
                    <td className="px-3 py-2">{byteCountFormatter.format(record.size || 0)}</td>
                    <td className="px-3 py-2 text-xs text-muted-foreground">{record.finishedAt || "-"}</td>
                    <td className="px-3 py-2 text-right">
                      <div className="flex justify-end gap-2">
                        {canBackupDownload && record.status === 2 ? (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() =>
                              setConfirmState({
                                title: t("dashboard.pages.health.backup.downloadTitle"),
                                description: t("dashboard.pages.health.backup.downloadDescription"),
                                confirmText: t("dashboard.pages.health.backup.download"),
                                onConfirm: () => {
                                  window.location.assign(`/api/admin/backup/download/${record.id}?confirm=DOWNLOAD_BACKUP`)
                                },
                              })
                            }
                          >
                            <DownloadIcon className="size-4" />
                            {t("dashboard.pages.health.backup.download")}
                          </Button>
                        ) : null}
                        {canBackupRestore && record.status === 2 && record.databaseType?.toLowerCase() === "mysql" ? (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() =>
                              setConfirmState({
                                title: t("dashboard.pages.health.backup.restoreTitle"),
                                description: t("dashboard.pages.health.backup.restoreDescription"),
                                confirmText: t("dashboard.pages.health.backup.restore"),
                                onConfirm: () => void restoreBackup(record.id),
                              })
                            }
                            disabled={Boolean(backup?.status?.restore?.running)}
                          >
                            <RotateCcwIcon className="size-4" />
                            {t("dashboard.pages.health.backup.restore")}
                          </Button>
                        ) : null}
                        {canBackupDelete && record.status !== 1 ? (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() =>
                              setConfirmState({
                                title: t("dashboard.pages.health.backup.deleteTitle"),
                                description: t("dashboard.pages.health.backup.deleteDescription"),
                                confirmText: t("dashboard.pages.health.backup.delete"),
                                onConfirm: () => void deleteBackup(record.id),
                              })
                            }
                          >
                            <Trash2Icon className="size-4" />
                            {t("dashboard.pages.health.backup.delete")}
                          </Button>
                        ) : null}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {backup?.results?.length === 0 ? <p className="p-4 text-sm text-muted-foreground">{t("dashboard.pages.health.backup.empty")}</p> : null}
          </div>
        </section>
      ) : null}

      <ConfirmDialog
        state={confirmState}
        onOpenChange={(open) => {
          if (!open) setConfirmState(null)
        }}
      />
    </div>
  )
}
