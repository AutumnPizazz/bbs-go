"use client"

import * as React from "react"
import { ExternalLinkIcon, KeyRoundIcon, XIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { apiFetch } from "@/lib/api/client"
import { useI18n } from "@/lib/i18n/provider"
import { toast } from "@/lib/toast"

type PasswordChangeDialogProps = {
  onClose: () => void
}

export function DashboardPasswordChangeDialog({
  onClose,
}: PasswordChangeDialogProps) {
  const { t } = useI18n()
  const [currentPassword, setCurrentPassword] = React.useState("")
  const [password, setPassword] = React.useState("")
  const [rePassword, setRePassword] = React.useState("")
  const [pending, setPending] = React.useState(false)
  const [staged, setStaged] = React.useState(false)

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (password !== rePassword) {
      toast.error(t("dashboard.user.passwordChange.mismatch"))
      return
    }

    setPending(true)
    try {
      await apiFetch("/api/admin/user/update_password", {
        method: "POST",
        body: { currentPassword, password, rePassword },
      })
      setCurrentPassword("")
      setPassword("")
      setRePassword("")
      setStaged(true)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t("composables.unknownError"))
    } finally {
      setPending(false)
    }
  }

  function openSignin() {
    window.open(
      "/user/signin?redirect=%2Fdashboard",
      "_blank",
      "noopener,noreferrer"
    )
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4">
      <div
        className="w-full max-w-md rounded-lg bg-background p-5 shadow-lg"
        role="dialog"
        aria-modal="true"
        aria-labelledby="dashboard-password-change-title"
      >
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2
            id="dashboard-password-change-title"
            className="flex items-center gap-2 text-lg font-semibold"
          >
            <KeyRoundIcon className="size-5" />
            {t("dashboard.user.passwordChange.title")}
          </h2>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={t("dialog.cancel")}
            onClick={onClose}
          >
            <XIcon />
          </Button>
        </div>

        {staged ? (
          <div className="space-y-5">
            <p className="text-sm text-muted-foreground">
              {t("dashboard.user.passwordChange.pendingDescription")}
            </p>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={onClose}>
                {t("dialog.cancel")}
              </Button>
              <Button type="button" onClick={openSignin}>
                <ExternalLinkIcon className="size-4" />
                {t("dashboard.user.passwordChange.openSignin")}
              </Button>
            </div>
          </div>
        ) : (
          <form className="space-y-4" onSubmit={submit}>
            <PasswordField
              id="current-password"
              label={t("dashboard.user.passwordChange.currentPassword")}
              value={currentPassword}
              autoComplete="current-password"
              onChange={setCurrentPassword}
            />
            <PasswordField
              id="new-password"
              label={t("dashboard.user.passwordChange.newPassword")}
              value={password}
              autoComplete="new-password"
              onChange={setPassword}
            />
            <PasswordField
              id="new-password-confirm"
              label={t("dashboard.user.passwordChange.confirmPassword")}
              value={rePassword}
              autoComplete="new-password"
              onChange={setRePassword}
            />
            <div className="flex justify-end gap-2 pt-2">
              <Button type="button" variant="outline" onClick={onClose}>
                {t("dialog.cancel")}
              </Button>
              <Button type="submit" disabled={pending}>
                <KeyRoundIcon className="size-4" />
                {pending
                  ? t("dashboard.user.passwordChange.submitting")
                  : t("dashboard.user.passwordChange.submit")}
              </Button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}

function PasswordField({
  id,
  label,
  value,
  autoComplete,
  onChange,
}: {
  id: string
  label: string
  value: string
  autoComplete: "current-password" | "new-password"
  onChange: (value: string) => void
}) {
  return (
    <div className="space-y-2">
      <label htmlFor={id} className="text-sm font-medium">
        {label}
      </label>
      <Input
        id={id}
        type="password"
        autoComplete={autoComplete}
        value={value}
        required
        onChange={(event) => onChange(event.currentTarget.value)}
      />
    </div>
  )
}
