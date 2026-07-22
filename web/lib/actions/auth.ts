import { apiFetch, toFormData } from "@/lib/api/client"
import type { LoginResult } from "@/lib/api/types"
import { safeRedirect } from "@/lib/site"

export interface AuthActionState {
  ok: boolean
  message?: string
  redirect?: string
}

function formString(formData: FormData, key: string) {
  const value = formData.get(key)
  return typeof value === "string" ? value : ""
}

function optionalFormString(formData: FormData, key: string) {
  const value = formString(formData, key)
  return value || undefined
}

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error && error.message ? error.message : fallback
}

export async function signinAction(
  _state: AuthActionState,
  formData: FormData
): Promise<AuthActionState> {
  try {
    const result = await apiFetch<LoginResult>("/api/login/signin", {
      method: "POST",
      body: toFormData({
        username: formString(formData, "username"),
        password: formString(formData, "password"),
        redirect: optionalFormString(formData, "redirect"),
        captchaId: optionalFormString(formData, "captchaId"),
        captchaCode: optionalFormString(formData, "captchaCode"),
        captchaProtocol: optionalFormString(formData, "captchaProtocol"),
      }),
    })

    return {
      ok: true,
      redirect: safeRedirect(result.redirect, `/user/${result.user.id}`),
    }
  } catch (error) {
    return { ok: false, message: getErrorMessage(error, "Sign in failed") }
  }
}

export async function thirdPartySignin(
  provider: "github" | "google" | "weixin",
  code: string,
  state: string
) {
  const login =
    provider === "github"
      ? "/api/login/github_login_submit"
      : provider === "google"
        ? "/api/login/google_login_submit"
        : "/api/login/wx_login_submit"
  const result = await apiFetch<LoginResult>(login, {
    method: "POST",
    body: toFormData({ code, state }),
  })
  return safeRedirect(result.redirect, `/user/${result.user.id}`)
}

export async function googleOneTapSignin(
  credential: string,
  redirect?: string
) {
  const result = await apiFetch<LoginResult>("/api/login/google_one_tap", {
    method: "POST",
    body: toFormData({ credential }),
  })
  return safeRedirect(redirect || result.redirect, `/user/${result.user.id}`)
}

export async function signoutAction() {
  await apiFetch<void>("/api/login/signout")
}
