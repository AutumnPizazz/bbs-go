"use client"

import * as React from "react"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { HtmlImagePreview } from "@/components/common/image-preview"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { adminGet, adminPostForm, type AdminRecord } from "@/lib/api/admin"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import { useCurrentUser } from "@/components/app/app-provider"
import { userHasPermission } from "@/lib/auth/roles"
import { msgError, msgSuccess } from "@/lib/toast"
import {
  NativeSelect,
  NativeSelectOption,
} from "@/components/ui/native-select"

type ReportDetailRecord = AdminRecord & {
  target?: Record<string, unknown>
  relatedReports?: AdminRecord[]
  relatedReportCount?: number
  pendingRelatedReportCount?: number
}

export default function DashboardUserReportsRoute() {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const [processingReport, setProcessingReport] =
    React.useState<ReportDetailRecord | null>(null)
  const [processingLoading, setProcessingLoading] = React.useState(false)
  const [submittingStatus, setSubmittingStatus] = React.useState<number | null>(
    null
  )
  const [submittingAction, setSubmittingAction] = React.useState<string | null>(
    null
  )
  const [reportScope, setReportScope] = React.useState<"current" | "object">(
    "current"
  )
  const [reloadKey, setReloadKey] = React.useState(0)
  const canProcess = userHasPermission(
    currentUser,
    PERMISSIONS.DASHBOARD_USER_REPORT_PROCESS
  )
  const canBatchProcess = userHasPermission(
    currentUser,
    PERMISSIONS.DASHBOARD_USER_REPORT_BATCH_PROCESS
  )

  async function openProcessReport(record: AdminRecord) {
    setProcessingLoading(true)
    try {
      const detail = await adminGet<ReportDetailRecord>(
        `/api/admin/user-report/${String(record.id)}`
      )
      setReportScope("current")
      setProcessingReport(detail)
    } catch (err) {
      msgError(
        err instanceof Error ? err.message : t("dashboard.errors.loadFailed")
      )
    } finally {
      setProcessingLoading(false)
    }
  }

  async function submitReportStatus(processStatus: 1 | 2) {
    if (!processingReport?.id) return
    setSubmittingStatus(processStatus)
    try {
      await adminPostForm("/api/admin/user-report/process", {
        id: processingReport.id as number,
        processStatus,
        scope: reportScope,
      })
      msgSuccess(
        processStatus === 1
          ? t("dashboard.messages.reportProcessed")
          : t("dashboard.messages.reportIgnored")
      )
      setProcessingReport(null)
      setReloadKey((current) => current + 1)
    } catch (err) {
      msgError(
        err instanceof Error ? err.message : t("dashboard.errors.actionFailed")
      )
    } finally {
      setSubmittingStatus(null)
    }
  }

  async function submitReportAction(action: string, days?: number) {
    if (!processingReport?.id || !canProcess) return
    setSubmittingAction(action)
    try {
      await adminPostForm("/api/admin/user-report/action", {
        id: processingReport.id as number,
        action,
        scope: reportScope,
        ...(days ? { days } : {}),
      })
      msgSuccess(t("dashboard.messages.reportProcessed"))
      setProcessingReport(null)
      setReloadKey((current) => current + 1)
    } catch (err) {
      msgError(
        err instanceof Error ? err.message : t("dashboard.errors.actionFailed")
      )
    } finally {
      setSubmittingAction(null)
    }
  }

  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "userReports"),
    description: dashboardData.desc(t, "userReports"),
    listEndpoint: "/api/admin/user-report/list",
    viewPermission: PERMISSIONS.DASHBOARD_USER_REPORT_VIEW,
    refreshKey: reloadKey,
    filters: [
      { name: "dataId", label: dashboardData.label(t, "dataId") },
      {
        name: "dataType",
        label: dashboardData.label(t, "dataType"),
        type: "select",
        options: dashboardData.reportDataTypeOptionsFor(t),
      },
      {
        name: "processStatus",
        label: dashboardData.label(t, "processStatus"),
        type: "select",
        options: dashboardData.reportProcessStatusOptionsFor(t),
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      {
        key: "dataType",
        label: dashboardData.label(t, "dataType"),
        render: (record) =>
          dashboardData.reportDataTypeCell(t, record.dataType),
      },
      { key: "dataId", label: dashboardData.label(t, "dataId") },
      {
        key: "target",
        label: dashboardData.label(t, "reportTarget"),
        render: (record) => dashboardData.reportTargetCell(t, record),
      },
      { key: "userId", label: dashboardData.label(t, "userId") },
      {
        key: "reason",
        label: dashboardData.label(t, "reason"),
        className: "min-w-72",
      },
      {
        key: "processStatus",
        label: dashboardData.label(t, "processStatus"),
        render: (record) =>
          dashboardData.reportProcessStatusCell(t, record.processStatus),
      },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    bulkActions: canBatchProcess
      ? [
          {
            label: t("dashboard.reportActions.batchProcess"),
            endpoint: "/api/admin/user-report/batch",
            previewEndpoint: "/api/admin/user-report/batch/preview",
            permission: PERMISSIONS.DASHBOARD_USER_REPORT_BATCH_PROCESS,
            payload: () => ({ action: "process" }),
            successMessage: t("dashboard.messages.reportProcessed"),
          },
          {
            label: t("dashboard.reportActions.batchIgnore"),
            endpoint: "/api/admin/user-report/batch",
            previewEndpoint: "/api/admin/user-report/batch/preview",
            permission: PERMISSIONS.DASHBOARD_USER_REPORT_BATCH_PROCESS,
            payload: () => ({ action: "ignore" }),
            successMessage: t("dashboard.messages.reportIgnored"),
          },
        ]
      : [],
    renderRowActions: (record) =>
      canProcess && Number(record.processStatus || 0) === 0 ? (
        <Button
          type="button"
          size="sm"
          variant="outline"
          disabled={processingLoading}
          onClick={() => void openProcessReport(record)}
        >
          {t("dashboard.reportActions.process")}
        </Button>
      ) : null,
  }

  return (
    <>
      <DashboardDataPage config={config} />
      <ReportProcessDialog
        record={processingReport}
        submittingStatus={submittingStatus}
        submittingAction={submittingAction}
        reportScope={reportScope}
        canProcess={canProcess}
        onClose={() => setProcessingReport(null)}
        onScopeChange={setReportScope}
        onSubmitStatus={(status) => void submitReportStatus(status)}
        onSubmitAction={(action, days) => void submitReportAction(action, days)}
      />
    </>
  )
}

