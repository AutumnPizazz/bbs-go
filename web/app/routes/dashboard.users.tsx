"use client"

import * as React from "react"
import { CheckCircleIcon, CopyIcon, KeyRoundIcon, MailIcon, XCircleIcon } from "lucide-react"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { DashboardPasswordChangeDialog } from "@/components/dashboard/dashboard-password-change-dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import { adminPostJson } from "@/lib/api/admin"
import { toast } from "@/lib/toast"
import type { AdminRecord } from "@/lib/api/admin"
import { useAppState } from "@/components/app/app-provider"
import { userHasPermission } from "@/lib/auth/roles"

const OWNER_ROLE = "owner"

function isOwnerRecord(record: AdminRecord) {
  const roles = record.roles
  if (typeof roles === "string") {
    return roles.split(",").some((r) => r.trim() === OWNER_ROLE)
  }
  if (Array.isArray(roles)) {
    return roles.some((r: unknown) => typeof r === "string" && r === OWNER_ROLE)
  }
  return false
}

export default function DashboardUsersRoute() {
  const { t } = useI18n()
  const [passwordChangePending, setPasswordChangePending] = React.useState(false)

  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "users"),
    description: dashboardData.desc(t, "users"),
    listEndpoint: "/api/admin/user/list",
    viewPermission: PERMISSIONS.DASHBOARD_USER_VIEW,
    detailEndpoint: (id) => `/api/admin/user/${id}`,
    createEndpoint: "/api/admin/user/create",
    createPermission: PERMISSIONS.DASHBOARD_USER_CREATE,
    updateEndpoint: "/api/admin/user/update",
    updatePermission: PERMISSIONS.DASHBOARD_USER_UPDATE,
    filters: [
      { name: "id", label: dashboardData.label(t, "id") },
      { name: "username", label: dashboardData.label(t, "username") },
      { name: "nickname", label: dashboardData.label(t, "nickname") },
      {
        name: "forbidden",
        label: dashboardData.label(t, "forbidden"),
        type: "select",
        options: [
          { label: t("dashboard.boolean.yes"), value: "true" },
          { label: t("dashboard.boolean.no"), value: "false" },
        ],
      },
    ],
    columns: [
      {
        key: "id",
        label: dashboardData.label(t, "id"),
        render: (record) => dashboardData.userLinkCell(record, record.id),
      },
      {
        key: "idEncode",
        label: dashboardData.label(t, "idEncode"),
        render: (record) => dashboardData.userLinkCell(record, record.idEncode),
      },
      {
        key: "avatar",
        label: dashboardData.label(t, "avatar"),
        render: (record) =>
          dashboardData.imageCell(
            record.avatar || record.smallAvatar,
            String(record.nickname || record.username || "")
          ),
      },
      {
        key: "username",
        label: dashboardData.label(t, "username"),
        render: (record) => dashboardData.userLinkCell(record, record.username),
      },
      {
        key: "nickname",
        label: dashboardData.label(t, "nickname"),
        render: (record) => dashboardData.userLinkCell(record, record.nickname),
      },
      {
        key: "contentAccessMode",
        label: dashboardData.label(t, "contentAccessMode"),
        render: (record) =>
          record.contentAccessMode === "all"
            ? t("dashboard.userAccess.all")
            : t("dashboard.userAccess.assignedCategories"),
      },
      {
        key: "forbidden",
        label: dashboardData.label(t, "forbidden"),
        render: (record) =>
          record.forbidden
            ? t("dashboard.boolean.yes")
            : t("dashboard.boolean.no"),
      },
      {
        key: "forbiddenEndTime",
        label: dashboardData.label(t, "forbiddenEndTime"),
        render: (record) =>
          Number(record.forbiddenEndTime) === -1
            ? t("dashboard.forbidden.permanent")
            : dashboardData.dateCell(record.forbiddenEndTime),
      },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      {
        name: "username",
        label: dashboardData.label(t, "username"),
      },
      {
        name: "nickname",
        label: dashboardData.label(t, "nickname"),
        required: true,
      },
      {
        name: "avatar",
        label: dashboardData.label(t, "avatar"),
        type: "image",
      },
      {
        name: "gender",
        label: dashboardData.label(t, "gender"),
        type: "select",
        options: [
          { label: t("dashboard.gender.male"), value: "Male" },
          { label: t("dashboard.gender.female"), value: "Female" },
        ],
      },
      { name: "homePage", label: dashboardData.label(t, "homePage") },
      {
        name: "description",
        label: dashboardData.label(t, "description"),
        type: "textarea",
      },
      {
        name: "roleIds",
        label: dashboardData.label(t, "roles"),
        type: "multiselect",
        optionsEndpoint: "/api/admin/role/roles",
        optionLabel: (record) =>
          String(record.name || record.code || record.id),
        optionValue: (record) => record.id as number,
        valueFromRecord: (record) =>
          Array.isArray(record.roleIds)
            ? record.roleIds.map((item) => String(item))
            : [],
      },
      {
        name: "password",
        label: dashboardData.label(t, "password"),
        type: "custom",
        render: ({ value, onChange, record }) => {
          // Create mode: show a normal password input
          if (!record) {
            return (
              <Input
                type="password"
                autoComplete="new-password"
                value={value === undefined || value === null ? "" : String(value)}
                onChange={(event) => onChange(event.target.value)}
              />
            )
          }
          // Edit mode — owner: show "修改密码" button that opens the staged dialog
          if (isOwnerRecord(record)) {
            return (
              <Button
                type="button"
                variant="outline"
                className="w-full"
                onClick={() => setPasswordChangePending(true)}
              >
                <KeyRoundIcon className="mr-2 size-4" />
                {t("dashboard.user.passwordChange.title")}
              </Button>
            )
          }
          // Edit mode — non-owner: show a normal password input
          return (
            <Input
              type="password"
              autoComplete="new-password"
              value={value === undefined || value === null ? "" : String(value)}
              onChange={(event) => onChange(event.target.value)}
            />
          )
        },
      },
      {
        name: "contentAccessMode",
        label: dashboardData.label(t, "contentAccessMode"),
        type: "select",
        required: true,
        options: [
          { label: t("dashboard.userAccess.all"), value: "all" },
          {
            label: t("dashboard.userAccess.assignedCategories"),
            value: "assigned_categories",
          },
        ],
      },
      {
        name: "categoryIds",
        label: dashboardData.label(t, "categoryIds"),
        type: "multiselect",
        optionsEndpoint: "/api/admin/category/options",
        visibleWhen: (values) =>
          values.contentAccessMode === "assigned_categories",
        optionLabel: dashboardData.treeOptionLabel,
        optionValue: (record) => record.id as number,
        valueFromRecord: (record) =>
          Array.isArray(record.categoryIds)
            ? record.categoryIds.map((item) => String(item))
            : [],
      },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        required: true,
        options: dashboardData.normalDeletedOptions(t),
      },
    ],
    rowActions: [
      {
        label: t("dashboard.forbidden.ban7Days"),
        endpoint: "/api/admin/user/forbidden",
        permission: PERMISSIONS.DASHBOARD_USER_FORBIDDEN,
        payload: (record) => ({
          userId: record.id as number,
          days: 7,
          reason: t("dashboard.forbidden.reasonTemporary"),
        }),
        confirm: t("dashboard.forbidden.confirm7Days"),
        successMessage: t("dashboard.forbidden.banned"),
        visible: (record) => !record.forbidden,
      },
      {
        label: t("dashboard.forbidden.banForever"),
        endpoint: "/api/admin/user/forbidden",
        permission: PERMISSIONS.DASHBOARD_USER_FORBIDDEN_FOREVER,
        payload: (record) => ({
          userId: record.id as number,
          days: -1,
          reason: t("dashboard.forbidden.reasonPermanent"),
        }),
        confirm: t("dashboard.forbidden.confirmForever"),
        successMessage: t("dashboard.forbidden.banned"),
        visible: (record) => !record.forbidden,
      },
      {
        label: t("dashboard.forbidden.remove"),
        endpoint: "/api/admin/user/forbidden",
        permission: PERMISSIONS.DASHBOARD_USER_FORBIDDEN,
        payload: (record) => ({ userId: record.id as number, days: 0 }),
        confirm: t("dashboard.forbidden.confirmRemove"),
        successMessage: t("dashboard.forbidden.removed"),
        visible: (record) => Boolean(record.forbidden),
      },
      {
        label: t("dashboard.user.passwordChange.title"),
        endpoint: "",
        permission: PERMISSIONS.DASHBOARD_USER_RESET_PASSWORD,
        visible: (record) => isOwnerRecord(record),
        onClick: () => setPasswordChangePending(true),
      },
      {
        label: t("dashboard.actions.resetPassword"),
        endpoint: "/api/admin/user/reset_password",
        permission: PERMISSIONS.DASHBOARD_USER_RESET_PASSWORD,
        payload: (record) => ({ userId: record.id as number }),
        confirm: t("dashboard.confirmResetPassword"),
        visible: (record) => !isOwnerRecord(record),
      },
    ],
    bulkActions: [
      {
        label: t("dashboard.forbidden.batchBan7Days"),
        endpoint: "/api/admin/user/batch",
        previewEndpoint: "/api/admin/user/batch/preview",
        permission: PERMISSIONS.DASHBOARD_USER_BATCH_FORBIDDEN,
        payload: () => ({
          action: "forbid",
          days: 7,
          reason: t("dashboard.forbidden.reasonTemporary"),
        }),
        successMessage: t("dashboard.forbidden.batchBanned"),
      },
      {
        label: t("dashboard.forbidden.batchBanForever"),
        endpoint: "/api/admin/user/batch",
        previewEndpoint: "/api/admin/user/batch/preview",
        permission: PERMISSIONS.DASHBOARD_USER_BATCH_FORBIDDEN_FOREVER,
        payload: () => ({
          action: "forbid",
          days: -1,
          reason: t("dashboard.forbidden.reasonPermanent"),
        }),
        successMessage: t("dashboard.forbidden.batchBanned"),
      },
      {
        label: t("dashboard.forbidden.batchRemove"),
        endpoint: "/api/admin/user/batch",
        previewEndpoint: "/api/admin/user/batch/preview",
        permission: PERMISSIONS.DASHBOARD_USER_BATCH_FORBIDDEN,
        payload: () => ({ action: "removeForbidden" }),
        successMessage: t("dashboard.forbidden.batchRemoved"),
      },
    ],
    transformSubmitValues: (values) => {
      if (!values.password) {
        const { password: _, ...rest } = values
        return rest
      }
      return values
    },
    onSubmitSuccess: (response: unknown) => {
      const data = response as Record<string, unknown> | null | undefined
      if (data?.pendingAdminPassword) {
        setPasswordChangePending(true)
      }
    },
  }

  return (
    <>
      <BatchRegisterPanel />
      <DashboardDataPage config={config} />
      {passwordChangePending ? (
        <DashboardPasswordChangeDialog
          onClose={() => setPasswordChangePending(false)}
        />
      ) : null}
    </>
  )
}

