"use client"

import * as React from "react"
import { PlayIcon, RefreshCwIcon, SendIcon } from "lucide-react"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import { useCurrentUser } from "@/components/app/app-provider"
import { userHasPermission } from "@/lib/auth/roles"
import { adminList, adminPostForm, type AdminRecord } from "@/lib/api/admin"
import { msgError, msgSuccess } from "@/lib/toast"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { NativeSelect, NativeSelectOption } from "@/components/ui/native-select"

export default function DashboardMessagesRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "messages"),
    description: dashboardData.desc(t, "messages"),
    listEndpoint: "/api/admin/message/list",
    viewPermission: PERMISSIONS.DASHBOARD_MESSAGE_VIEW,
    detailEndpoint: (id) => `/api/admin/message/${id}`,
    createEndpoint: "/api/admin/message/create",
    createPermission: PERMISSIONS.DASHBOARD_MESSAGE_SEND,
    updateEndpoint: "/api/admin/message/update",
    updatePermission: PERMISSIONS.DASHBOARD_MESSAGE_SEND,
    deleteEndpoint: "/api/admin/message/delete",
    deletePermission: PERMISSIONS.DASHBOARD_MESSAGE_DELETE,
    deleteMode: "formIds",
    filters: [
      { name: "userId", label: dashboardData.label(t, "userId") },
      { name: "type", label: dashboardData.label(t, "type") },
      { name: "title", label: dashboardData.label(t, "title") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: [
          { label: dashboardData.label(t, "unread"), value: 0 },
          { label: dashboardData.label(t, "read"), value: 1 },
          { label: dashboardData.label(t, "deleted"), value: 2 },
        ],
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "title", label: dashboardData.label(t, "title"), className: "min-w-56" },
      { key: "type", label: dashboardData.label(t, "type") },
      { key: "status", label: dashboardData.label(t, "status") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    detailFields: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "fromId", label: dashboardData.label(t, "fromId") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "title", label: dashboardData.label(t, "title") },
      { key: "content", label: dashboardData.label(t, "content") },
      { key: "type", label: dashboardData.label(t, "type") },
      { key: "status", label: dashboardData.label(t, "status") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      {
        name: "userId",
        label: dashboardData.label(t, "userId"),
        required: true,
        type: "number",
        min: 1,
      },
      { name: "title", label: dashboardData.label(t, "title"), required: true },
      {
        name: "content",
        label: dashboardData.label(t, "content"),
        required: true,
        type: "textarea",
        colSpan: 2,
      },
    ],
    canDelete: (record) => record.status !== 2,
  }

  return (
    <>
      <DashboardDataPage config={config} />
      <MessageBroadcastPanel />
    </>
  )
}

type MessageTask = AdminRecord & {
  id?: number
  status?: number
  totalCount?: number
  sentCount?: number
  failedCount?: number
}

