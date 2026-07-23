import { AUTH_COOKIE } from "@/lib/cookies"

import { apiFetch, type ApiRequestOptions } from "./client"

function readRequestToken(request?: Request) {
  const cookieHeader = request?.headers.get("cookie")
  if (!cookieHeader) return undefined

  for (const part of cookieHeader.split(";")) {
    const separator = part.indexOf("=")
    if (separator < 0) continue
    const name = part.slice(0, separator).trim()
    if (name !== AUTH_COOKIE) continue
    return decodeURIComponent(part.slice(separator + 1).trim())
  }

  return undefined
}

export async function serverApiFetch<T>(
  path: string,
  options: ApiRequestOptions = {}
) {
  const token = options.token ?? readRequestToken(options.request)
  return apiFetch<T>(path, { ...options, token })
}
