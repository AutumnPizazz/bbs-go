"use client"

import { useRouter, useSearchParams } from "@/lib/router/navigation"
import { Lock } from "lucide-react"
import {
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  Suspense,
} from "react"

import type { AuthActionState } from "@/lib/actions/auth"
import {
  CaptchaChallenge,
  type CaptchaChallengeHandle,
} from "@/components/auth/captcha-field"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import type { SiteConfig } from "@/lib/api/types"
import type { LoginResult } from "@/lib/api/types"
import { apiFetch } from "@/lib/api/client"
import { useI18n } from "@/lib/i18n/provider"
import { safeRedirect } from "@/lib/site"
import { useToastActions } from "@/lib/toast"
import { cn } from "@/lib/utils"

const initialState: AuthActionState = { ok: false }

function enabled(value?: { enabled?: boolean }) {
  return value?.enabled !== false
}

function authRedirect(path?: string) {
  const safe = safeRedirect(path, "")
  if (
    !safe ||
    safe === "/user/signin" ||
    safe.startsWith("/user/signin?") ||
    safe.startsWith("/user/signin/")
  ) {
    return undefined
  }

  return safe
}

function navigateAfterAuth(path: string) {
  window.location.assign(path)
}

export function SigninForm({
  redirect,
  config,
}: {
  redirect?: string
  config?: SiteConfig | null
}) {
  return (
    <Suspense fallback={null}>
      <SigninFormContent redirect={redirect} config={config} />
    </Suspense>
  )
}

function SigninFormContent({
  redirect,
  config,
}: {
  redirect?: string
  config?: SiteConfig | null
}) {
  const searchParams = useSearchParams()
  const searchRedirect = searchParams.get("redirect") || undefined
  const effectiveRedirect = useMemo(
    () => authRedirect(redirect || searchRedirect),
    [redirect, searchRedirect]
  )
  const loginConfig = config?.loginConfig
  const passwordEnabled = enabled(loginConfig?.passwordLogin)
  const thirdPartyEnabled = Boolean(
    loginConfig?.githubLogin?.enabled ||
    loginConfig?.googleLogin?.enabled ||
    loginConfig?.weixinLogin?.enabled
  )
  const { t } = useI18n()

  return (
    <div className="signin-card mx-auto max-w-[600px]">
      <div className="py-2 break-all">
        {passwordEnabled ? (
          <div className="login-title">
            <Lock className="login-title-icon" />
            <h2>{t("user.signin.passwordLogin")}</h2>
          </div>
        ) : null}

        {passwordEnabled ? (
          <PasswordLoginForm redirect={effectiveRedirect} />
        ) : null}

        {thirdPartyEnabled ? (
          <ThirdPartyLogin
            config={config}
            hasAccountLogin={passwordEnabled}
            redirect={effectiveRedirect}
          />
        ) : null}
      </div>
    </div>
  )
}

