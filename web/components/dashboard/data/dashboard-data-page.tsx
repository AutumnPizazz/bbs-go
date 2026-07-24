"use client"

import * as React from "react"

import type { AdminFormValue, AdminRecord } from "@/lib/api/admin"
import { adminPostForm } from "@/lib/api/admin"
import type { PermissionCode } from "@/lib/auth/permissions.generated"
import { useCurrentUser } from "@/components/app/app-provider"
import { ErrorPage } from "@/components/common/error-page"
import { ConfirmDialog } from "@/components/dashboard/confirm-dialog"
import { userHasPermission } from "@/lib/auth/roles"
import { useI18n } from "@/lib/i18n/provider"
import { msgSuccess } from "@/lib/toast"

import { DashboardDataBulkDialog } from "./dashboard-data-bulk-dialog"
import { DashboardDataDetailDialog } from "./dashboard-data-detail-dialog"
import { DashboardDataFormDialog } from "./dashboard-data-form-dialog"
import { DashboardDataPasswordDialog } from "./dashboard-data-password-dialog"
import { DashboardDataTable } from "./dashboard-data-table"
import { DashboardDataToolbar } from "./dashboard-data-toolbar"
import type {
  DashboardDataBulkAction,
  DashboardDataBulkPreview,
  DashboardDataBulkResult,
  DashboardDataPageConfig,
} from "./dashboard-data-types"
import { useDashboardDataPage } from "./use-dashboard-data-page"

