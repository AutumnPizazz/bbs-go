"use client"

import { RefreshCwIcon, ShieldAlertIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useI18n } from "@/lib/i18n/provider"

import type {
  DashboardDataBulkAction,
  DashboardDataBulkPreview,
  DashboardDataBulkResult,
} from "./dashboard-data-types"

export function DashboardDataBulkDialog({
  action,
  selectedCount,
  preview,
  result,
  confirmText,
  submitting,
  error,
  onConfirmTextChange,
  onClose,
  onConfirm,
  onRetry,
}: {
  action: DashboardDataBulkAction | null
  selectedCount: number
  preview: DashboardDataBulkPreview | null
  result: DashboardDataBulkResult | null
  confirmText: string
  submitting: boolean
  error: string | null
  onConfirmTextChange: (value: string) => void
  onClose: () => void
  onConfirm: () => void
  onRetry: () => void
}) {
  const { t } = useI18n()
  const expectedText = String(preview?.confirmText || "")
  const canConfirm = Boolean(
    action && !result && preview && expectedText && confirmText === expectedText
  )

  return (
    <Dialog open={Boolean(action)} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ShieldAlertIcon className="size-5 text-destructive" />
            {action?.label || t("dashboard.bulk.title")}
          </DialogTitle>
          <DialogDescription>
            {t("dashboard.bulk.scope", { count: selectedCount })}
          </DialogDescription>
        </DialogHeader>

        {result ? (
          <div className="grid gap-3 rounded-md border border-destructive/25 bg-destructive/10 p-3 text-sm">
            <div className="font-medium">
              {t("dashboard.bulk.result", {
                processed: Number(result.processedCount || 0),
                skipped: Number(result.skippedCount || 0),
                failed: Number(
                  result.failedCount || result.failures?.length || 0
                ),
              })}
            </div>
            {result.failures?.length ? (
              <div className="grid gap-2">
                <div className="font-medium">
                  {t("dashboard.bulk.failures", {
                    count: result.failures.length,
                  })}
                </div>
                <ul className="max-h-40 space-y-1 overflow-auto text-muted-foreground">
                  {result.failures.map((failure) => (
                    <li key={`${String(failure.id)}-${failure.message}`}>
                      #{String(failure.id)}: {failure.message}
                    </li>
                  ))}
                </ul>
              </div>
            ) : null}
          </div>
        ) : preview ? (
          <div className="grid gap-2 rounded-md border bg-muted/20 p-3 text-sm">
            <div>
              {t("dashboard.bulk.eligible", {
                count: Number(preview.eligibleCount || 0),
              })}
            </div>
            <div className="text-muted-foreground">
              {t("dashboard.bulk.ineligible", {
                count: Number(preview.ineligibleCount || 0),
              })}
            </div>
          </div>
        ) : null}

        {!result ? (
          <div className="grid gap-2">
            <Label htmlFor="dashboard-bulk-confirmation">
              {t("dashboard.bulk.confirmationLabel")}
            </Label>
            <Input
              id="dashboard-bulk-confirmation"
              value={confirmText}
              placeholder={expectedText}
              autoComplete="off"
              onChange={(event) => onConfirmTextChange(event.target.value)}
            />
          </div>
        ) : null}

        {error ? (
          <div className="rounded-md border border-destructive/25 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {error}
          </div>
        ) : null}

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            {t("common.cancel")}
          </Button>
          {result ? (
            <Button
              type="button"
              variant="outline"
              disabled={!result.failures?.length || submitting}
              onClick={onRetry}
            >
              <RefreshCwIcon />
              {submitting
                ? t("dashboard.loading")
                : t("dashboard.bulk.retryFailures")}
            </Button>
          ) : (
            <Button
              type="button"
              variant="destructive"
              disabled={!canConfirm || submitting}
              onClick={onConfirm}
            >
              <ShieldAlertIcon />
              {submitting
                ? t("dashboard.loading")
                : t("dashboard.bulk.confirm")}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
