"use client"

import { ShieldAlertIcon } from "lucide-react"

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
} from "./dashboard-data-types"

export function DashboardDataBulkDialog({
  action,
  selectedCount,
  preview,
  confirmText,
  submitting,
  error,
  onConfirmTextChange,
  onClose,
  onConfirm,
}: {
  action: DashboardDataBulkAction | null
  selectedCount: number
  preview: DashboardDataBulkPreview | null
  confirmText: string
  submitting: boolean
  error: string | null
  onConfirmTextChange: (value: string) => void
  onClose: () => void
  onConfirm: () => void
}) {
  const { t } = useI18n()
  const expectedText = String(preview?.confirmText || "")
  const canConfirm = Boolean(
    action && preview && expectedText && confirmText === expectedText
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

        {preview ? (
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

        {error ? (
          <div className="rounded-md border border-destructive/25 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {error}
          </div>
        ) : null}

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            {t("common.cancel")}
          </Button>
          <Button
            type="button"
            variant="destructive"
            disabled={!canConfirm || submitting}
            onClick={onConfirm}
          >
            <ShieldAlertIcon />
            {submitting ? t("dashboard.loading") : t("dashboard.bulk.confirm")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
