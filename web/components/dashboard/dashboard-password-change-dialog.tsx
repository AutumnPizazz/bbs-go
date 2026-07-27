"use client"

import * as React from "react"
import {
  CheckCircleIcon,
  CircleIcon,
  ExternalLinkIcon,
  InfoIcon,
  KeyRoundIcon,
  XIcon,
} from "lucide-react"

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
      toast.error(
        error instanceof Error ? error.message : t("composables.unknownError")
      )
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
            disabled={pending}
          >
            <XIcon />
          </Button>
        </div>

        {staged ? (
          <StagedView onClose={onClose} onOpenSignin={openSignin} />
        ) : (
          <FormView
            currentPassword={currentPassword}
            password={password}
            rePassword={rePassword}
            pending={pending}
            onCurrentPasswordChange={setCurrentPassword}
            onPasswordChange={setPassword}
            onRePasswordChange={setRePassword}
            onSubmit={submit}
            onClose={onClose}
          />
        )}
      </div>
    </div>
  )
}

function FormView({
  currentPassword,
  password,
  rePassword,
  pending,
  onCurrentPasswordChange,
  onPasswordChange,
  onRePasswordChange,
  onSubmit,
  onClose,
}: {
  currentPassword: string
  password: string
  rePassword: string
  pending: boolean
  onCurrentPasswordChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onRePasswordChange: (value: string) => void
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void
  onClose: () => void
}) {
  const { t } = useI18n()

  return (
    <form className="space-y-4" onSubmit={onSubmit}>
      <div className="flex items-start gap-2 rounded-md bg-blue-50 px-3 py-2 text-sm text-blue-800 dark:bg-blue-950 dark:text-blue-200">
        <InfoIcon className="mt-0.5 size-4 shrink-0" />
        <span>{t("dashboard.user.passwordChange.formDescription")}</span>
      </div>

      <PasswordField
        id="current-password"
        label={t("dashboard.user.passwordChange.currentPassword")}
        value={currentPassword}
        autoComplete="current-password"
        onChange={onCurrentPasswordChange}
      />
      <PasswordField
        id="new-password"
        label={t("dashboard.user.passwordChange.newPassword")}
        value={password}
        autoComplete="new-password"
        onChange={onPasswordChange}
      />
      <PasswordField
        id="new-password-confirm"
        label={t("dashboard.user.passwordChange.confirmPassword")}
        value={rePassword}
        autoComplete="new-password"
        onChange={onRePasswordChange}
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
  )
}

function StagedView({
  onClose,
  onOpenSignin,
}: {
  onClose: () => void
  onOpenSignin: () => void
}) {
  const { t } = useI18n()

  const steps = [
    t("dashboard.user.passwordChange.pendingStep1"),
    t("dashboard.user.passwordChange.pendingStep2"),
    t("dashboard.user.passwordChange.pendingStep3"),
  ]

  return (
    <div className="space-y-5">
      {/* success header */}
      <div className="flex items-center gap-3 rounded-md bg-green-50 px-4 py-3 dark:bg-green-950">
        <CheckCircleIcon className="size-6 shrink-0 text-green-600 dark:text-green-400" />
        <div>
          <p className="text-sm font-semibold text-green-800 dark:text-green-200">
            {t("dashboard.user.passwordChange.pendingTitle")}
          </p>
          <p className="mt-0.5 text-xs text-green-700 dark:text-green-300">
            {t("dashboard.user.passwordChange.pendingNotice")}
          </p>
        </div>
      </div>

      {/* numbered steps */}
      <ol className="space-y-3">
        {steps.map((step, index) => (
          <li key={index} className="flex items-start gap-3">
            <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold text-muted-foreground">
              {index + 1}
            </span>
            <span className="pt-0.5 text-sm">{step}</span>
          </li>
        ))}
      </ol>

      <div className="flex justify-end gap-2 pt-1">
        <Button type="button" variant="outline" onClick={onClose}>
          {t("dialog.cancel")}
        </Button>
        <Button type="button" onClick={onOpenSignin}>
          <ExternalLinkIcon className="size-4" />
          {t("dashboard.user.passwordChange.openSignin")}
        </Button>
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
