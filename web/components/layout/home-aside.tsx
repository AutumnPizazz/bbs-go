"use client"

import * as React from "react"
import Link from "@/components/common/link"

import { useAppState } from "@/components/app/app-provider"
import { apiFetch } from "@/lib/api/client"
import type { FriendLink } from "@/lib/api/misc"
import { useI18n } from "@/lib/i18n/provider"

function WidgetCard({
  title,
  actions,
  children,
}: {
  title?: string
  actions?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <section className="rounded-md bg-background px-3 py-1">
      {title || actions ? (
        <div className="flex items-center justify-between border-b py-2 text-base font-medium">
          <span>{title}</span>
          {actions ? (
            <div className="shrink-0 text-sm font-normal">{actions}</div>
          ) : null}
        </div>
      ) : null}
      <div className="py-2 break-all">{children}</div>
    </section>
  )
}

function SiteNotice({ title, content }: { title: string; content?: string }) {
  if (!content) return null

  return (
    <WidgetCard title={title}>
      <div
        className="prose prose-sm max-w-none text-sm text-muted-foreground"
        dangerouslySetInnerHTML={{ __html: content }}
      />
    </WidgetCard>
  )
}

function FriendLinks({
  title,
  more,
  links,
}: {
  title: string
  more: string
  links: FriendLink[]
}) {
  if (!links.length) return null

  return (
    <WidgetCard
      title={title}
      actions={
        <Link
          href="/links"
          className="text-muted-foreground hover:text-primary"
        >
          {more}
        </Link>
      }
    >
      <ul className="links">
        {links.map((link) => (
          <li key={link.id} className="link">
            <a
              href={link.url || "#"}
              title={link.title}
              className="link-title"
              target="_blank"
              rel="noreferrer"
            >
              {link.title}
            </a>
            {link.summary ? (
              <p className="link-summary">{link.summary}</p>
            ) : null}
          </li>
        ))}
      </ul>
    </WidgetCard>
  )
}

export function HomeAside() {
  const { config } = useAppState()
  const { t } = useI18n()
  const [friendLinks, setFriendLinks] = React.useState<FriendLink[]>([])

  React.useEffect(() => {
    let mounted = true
    void apiFetch<FriendLink[]>("/api/link/top_links")
      .catch(() => [])
      .then((nextLinks) => {
        if (mounted) setFriendLinks(Array.isArray(nextLinks) ? nextLinks : [])
      })

    return () => {
      mounted = false
    }
  }, [])

  return (
    <>
      <SiteNotice
        title={t("component.siteNotice.title")}
        content={config?.siteNotification}
      />
      <FriendLinks
        title={t("component.friendLinks.title")}
        more={t("component.friendLinks.more")}
        links={friendLinks}
      />
    </>
  )
}