function PasswordLoginForm({ redirect }: { redirect?: string }) {
  const { t } = useI18n()
  const { msgError } = useToastActions()
  const [state, setState] = useState<AuthActionState>(initialState)
  const [pending, setPending] = useState(false)
  const captchaRef = useRef<CaptchaChallengeHandle>(null)
  const formRef = useRef<HTMLFormElement>(null)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const lastErrorRef = useRef<string | undefined>(undefined)

  useEffect(() => {
    if (state.ok && state.redirect) {
      navigateAfterAuth(state.redirect)
    } else if (state.message && lastErrorRef.current !== state.message) {
      lastErrorRef.current = state.message
      msgError(state.message)
      captchaRef.current?.reset()
    }
  }, [msgError, state])

  return (
    <div className="password-login mx-auto max-w-[400px]">
      <form
        ref={formRef}
        className="space-y-6"
        onSubmit={async (event) => {
          event.preventDefault()
          if (!username) {
            msgError(t("user.signin.password.usernameRequired"))
            return
          }
          if (!password) {
            msgError(t("user.signin.password.passwordRequired"))
            return
          }
          if (!captchaRef.current?.hasCaptcha()) {
            void captchaRef.current?.open()
            return
          }

          setPending(true)
          try {
            const result = await apiFetch<LoginResult>("/api/login/signin", {
              method: "POST",
              body: new FormData(formRef.current!),
            })
            document.cookie = `bbsgo_token=${encodeURIComponent(result.token)}; Path=/; Max-Age=86400; SameSite=Lax`
            setState({ ok: true, redirect: safeRedirect(result.redirect, `/user/${result.user.id}`) })
          } catch (error) {
            setState({ ok: false, message: error instanceof Error ? error.message : "Sign in failed" })
          } finally {
            setPending(false)
          }
        }}
      >
        <input type="hidden" name="redirect" value={redirect || ""} />
        <CaptchaChallenge
          ref={captchaRef}
          onVerified={() => formRef.current?.requestSubmit()}
        />
        <div className="space-y-2">
          <Input
            name="username"
            autoComplete="username"
            placeholder={t("user.signin.password.usernamePlaceholder")}
            value={username}
            onChange={(event) => setUsername(event.currentTarget.value)}
          />
        </div>
        <div className="space-y-2">
          <Input
            name="password"
            type="password"
            autoComplete="current-password"
            placeholder={t("user.signin.password.passwordPlaceholder")}
            value={password}
            onChange={(event) => setPassword(event.currentTarget.value)}
          />
        </div>
        <Button type="submit" className="w-full" disabled={pending}>
          {t("user.signin.password.loginBtn")}
        </Button>
      </form>
    </div>
  )
}

function ThirdPartyLogin({
  config,
  hasAccountLogin,
  redirect,
}: {
  config?: SiteConfig | null
  hasAccountLogin: boolean
  redirect?: string
}) {
  const { t } = useI18n()
  const [weixinDialogOpen, setWeixinDialogOpen] = useState(false)
  const loginConfig = config?.loginConfig

  return (
    <div className="third-party-login">
      {hasAccountLogin ? (
        <div className="third-party-hint">
          <span>{t("user.signin.thirdPartyLogin")}</span>
        </div>
      ) : null}
      <div className="third-party-buttons">
        {loginConfig?.githubLogin?.enabled ? (
          <OAuthButton provider="github" redirect={redirect} />
        ) : null}
        {loginConfig?.googleLogin?.enabled ? (
          <OAuthButton provider="google" redirect={redirect} />
        ) : null}
        {loginConfig?.weixinLogin?.enabled ? (
          <Button
            type="button"
            variant="outline"
            className="weixin-trigger-btn w-full"
            onClick={() => setWeixinDialogOpen(true)}
          >
            <img
              src="/wechat.svg"
              alt="WeChat"
              className="weixin-icon mr-2 h-4 w-4"
            />
            {t("user.signin.weixinLogin")}
          </Button>
        ) : null}
      </div>
      {weixinDialogOpen && loginConfig?.weixinLogin?.enabled ? (
        <WeixinLoginDialog
          redirect={redirect}
          onClose={() => setWeixinDialogOpen(false)}
        />
      ) : null}
    </div>
  )
}

function OAuthButton({
  provider,
  redirect,
}: {
  provider: "github" | "google"
  redirect?: string
}) {
  const { t } = useI18n()
  const [loading, setLoading] = useState(false)
  const path =
    provider === "github"
      ? "/api/login/github_login_config"
      : "/api/login/google_login_config"
  const buttonText =
    provider === "github"
      ? t("user.signin.githubLoginButton")
      : t("user.signin.googleLoginButton")
  const loadingText =
    provider === "github"
      ? t("user.signin.githubLoggingIn")
      : t("user.signin.googleLoggingIn")

  return (
    <Button
      type="button"
      variant="outline"
      className="w-full"
      disabled={loading}
      onClick={async () => {
        setLoading(true)
        try {
          const data = await apiFetch<{ authUrl?: string }>(path, {
            params: { redirect: redirect || "/" },
          })
          if (data.authUrl) window.location.href = data.authUrl
        } catch {
          setLoading(false)
        }
      }}
    >
      <ProviderIcon provider={provider} loading={loading} />
      {loading ? loadingText : buttonText}
    </Button>
  )
}