function ReportProcessDialog({
  record,
  submittingStatus,
  submittingAction,
  reportScope,
  canProcess,
  onClose,
  onScopeChange,
  onSubmitStatus,
  onSubmitAction,
}: {
  record: ReportDetailRecord | null
  submittingStatus: number | null
  submittingAction: string | null
  reportScope: "current" | "object"
  canProcess: boolean
  onClose: () => void
  onScopeChange: (scope: "current" | "object") => void
  onSubmitStatus: (status: 1 | 2) => void
  onSubmitAction: (action: string, days?: number) => void
}) {
  const { t } = useI18n()
  const target =
    record?.target && !record.target.missing ? record.target : undefined
  const dataType = String(record?.dataType || "")
  const targetUrl =
    target?.url && dataType !== "comment" ? String(target.url) : undefined

  return (
    <Dialog open={Boolean(record)} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t("dashboard.reportProcess.title")}</DialogTitle>
          <DialogDescription>
            {record
              ? `${dashboardData.reportDataTypeCell(t, record.dataType)} #${String(record.dataId || "-")}`
              : ""}
          </DialogDescription>
        </DialogHeader>

        {record ? (
          <div className="grid gap-4">
            <div className="grid gap-1.5">
              <div className="text-xs font-medium text-muted-foreground">
                {dashboardData.label(t, "reason")}
              </div>
              <div className="rounded-md border bg-muted/20 px-3 py-2 text-sm whitespace-pre-wrap">
                {String(record.reason || "-")}
              </div>
            </div>

            {Number(record.relatedReportCount || 0) > 0 ? (
              <div className="grid gap-1.5">
                <div className="text-xs font-medium text-muted-foreground">
                  {t("dashboard.reportProcess.scope")}
                </div>
                <NativeSelect
                  value={reportScope}
                  onChange={(event) =>
                    onScopeChange(event.target.value as "current" | "object")
                  }
                  disabled={Boolean(submittingStatus || submittingAction)}
                  className="w-full"
                >
                  <NativeSelectOption value="current">
                    {t("dashboard.reportProcess.scopeCurrent")}
                  </NativeSelectOption>
                  <NativeSelectOption value="object">
                    {t("dashboard.reportProcess.scopeObject", {
                      count: Number(record.pendingRelatedReportCount || 0),
                    })}
                  </NativeSelectOption>
                </NativeSelect>
              </div>
            ) : null}

            <div className="grid gap-1.5">
              <div className="text-xs font-medium text-muted-foreground">
                {dashboardData.label(t, "reportTarget")}
              </div>
              <div className="grid max-h-[52vh] gap-3 overflow-auto rounded-md border bg-muted/20 p-3 text-sm">
                {target ? (
                  <ReportTargetPreview target={target} />
                ) : (
                  <div className="text-muted-foreground">
                    {t("dashboard.reportProcess.targetUnavailable")}
                  </div>
                )}
              </div>
            </div>
          </div>
        ) : null}

        <DialogFooter>
          {targetUrl ? (
            <Button type="button" variant="outline" asChild>
              <a href={targetUrl} target="_blank" rel="noreferrer">
                {t("dashboard.reportActions.openTarget")}
              </a>
            </Button>
          ) : null}
          {canProcess && record?.dataType === "topic" && Number(target?.status) === 0 ? (
            <Button
              type="button"
              variant="destructive"
              disabled={Boolean(submittingStatus || submittingAction)}
              onClick={() => onSubmitAction("delete")}
            >
              {t("dashboard.actions.delete")}
            </Button>
          ) : null}
          {canProcess && record?.dataType === "topic" && Number(target?.status) === 1 ? (
            <Button
              type="button"
              variant="outline"
              disabled={Boolean(submittingStatus || submittingAction)}
              onClick={() => onSubmitAction("restore")}
            >
              {t("dashboard.actions.undelete")}
            </Button>
          ) : null}
          {canProcess && record?.dataType === "comment" && Number(target?.status) === 0 ? (
            <Button
              type="button"
              variant="destructive"
              disabled={Boolean(submittingStatus || submittingAction)}
              onClick={() => onSubmitAction("delete")}
            >
              {t("dashboard.actions.delete")}
            </Button>
          ) : null}
          {canProcess && record?.dataType === "user" ? (
            <Button
              type="button"
              variant="outline"
              disabled={Boolean(submittingStatus || submittingAction)}
              onClick={() => onSubmitAction("forbid", 7)}
            >
              {t("component.userCenterSidebar.forbidden7Days")}
            </Button>
          ) : null}
          <Button
            type="button"
            variant="outline"
            disabled={Boolean(submittingStatus || submittingAction)}
            onClick={() => onSubmitStatus(2)}
          >
            {submittingStatus === 2
              ? t("dashboard.actions.save")
              : t("dashboard.reportActions.markIgnored")}
          </Button>
          <Button
            type="button"
            disabled={Boolean(submittingStatus || submittingAction)}
            onClick={() => onSubmitStatus(1)}
          >
            {submittingStatus === 1
              ? t("dashboard.actions.save")
              : t("dashboard.reportActions.markProcessed")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function ReportTargetPreview({ target }: { target: Record<string, unknown> }) {
  const title = target.title || target.nickname || target.username
  const content = target.content || target.description || target.summary
  const contentType = String(target.contentType || "")
  const meta = [
    target.username ? `@${String(target.username)}` : "",
    target.userId ? `userId: ${String(target.userId)}` : "",
    target.entityType ? `entityType: ${String(target.entityType)}` : "",
    target.entityId ? `entityId: ${String(target.entityId)}` : "",
  ].filter(Boolean)

  return (
    <>
      {title ? <div className="font-medium">{String(title)}</div> : null}
      {meta.length ? (
        <div className="text-xs text-muted-foreground">{meta.join(" · ")}</div>
      ) : null}
      {content ? (
        contentType === "html" ? (
          <HtmlImagePreview
            html={String(content)}
            className="bbs-content max-w-none break-words [&_img]:cursor-zoom-in"
          />
        ) : (
          <div className="whitespace-pre-wrap break-words">
            {String(content)}
          </div>
        )
      ) : (
        <div>-</div>
      )}
    </>
  )
}
