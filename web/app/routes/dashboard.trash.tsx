"use client"

import * as React from "react"
import { RotateCcwIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { DashboardPagination } from "@/components/dashboard/pagination-controls"
import { useCurrentUser } from "@/components/app/app-provider"
import { ErrorPage } from "@/components/common/error-page"
import {
  ConfirmDialog,
  type ConfirmDialogState,
} from "@/components/common/confirm-dialog"
import { adminList, adminPostForm, type AdminRecord } from "@/lib/api/admin"
import { formatDateTime } from "@/lib/format"
import { userHasPermission } from "@/lib/auth/roles"
import { useI18n } from "@/lib/i18n/provider"
import { msgError, msgSuccess } from "@/lib/toast"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

type TrashRecord = AdminRecord & {
  id?: number | string
  title?: string
  content?: string
  fileName?: string
  status?: number
  createTime?: number
  user?: { id?: number; nickname?: string; username?: string }
}

export default function DashboardTrashRoute() {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const [tab, setTab] = React.useState("topics")
  const [records, setRecords] = React.useState<TrashRecord[]>([])
  const [paging, setPaging] = React.useState<{ total: number; totalPages: number } | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [page, setPage] = React.useState(1)
  const [confirmState, setConfirmState] = React.useState<ConfirmDialogState>(null)

  const canManageTopics = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_TOPIC_DELETE)
  const canManageComments = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_COMMENT_DELETE)
  const canManageAttachments = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_ATTACHMENT_DELETE)

  if (!canManageTopics && !canManageComments && !canManageAttachments) {
    return <ErrorPage statusCode={403} />
  }

  const endpoints: Record<string, string> = {
    topics: "/api/admin/trash/topics",
    comments: "/api/admin/trash/comments",
    attachments: "/api/admin/trash/attachments",
  }

  const restoreEndpoints: Record<string, string> = {
    topics: "/api/admin/topic/undelete",
    comments: "/api/admin/trash/comments/restore",
    attachments: "/api/admin/trash/attachments/restore",
  }

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const endpoint = endpoints[tab]
      const params: Record<string, string | number> = { page, pageSize: 20 }
      const result = await adminList(endpoint, params)
      setRecords((result?.results as TrashRecord[]) || [])
      if (result?.page) {
        setPaging({
          total: result.page.total || 0,
          totalPages: Math.ceil((result.page.total || 0) / 20),
        })
      }
    } catch (error) {
      msgError(error instanceof Error ? error.message : t("dashboard.errors.loadFailed"))
    } finally {
      setLoading(false)
    }
  }, [tab, page, t])

  React.useEffect(() => {
    void load()
  }, [load])

  async function handleRestore(id: number | string, label: string) {
    try {
      const endpoint = restoreEndpoints[tab]
      await adminPostForm(endpoint, { id })
      msgSuccess(t("dashboard.pages.trash.restored", { label }))
      await load()
    } catch (error) {
      msgError(error instanceof Error ? error.message : t("dashboard.errors.actionFailed"))
    }
  }

  function renderRecord(record: TrashRecord) {
    const label = record.title || record.fileName || String(record.id)
    return (
      <tr key={String(record.id)} className="border-b last:border-0">
        <td className="max-w-[280px] truncate px-3 py-2 text-sm">
          {tab === "attachments" ? (
            <span className="font-mono text-xs">{record.fileName || record.id}</span>
          ) : (
            <span>{label || record.id}</span>
          )}
        </td>
        <td className="px-3 py-2 text-xs text-muted-foreground">
          {tab !== "attachments" && record.user ? (record.user.nickname || record.user.username || record.user.id) : "-"}
        </td>
        <td className="px-3 py-2 text-xs text-muted-foreground">
          {formatDateTime(record.createTime)}
        </td>
        <td className="px-3 py-2 text-right">
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              setConfirmState({
                title: t("dashboard.pages.trash.restoreConfirmTitle"),
                description: t("dashboard.pages.trash.restoreConfirmDesc", { label }),
                confirmText: t("dashboard.actions.restore"),
                onConfirm: () => void handleRestore(record.id!, label),
              })
            }
          >
            <RotateCcwIcon className="size-4" />
            {t("dashboard.actions.restore")}
          </Button>
        </td>
      </tr>
    )
  }

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 md:p-6">
      <div>
        <h1 className="text-xl font-semibold">{t("dashboard.pages.trash.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("dashboard.pages.trash.description")}</p>
      </div>

      <Tabs value={tab} onValueChange={(v) => { setTab(v); setPage(1); setRecords([]) }}>
        <div className="overflow-x-auto rounded-lg border bg-[var(--dashboard-panel)] p-2 shadow-xs">
          <TabsList className="h-auto min-w-max justify-start bg-transparent p-0">
            {canManageTopics ? (
              <TabsTrigger value="topics">{t("dashboard.pages.trash.tabTopics")}</TabsTrigger>
            ) : null}
            {canManageComments ? (
              <TabsTrigger value="comments">{t("dashboard.pages.trash.tabComments")}</TabsTrigger>
            ) : null}
            {canManageAttachments ? (
              <TabsTrigger value="attachments">{t("dashboard.pages.trash.tabAttachments")}</TabsTrigger>
            ) : null}
          </TabsList>
        </div>

        <div className="mt-4 rounded-lg border bg-[var(--dashboard-panel)] shadow-xs">
          {loading ? (
            <p className="p-4 text-sm text-muted-foreground">{t("dashboard.loading")}</p>
          ) : records.length === 0 ? (
            <p className="p-4 text-sm text-muted-foreground">{t("dashboard.pages.trash.empty")}</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[600px] text-sm">
                <thead className="border-b bg-muted/30 text-left text-xs text-muted-foreground">
                  <tr>
                    <th className="px-3 py-2">
                      {tab === "attachments" ? t("dashboard.pages.health.backup.file") : t("dashboard.fields.name")}
                    </th>
                    <th className="px-3 py-2">{t("dashboard.fields.user")}</th>
                    <th className="px-3 py-2">{t("dashboard.fields.createTime")}</th>
                    <th className="px-3 py-2 text-right">{t("dashboard.actions.title")}</th>
                  </tr>
                </thead>
                <tbody>{records.map(renderRecord)}</tbody>
              </table>
            </div>
          )}
        </div>

        {paging && paging.totalPages > 1 ? (
          <DashboardPagination
            page={page}
            pageCount={paging.totalPages}
            total={paging.total}
            limit={20}
            loading={loading}
            onPageChange={setPage}
          />
        ) : null}
      </Tabs>

      <ConfirmDialog
        state={confirmState}
        onOpenChange={(open) => { if (!open) setConfirmState(null) }}
      />
    </div>
  )
}