export function DashboardDataPage({
  config,
}: {
  config: DashboardDataPageConfig
}) {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const formId = "dashboard-data-form"
  const canUse = (permission?: PermissionCode) =>
    !permission || userHasPermission(currentUser, permission)
  const visibleConfig = React.useMemo(
    () => ({
      ...config,
      rowActions: config.rowActions?.filter((action) =>
        canUse(action.permission)
      ),
      bulkActions: config.bulkActions?.filter((action) =>
        canUse(action.permission)
      ),
    }),
    [config, currentUser]
  )
  const canView = canUse(config.viewPermission)
  const state = useDashboardDataPage({
    config: visibleConfig,
    messages: {
      loadFailed: t("dashboard.errors.loadFailed"),
      saveFailed: t("dashboard.errors.saveFailed"),
      deleteFailed: t("dashboard.errors.deleteFailed"),
      actionFailed: t("dashboard.errors.actionFailed"),
      sameLevelSortOnly: t("dashboard.errors.sameLevelSortOnly"),
      required: t("dashboard.errors.required"),
      invalidUrl: t("dashboard.errors.invalidUrl"),
      invalidNumber: t("dashboard.errors.invalidNumber"),
      minValue: (min) => t("dashboard.errors.minValue", { min }),
      maxValue: (max) => t("dashboard.errors.maxValue", { max }),
      saved: t("dashboard.messages.saved"),
      deleted: t("dashboard.messages.deleted"),
      actionDone: t("dashboard.messages.actionDone"),
      confirmDelete: t("dashboard.confirmDelete"),
      deleteAction: t("dashboard.actions.delete"),
      saveViewPrompt: t("dashboard.actions.saveViewPrompt"),
    },
  })

  const [selectedIds, setSelectedIds] = React.useState<Set<string>>(
    () => new Set()
  )
  const [bulkState, setBulkState] = React.useState<{
    action: DashboardDataBulkAction
    ids: Array<string | number>
    selectedCount: number
    preview: DashboardDataBulkPreview | null
    result: DashboardDataBulkResult | null
    confirmText: string
    submitting: boolean
    error: string | null
  } | null>(null)
  const [recordCache, setRecordCache] = React.useState(
    () => new Map<string, AdminRecord>()
  )
  const selectableRecords = state.displayRecords.filter(
    (record) => record.id !== undefined && record.id !== null
  )
  const selectedRecords = Array.from(selectedIds).map(
    (id) => recordCache.get(id) || ({ id } satisfies AdminRecord)
  )
  const allSelected =
    selectableRecords.length > 0 &&
    selectableRecords.every((record) => selectedIds.has(String(record.id)))

  React.useEffect(() => {
    if (!selectableRecords.length) return
    setRecordCache((current) => {
      const next = new Map(current)
      let changed = false
      selectableRecords.forEach((record) => {
        const id = String(record.id)
        if (next.get(id) !== record) {
          next.set(id, record)
          changed = true
        }
      })
      return changed ? next : current
    })
  }, [state.displayRecords])

  function toggleSelectedRecord(record: Record<string, unknown>) {
    if (record.id === undefined || record.id === null) return
    const id = String(record.id)
    setSelectedIds((current) => {
      const next = new Set(current)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  function toggleAllSelected() {
    setSelectedIds(
      allSelected
        ? new Set()
        : new Set(selectableRecords.map((record) => String(record.id)))
    )
  }

  function bulkIds(records: Array<Record<string, unknown>>) {
    return records
      .map((record) => record.id)
      .filter(
        (id): id is string | number =>
          typeof id === "string" || typeof id === "number"
      )
  }

  function recordsForIds(ids: Array<string | number>) {
    return ids.map(
      (id) => recordCache.get(String(id)) || ({ id } satisfies AdminRecord)
    )
  }

  async function openBulkAction(
    action: DashboardDataBulkAction,
    retryIds?: Array<string | number>
  ) {
    const ids = retryIds || bulkIds(selectedRecords)
    if (!ids.length) return
    setBulkState({
      action,
      ids,
      selectedCount: ids.length,
      preview: null,
      result: null,
      confirmText: "",
      submitting: true,
      error: null,
    })
    try {
      const data = await adminPostForm<DashboardDataBulkPreview>(
        action.previewEndpoint,
        { ...(action.payload?.(recordsForIds(ids)) ?? {}), ids }
      )
      setBulkState((current) =>
        current
          ? { ...current, preview: data, submitting: false, error: null }
          : current
      )
    } catch (err) {
      setBulkState((current) =>
        current
          ? {
              ...current,
              error:
                err instanceof Error
                  ? err.message
                  : t("dashboard.errors.actionFailed"),
              submitting: false,
            }
          : current
      )
    }
  }

  async function performBulkAction() {
    if (!bulkState?.preview) return
    const current = bulkState
    setBulkState({ ...current, submitting: true, error: null })
    try {
      const result = await adminPostForm<DashboardDataBulkResult>(
        current.action.endpoint,
        {
          ...(current.action.payload?.(recordsForIds(current.ids)) ?? {}),
          ids: current.ids,
          confirmText: current.confirmText,
        }
      )
      const failures = result.failures || []
      await state.load()
      if (failures.length) {
        setSelectedIds(new Set(failures.map((failure) => String(failure.id))))
        setBulkState({
          ...current,
          preview: null,
          result,
          confirmText: "",
          submitting: false,
          error: null,
        })
        return
      }
      msgSuccess(
        current.action.successMessage || t("dashboard.messages.actionDone")
      )
      setBulkState(null)
      setSelectedIds(new Set())
      setRecordCache(new Map())
    } catch (err) {
      setBulkState({
        ...current,
        submitting: false,
        error:
          err instanceof Error
            ? err.message
            : t("dashboard.errors.actionFailed"),
      })
    }
  }

  function retryBulkFailures() {
    if (!bulkState?.result?.failures?.length) return
    void openBulkAction(
      bulkState.action,
      bulkState.result.failures.map((failure) => failure.id)
    )
  }

  if (!canView) {
    return <ErrorPage statusCode={403} />
  }

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-4 md:p-6">
      <DashboardDataToolbar
        filters={visibleConfig.filters}
        values={state.filters}
        asyncOptions={state.asyncOptions}
        loading={state.loading}
        canCreate={Boolean(
          visibleConfig.createEndpoint &&
          visibleConfig.formFields?.length &&
          canUse(visibleConfig.createPermission)
        )}
        error={state.error}
        searchLabel={t("dashboard.actions.search")}
        refreshLabel={t("dashboard.actions.refresh")}
        createLabel={t("dashboard.actions.create")}
        saveLabel={t("dashboard.actions.saveView")}
        onFilterChange={state.updateFilter}
        onRefresh={() => void state.load()}
        onCreate={state.openCreate}
        onSaveFilters={state.saveFilters}
        savedViews={state.savedViews}
        onLoadView={state.loadSavedView}
        loadViewLabel={t("dashboard.actions.loadView")}
        clearSelectionLabel={t("dashboard.bulk.clearSelection")}
        onClearSelection={() => {
          setSelectedIds(new Set())
          setRecordCache(new Map())
        }}
        bulkActions={visibleConfig.bulkActions?.map((action) => ({
          label: action.label,
          onClick: () => void openBulkAction(action),
        }))}
        selectedCount={selectedRecords.length}
        selectedLabel={(count) => t("dashboard.bulk.selected", { count })}
      />

      <DashboardDataTable
        config={visibleConfig}
        records={state.displayRecords}
        loading={state.loading}
        page={state.page}
        pageCount={state.pageCount}
        total={state.total}
        limit={state.limit}
        labels={{
          actions: t("dashboard.actions.title"),
          loading: t("dashboard.loading"),
          noData: t("common.noData"),
          moveUp: t("dashboard.actions.moveUp"),
          moveDown: t("dashboard.actions.moveDown"),
          expand: t("dashboard.actions.expand"),
          collapse: t("dashboard.actions.collapse"),
          view: t("dashboard.actions.view"),
          edit: t("dashboard.actions.edit"),
          delete: t("dashboard.actions.delete"),
          selectAll: t("dashboard.bulk.selectAll"),
          selectRow: t("dashboard.bulk.selectRow"),
        }}
        onPageChange={(nextPage) => state.updateFilter("page", nextPage)}
        onLimitChange={(nextLimit) =>
          state.setFilters((current) => ({
            ...current,
            page: 1,
            limit: nextLimit,
          }))
        }
        onMove={(index, direction) => void state.moveRecord(index, direction)}
        onReorder={(fromIndex, toIndex) =>
          void state.reorderRecord(fromIndex, toIndex)
        }
        canMove={state.canMoveRecord}
        canSort={canUse(visibleConfig.sortPermission)}
        canUpdate={canUse(visibleConfig.updatePermission)}
        canDelete={canUse(visibleConfig.deletePermission)}
        onRunAction={(action, record) => void state.runAction(action, record)}
        onView={(record) => void state.openView(record)}
        onEdit={(record) => void state.openEdit(record)}
        onDelete={state.requestDelete}
        isTreeRecordCollapsed={state.isTreeRecordCollapsed}
        onToggleTreeRecord={state.toggleTreeRecord}
        selectable={Boolean(visibleConfig.bulkActions?.length)}
        selectedIds={selectedIds}
        allSelected={allSelected}
        onToggleRecord={toggleSelectedRecord}
        onToggleAll={toggleAllSelected}
      />

      <DashboardDataFormDialog
        open={Boolean(state.editing)}
        formId={formId}
        title={
          state.formValues.id
            ? t("dashboard.actions.edit")
            : t("dashboard.actions.create")
        }
        fields={state.visibleFormFields}
        values={state.formValues}
        errors={state.formErrors}
        asyncOptions={state.asyncOptions}
        submitting={state.submitting}
        cancelLabel={t("common.cancel")}
        confirmLabel={t("common.confirm")}
        onOpenChange={(open) => {
          if (!open) state.setEditing(null)
        }}
        onSubmit={(event) => void state.submitForm(event)}
        onValueChange={(name, value: AdminFormValue) =>
          state.setFormValues((current) => ({
            ...current,
            [name]: value,
          }))
        }
      />

      <DashboardDataDetailDialog
        record={state.viewing}
        fields={config.detailFields}
        title={t("dashboard.actions.view")}
        cancelLabel={t("common.cancel")}
        onClose={() => state.setViewing(null)}
      />

      <DashboardDataPasswordDialog
        password={state.passwordResult}
        title={t("dashboard.resetPassword.title")}
        passwordLabel={t("dashboard.resetPassword.newPassword")}
        copyLabel={t("dashboard.resetPassword.copy")}
        copiedMessage={t("dashboard.resetPassword.copied")}
        cancelLabel={t("common.cancel")}
        onClose={() => state.setPasswordResult(null)}
      />

      <ConfirmDialog
        state={state.confirmState}
        onOpenChange={(open) => {
          if (!open) state.setConfirmState(null)
        }}
      />

      <DashboardDataBulkDialog
        action={bulkState?.action || null}
        selectedCount={bulkState?.selectedCount || 0}
        preview={bulkState?.preview || null}
        result={bulkState?.result || null}
        confirmText={bulkState?.confirmText || ""}
        submitting={Boolean(bulkState?.submitting)}
        error={bulkState?.error || null}
        onConfirmTextChange={(value) =>
          setBulkState((current) =>
            current ? { ...current, confirmText: value } : current
          )
        }
        onClose={() => setBulkState(null)}
        onConfirm={() => void performBulkAction()}
        onRetry={retryBulkFailures}
      />
    </div>
  )
}
