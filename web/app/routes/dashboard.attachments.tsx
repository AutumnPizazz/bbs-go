"use client"

import * as React from "react"
import { PlayIcon } from "lucide-react"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import { useCurrentUser } from "@/components/app/app-provider"
import { userHasPermission } from "@/lib/auth/roles"
import { adminPostForm } from "@/lib/api/admin"
import { msgError, msgSuccess } from "@/lib/toast"
import { Button } from "@/components/ui/button"

export default function DashboardAttachmentsRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "attachments"),
    description: dashboardData.desc(t, "attachments"),
    listEndpoint: "/api/admin/attachment/list",
    viewPermission: PERMISSIONS.DASHBOARD_ATTACHMENT_VIEW,
    detailEndpoint: (id) => `/api/admin/attachment/${id}`,
    deleteEndpoint: "/api/admin/attachment/delete",
    deletePermission: PERMISSIONS.DASHBOARD_ATTACHMENT_DELETE,
    deleteMode: "formIds",
    filters: [
      { name: "fileName", label: dashboardData.label(t, "fileName") },
      { name: "fileType", label: dashboardData.label(t, "fileType") },
      { name: "userId", label: dashboardData.label(t, "userId") },
      { name: "topicId", label: dashboardData.label(t, "topicId") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: [
          { label: dashboardData.label(t, "active"), value: 0 },
          { label: dashboardData.label(t, "deleted"), value: 1 },
        ],
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id"), className: "min-w-44" },
      { key: "fileName", label: dashboardData.label(t, "fileName"), className: "min-w-48" },
      { key: "fileType", label: dashboardData.label(t, "fileType") },
      { key: "fileSize", label: dashboardData.label(t, "fileSize") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "topicId", label: dashboardData.label(t, "topicId") },
      { key: "downloadCount", label: dashboardData.label(t, "downloadCount") },
      { key: "status", label: dashboardData.label(t, "status") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    detailFields: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "fileName", label: dashboardData.label(t, "fileName") },
      { key: "fileUrl", label: dashboardData.label(t, "fileUrl") },
      { key: "fileType", label: dashboardData.label(t, "fileType") },
      { key: "fileSize", label: dashboardData.label(t, "fileSize") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "topicId", label: dashboardData.label(t, "topicId") },
      { key: "downloadCount", label: dashboardData.label(t, "downloadCount") },
      { key: "status", label: dashboardData.label(t, "status") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    canDelete: (record) => record.status !== 1,
  }

  return (
    <>
      <DashboardDataPage config={config} />
      <AttachmentCleanupPanel />
    </>
  )
}

function AttachmentCleanupPanel() {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const canCleanup = userHasPermission(
    currentUser,
    PERMISSIONS.DASHBOARD_ATTACHMENT_CLEANUP
  )
  const [running, setRunning] = React.useState(false)

  if (!canCleanup) return null

  async function startCleanup() {
    setRunning(true)
    try {
      const before = Date.now() - 7 * 24 * 60 * 60 * 1000
      await adminPostForm("/api/admin/attachment/cleanup-orphans", { before })
      msgSuccess(t("dashboard.messages.attachmentCleanupStarted"))
    } catch (err) {
      msgError(
        err instanceof Error ? err.message : t("dashboard.errors.actionFailed")
      )
    } finally {
      setRunning(false)
    }
  }

  return (
    <section className="border-t p-4 md:p-6">
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-[var(--dashboard-panel)] p-4 shadow-xs">
        <div>
          <h2 className="font-medium">
            {t("dashboard.pages.attachments.cleanupTitle")}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {t("dashboard.pages.attachments.cleanupDescription")}
          </p>
        </div>
        <Button type="button" onClick={() => void startCleanup()} disabled={running}>
          <PlayIcon className="size-4" />
          {t("dashboard.pages.health.start")}
        </Button>
      </div>
    </section>
  )
}