function BatchRegisterPanel() {
  const { t } = useI18n()
  const { currentUser } = useAppState()
  const canCreate = userHasPermission(currentUser, PERMISSIONS.DASHBOARD_USER_CREATE)
  const [expanded, setExpanded] = React.useState(false)
  const [emails, setEmails] = React.useState("")
  const [pending, setPending] = React.useState(false)
  const [results, setResults] = React.useState<BatchRegisterResult[] | null>(null)
  const [summary, setSummary] = React.useState<BatchRegisterSummary | null>(null)
  const [copied, setCopied] = React.useState(false)

  if (!canCreate) return null

  async function submit() {
    if (!emails.trim()) return
    setPending(true)
    setResults(null)
    setSummary(null)
    try {
      const data = await adminPostJson<BatchRegisterResponse>("/api/admin/user/batch_register", { emails })
      setResults(data.results)
      setSummary(data.summary)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("composables.unknownError"))
    } finally {
      setPending(false)
    }
  }

  function copyAll() {
    if (!results) return
    const lines = results.map((r) => `${r.username}\t${r.email}\t${r.password ?? "-"}\t${t(`dashboard.batchRegister.status.${r.status}`)}`)
    void navigator.clipboard.writeText(lines.join("\n"))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  function copyPassword(pw: string) {
    void navigator.clipboard.writeText(pw)
    toast.success(t("dashboard.batchRegister.copied"))
  }

  return (
    <section className="border-b px-4 py-3 md:px-6">
      <button
        type="button"
        className="flex w-full items-center gap-2 text-sm font-medium text-muted-foreground hover:text-foreground"
        onClick={() => setExpanded(!expanded)}
      >
        <MailIcon className="size-4" />
        {t("dashboard.batchRegister.title")}
        <span className="text-xs">{expanded ? "▲" : "▼"}</span>
      </button>

      {expanded ? (
        <div className="mt-3 space-y-4">
          <p className="text-sm text-muted-foreground">
            {t("dashboard.batchRegister.description")}
          </p>

          <div className="space-y-2">
            <label className="text-sm font-medium">
              {t("dashboard.batchRegister.emailsLabel")}
            </label>
            <textarea
              className="w-full min-h-24 rounded-md border bg-background px-3 py-2 text-sm"
              placeholder={t("dashboard.batchRegister.emailsPlaceholder")}
              value={emails}
              onChange={(e) => setEmails(e.target.value)}
              disabled={pending}
            />
          </div>

          <Button type="button" disabled={pending || !emails.trim()} onClick={() => void submit()}>
            {pending ? t("dashboard.batchRegister.submitting") : t("dashboard.batchRegister.submit")}
          </Button>

          {results ? (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold">
                  {t("dashboard.batchRegister.resultTitle")}
                </h3>
                <Button type="button" variant="outline" size="sm" onClick={copyAll}>
                  <CopyIcon className="size-3.5" />
                  {copied ? t("dashboard.batchRegister.copied") : t("dashboard.batchRegister.copyAll")}
                </Button>
              </div>

              {summary ? (
                <p className="text-xs text-muted-foreground">
                  {t("dashboard.batchRegister.summary")
                    .replace("{total}", String(summary.total))
                    .replace("{created}", String(summary.created))
                    .replace("{skipped}", String(summary.skipped))
                    .replace("{failed}", String(summary.failed))}
                </p>
              ) : null}

              <div className="overflow-x-auto rounded-md border">
                <table className="w-full text-sm">
                  <thead className="bg-muted">
                    <tr>
                      <th className="px-3 py-2 text-left">{t("dashboard.batchRegister.columns.username")}</th>
                      <th className="px-3 py-2 text-left">{t("dashboard.batchRegister.columns.email")}</th>
                      <th className="px-3 py-2 text-left">{t("dashboard.batchRegister.columns.password")}</th>
                      <th className="px-3 py-2 text-left">{t("dashboard.batchRegister.columns.status")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.map((r, i) => (
                      <tr key={i} className="border-t">
                        <td className="px-3 py-1.5">{r.username}</td>
                        <td className="px-3 py-1.5">{r.email}</td>
                        <td className="px-3 py-1.5">
                          {r.password ? (
                            <span className="inline-flex items-center gap-1">
                              <code className="text-xs">{r.password}</code>
                              <button
                                type="button"
                                className="text-muted-foreground hover:text-foreground"
                                onClick={() => copyPassword(r.password!)}
                              >
                                <CopyIcon className="size-3" />
                              </button>
                            </span>
                          ) : (
                            <span className="text-muted-foreground">—</span>
                          )}
                        </td>
                        <td className="px-3 py-1.5">
                          <span
                            className={`inline-flex items-center gap-1 text-xs ${
                              r.status === "created"
                                ? "text-green-600"
                                : r.status === "skipped"
                                  ? "text-amber-600"
                                  : "text-red-600"
                            }`}
                          >
                            {r.status === "created" ? (
                              <CheckCircleIcon className="size-3" />
                            ) : r.status === "skipped" ? (
                              <span className="text-xs">⚠</span>
                            ) : (
                              <XCircleIcon className="size-3" />
                            )}
                            {t(`dashboard.batchRegister.status.${r.status}`)}
                            {r.reason ? (
                              <span className="text-muted-foreground">({r.reason})</span>
                            ) : null}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ) : null}
        </div>
      ) : null}
    </section>
  )
}

type BatchRegisterResult = {
  username: string
  email: string
  password?: string
  status: string
  reason?: string
}

type BatchRegisterSummary = {
  total: number
  created: number
  skipped: number
  failed: number
}

type BatchRegisterResponse = {
  results: BatchRegisterResult[]
  summary: BatchRegisterSummary
}