function MessageBroadcastPanel() {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const canBroadcast = userHasPermission(
    currentUser,
    PERMISSIONS.DASHBOARD_MESSAGE_BROADCAST
  )
  const canViewTasks = userHasPermission(
    currentUser,
    PERMISSIONS.DASHBOARD_MESSAGE_TASK_VIEW
  )
  const [title, setTitle] = React.useState("")
  const [content, setContent] = React.useState("")
  const [targetType, setTargetType] = React.useState("all")
  const [targetId, setTargetId] = React.useState("")
  const [recipientCount, setRecipientCount] = React.useState<number | null>(null)
  const [tasks, setTasks] = React.useState<MessageTask[]>([])
  const [loading, setLoading] = React.useState(false)

  const loadTasks = React.useCallback(async () => {
    if (!canViewTasks) return
    try {
      const page = await adminList<MessageTask>("/api/admin/message/task/list", {
        page: 1,
        pageSize: 20,
      })
      setTasks(page.results)
    } catch (err) {
      msgError(err instanceof Error ? err.message : t("dashboard.errors.loadFailed"))
    }
  }, [canViewTasks, t])

  React.useEffect(() => {
    void loadTasks()
  }, [loadTasks])

  async function preview() {
    try {
      const result = await adminPostForm<{ totalCount?: number }>(
        "/api/admin/message/task/preview",
        { targetType, ...(targetId ? { targetId: Number(targetId) } : {}) }
      )
      setRecipientCount(Number(result.totalCount || 0))
    } catch (err) {
      msgError(err instanceof Error ? err.message : t("dashboard.errors.actionFailed"))
    }
  }

  async function createDraft() {
    setLoading(true)
    try {
      await adminPostForm("/api/admin/message/task/create", {
        title,
        content,
        targetType,
        ...(targetId ? { targetId: Number(targetId) } : {}),
      })
      msgSuccess(t("dashboard.messages.messageTaskCreated"))
      setTitle("")
      setContent("")
      setRecipientCount(null)
      await loadTasks()
    } catch (err) {
      msgError(err instanceof Error ? err.message : t("dashboard.errors.actionFailed"))
    } finally {
      setLoading(false)
    }
  }

  async function startTask(id: number, retry: boolean) {
    setLoading(true)
    try {
      await adminPostForm(retry ? "/api/admin/message/task/retry" : "/api/admin/message/task/send", { id })
      msgSuccess(t(retry ? "dashboard.messages.messageTaskRetried" : "dashboard.messages.messageTaskStarted"))
      await loadTasks()
    } catch (err) {
      msgError(err instanceof Error ? err.message : t("dashboard.errors.actionFailed"))
    } finally {
      setLoading(false)
    }
  }

  if (!canBroadcast && !canViewTasks) return null

  return (
    <section className="flex flex-1 flex-col gap-4 border-t p-4 md:p-6">
      {canBroadcast ? (
        <div className="grid gap-3 rounded-lg border bg-[var(--dashboard-panel)] p-4 shadow-xs">
          <div>
            <h2 className="font-medium">{t("dashboard.messageTask.title")}</h2>
            <p className="mt-1 text-sm text-muted-foreground">{t("dashboard.messageTask.description")}</p>
          </div>
          <div className="grid gap-3 md:grid-cols-2">
            <Input value={title} onChange={(event) => setTitle(event.target.value)} placeholder={t("dashboard.fields.title")} />
            <NativeSelect value={targetType} onChange={(event) => setTargetType(event.target.value)} className="w-full">
              <NativeSelectOption value="all">{t("dashboard.messageTask.targetAll")}</NativeSelectOption>
              <NativeSelectOption value="role">{t("dashboard.messageTask.targetRole")}</NativeSelectOption>
              <NativeSelectOption value="category">{t("dashboard.messageTask.targetCategory")}</NativeSelectOption>
            </NativeSelect>
          </div>
          {targetType !== "all" ? <Input type="number" min={1} value={targetId} onChange={(event) => setTargetId(event.target.value)} placeholder={t("dashboard.messageTask.targetId")} /> : null}
          <Textarea value={content} onChange={(event) => setContent(event.target.value)} placeholder={t("dashboard.fields.content")} />
          <div className="flex flex-wrap items-center gap-2">
            <Button type="button" variant="outline" onClick={() => void preview()} disabled={loading}><RefreshCwIcon className="size-4" />{t("dashboard.messageTask.preview")}</Button>
            <Button type="button" onClick={() => void createDraft()} disabled={loading || !title.trim() || !content.trim()}><SendIcon className="size-4" />{t("dashboard.messageTask.saveDraft")}</Button>
            {recipientCount !== null ? <span className="text-sm text-muted-foreground">{t("dashboard.messageTask.recipientCount", { count: recipientCount })}</span> : null}
          </div>
        </div>
      ) : null}
      {canViewTasks ? (
        <div className="grid gap-3 rounded-lg border bg-[var(--dashboard-panel)] p-4 shadow-xs">
          <div className="flex items-center justify-between gap-3"><h2 className="font-medium">{t("dashboard.messageTask.records")}</h2><Button type="button" size="sm" variant="outline" onClick={() => void loadTasks()} disabled={loading}><RefreshCwIcon className="size-4" />{t("dashboard.actions.refresh")}</Button></div>
          {tasks.length === 0 ? <p className="text-sm text-muted-foreground">{t("dashboard.messageTask.empty")}</p> : <div className="grid gap-2">{tasks.map((task) => <div key={String(task.id)} className="flex flex-wrap items-center gap-3 rounded-md border p-3 text-sm"><span className="min-w-40 font-medium">{String(task.title || "-")}</span><span>{t("dashboard.messageTask.recipientCount", { count: Number(task.totalCount || 0) })}</span><span>{Number(task.sentCount || 0)} / {Number(task.failedCount || 0)}</span><span className="text-muted-foreground">{String(task.status ?? "-")}</span>{canBroadcast && Number(task.status) === 0 ? <Button size="sm" onClick={() => void startTask(Number(task.id), false)} disabled={loading}><PlayIcon className="size-4" />{t("dashboard.messageTask.send")}</Button> : null}{canBroadcast && Number(task.status) === 3 ? <Button size="sm" variant="outline" onClick={() => void startTask(Number(task.id), true)} disabled={loading}><RefreshCwIcon className="size-4" />{t("dashboard.messageTask.retry")}</Button> : null}</div>)}</div>}
        </div>
      ) : null}
    </section>
  )
}