function ProviderIcon({
  provider,
  loading,
}: {
  provider: "github" | "google"
  loading: boolean
}) {
  if (loading) return null
  if (provider === "github") {
    return (
      <svg
        className="mr-2 h-5 w-5"
        viewBox="0 0 24 24"
        fill="currentColor"
        aria-hidden="true"
      >
        <path
          fillRule="evenodd"
          clipRule="evenodd"
          d="M12 1.5C6.201 1.5 1.5 6.201 1.5 12c0 4.636 3.007 8.566 7.18 9.955.525.098.72-.228.72-.507 0-.25-.01-1.077-.014-1.953-2.92.635-3.536-1.236-3.536-1.236-.477-1.211-1.165-1.534-1.165-1.534-.952-.651.072-.638.072-.638 1.053.074 1.607 1.08 1.607 1.08.936 1.603 2.455 1.14 3.053.872.094-.678.366-1.14.666-1.402-2.33-.265-4.78-1.165-4.78-5.185 0-1.145.41-2.08 1.08-2.815-.108-.265-.468-1.332.103-2.777 0 0 .88-.282 2.88 1.075.835-.232 1.73-.348 2.62-.352.89.004 1.785.12 2.62.352 2-1.357 2.88-1.075 2.88-1.075.571 1.445.211 2.512.103 2.777.67.735 1.08 1.67 1.08 2.815 0 4.03-2.455 4.917-4.79 5.177.376.323.712.96.712 1.935 0 1.397-.013 2.522-.013 2.866 0 .281.19.61.725.507C19.494 20.562 22.5 16.635 22.5 12c0-5.799-4.701-10.5-10.5-10.5z"
        />
      </svg>
    )
  }
  return (
    <svg
      className="mr-2 h-5 w-5"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        fill="#EA4335"
      />
    </svg>
  )
}

function WeixinLogin({ redirect }: { redirect?: string }) {
  const { t } = useI18n()
  const reactId = useId()
  const containerId = `login_container_${reactId.replaceAll(":", "")}`
  const [message, setMessage] = useState("")

  useEffect(() => {
    let active = true

    async function init() {
      try {
        await loadScript(
          "https://res.wx.qq.com/connect/zh_CN/htmledition/js/wxLogin.js"
        )
        const data = await apiFetch<{
          appid?: string
          scope?: string
          redirect_uri?: string
          state?: string
        }>("/api/login/wx_login_config", {
          params: { redirect: redirect || "/" },
        })
        if (!active || !window.WxLogin) return
        new window.WxLogin({
          self_redirect: false,
          id: containerId,
          appid: data.appid,
          scope: data.scope,
          redirect_uri: data.redirect_uri,
          state: data.state || "",
        })
      } catch {
        if (active) setMessage(t("captcha.loadFailed"))
      }
    }

    void init()
    return () => {
      active = false
    }
  }, [containerId, redirect, t])

  return (
    <div className="weixin-qr-container">
      {message ? (
        <p className="text-sm text-muted-foreground">{message}</p>
      ) : null}
      <div
        id={containerId}
        className={cn("login_container", message && "hidden")}
      />
    </div>
  )
}

function WeixinLoginDialog({
  redirect,
  onClose,
}: {
  redirect?: string
  onClose: () => void
}) {
  const { t } = useI18n()

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4">
      <div className="w-[380px] max-w-[380px] rounded-lg bg-background p-5 shadow-lg">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">
            {t("user.signin.weixinLogin")}
          </h2>
          <button
            type="button"
            className="text-sm text-muted-foreground"
            onClick={onClose}
          >
            {t("dialog.cancel")}
          </button>
        </div>
        <div className="px-2 pb-4">
          <WeixinLogin redirect={redirect} />
        </div>
      </div>
    </div>
  )
}

function loadScript(src: string) {
  return new Promise<void>((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>(
      `script[src="${src}"]`
    )
    if (existing) {
      resolve()
      return
    }
    const script = document.createElement("script")
    script.src = src
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error("Failed to load sign-in script"))
    document.head.appendChild(script)
  })
}

declare global {
  interface Window {
    WxLogin?: new (options: Record<string, unknown>) => unknown
  }
}
