"use client"

import * as React from "react"
import {
  LayoutDashboardIcon,
  MessageSquareIcon,
  Settings2Icon,
  UsersIcon,
} from "lucide-react"

import { NavMain } from "@/components/dashboard/nav-main"
import { NavUser } from "@/components/dashboard/nav-user"
import { SidebarBrand } from "@/components/dashboard/sidebar-brand"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"
import { useCurrentUser } from "@/components/app/app-provider"
import { userHasPermission } from "@/lib/auth/roles"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const currentUser = useCurrentUser()
  const { t } = useI18n()

  const data = {
    user: {
      name:
        currentUser?.nickname ||
        currentUser?.username ||
        t("dashboard.user.anonymous"),
      avatar: currentUser?.smallAvatar || currentUser?.avatar || "",
    },
    brand: {
      name: t("dashboard.brand.name"),
      logoSrc: "/logo.png",
      description: t("dashboard.brand.plan"),
    },
    navMain: [
      {
        title: t("dashboard.nav.workspace"),
        url: "/dashboard",
        icon: LayoutDashboardIcon,
        permission: PERMISSIONS.DASHBOARD_VIEW,
      },
      {
        title: t("dashboard.nav.content"),
        url: "/dashboard/content",
        icon: MessageSquareIcon,
        permission: PERMISSIONS.DASHBOARD_TOPIC_VIEW,
        items: [
          {
            title: t("dashboard.nav.topics"),
            url: "/dashboard/topics",
            permission: PERMISSIONS.DASHBOARD_TOPIC_VIEW,
          },
          {
            title: t("dashboard.nav.categories"),
            url: "/dashboard/categories",
            permission: PERMISSIONS.DASHBOARD_CATEGORY_VIEW,
          },
          {
            title: t("dashboard.nav.links"),
            url: "/dashboard/links",
            permission: PERMISSIONS.DASHBOARD_LINK_VIEW,
          },
          {
            title: t("dashboard.nav.comments"),
            url: "/dashboard/comments",
            permission: PERMISSIONS.DASHBOARD_COMMENT_VIEW,
          },
          {
            title: t("dashboard.nav.tags"),
            url: "/dashboard/tags",
            permission: PERMISSIONS.DASHBOARD_TAG_VIEW,
          },
          {
            title: t("dashboard.nav.attachments"),
            url: "/dashboard/attachments",
            permission: PERMISSIONS.DASHBOARD_ATTACHMENT_VIEW,
          },
          {
            title: t("dashboard.nav.votes"),
            url: "/dashboard/votes",
            permission: PERMISSIONS.DASHBOARD_VOTE_VIEW,
            items: [
              {
                title: t("dashboard.nav.votes"),
                url: "/dashboard/votes",
                permission: PERMISSIONS.DASHBOARD_VOTE_VIEW,
              },
              {
                title: t("dashboard.nav.voteOptions"),
                url: "/dashboard/vote-options",
                permission: PERMISSIONS.DASHBOARD_VOTE_OPTION_VIEW,
              },
            ],
          },
        ],
      },
      {
        title: t("dashboard.nav.community"),
        url: "/dashboard/users",
        icon: UsersIcon,
        permission: PERMISSIONS.DASHBOARD_USER_VIEW,
        items: [
          {
            title: t("dashboard.nav.userList"),
            url: "/dashboard/users",
            permission: PERMISSIONS.DASHBOARD_USER_VIEW,
          },
          {
            title: t("dashboard.nav.userReports"),
            url: "/dashboard/user-reports",
            permission: PERMISSIONS.DASHBOARD_USER_REPORT_VIEW,
          },
          {
            title: t("dashboard.nav.favorites"),
            url: "/dashboard/favorites",
            permission: PERMISSIONS.DASHBOARD_FAVORITE_VIEW,
          },
          {
            title: t("dashboard.nav.messages"),
            url: "/dashboard/messages",
            permission: PERMISSIONS.DASHBOARD_MESSAGE_VIEW,
          },
        ],
      },
      {
        title: t("dashboard.nav.system"),
        url: "/dashboard/settings",
        icon: Settings2Icon,
        permission: PERMISSIONS.DASHBOARD_SETTING_VIEW,
        items: [
          {
            title: t("dashboard.nav.siteSettings"),
            url: "/dashboard/settings",
            permission: PERMISSIONS.DASHBOARD_SETTING_VIEW,
          },
          {
            title: t("dashboard.nav.roles"),
            url: "/dashboard/roles",
            permission: PERMISSIONS.DASHBOARD_ROLE_VIEW,
          },
          {
            title: t("dashboard.nav.operateLogs"),
            url: "/dashboard/operate-logs",
            permission: PERMISSIONS.DASHBOARD_OPERATE_LOG_VIEW,
          },
          {
            title: t("dashboard.nav.dictionaries"),
            url: "/dashboard/dictionaries",
            permission: PERMISSIONS.DASHBOARD_DICT_TYPE_VIEW,
            items: [
              {
                title: t("dashboard.nav.dictionaryTypes"),
                url: "/dashboard/dictionaries",
                permission: PERMISSIONS.DASHBOARD_DICT_TYPE_VIEW,
              },
              {
                title: t("dashboard.nav.dictionaryItems"),
                url: "/dashboard/dictionary-items",
                permission: PERMISSIONS.DASHBOARD_DICT_VIEW,
              },
            ],
          },
          {
            title: t("dashboard.nav.voteRecords"),
            url: "/dashboard/vote-records",
            permission: PERMISSIONS.DASHBOARD_VOTE_RECORD_VIEW,
          },
          {
            title: t("dashboard.nav.health"),
            url: "/dashboard/health",
            permission: PERMISSIONS.DASHBOARD_HEALTH_VIEW,
          },
          {
            title: t("dashboard.nav.trash"),
            url: "/dashboard/trash",
            permission: PERMISSIONS.DASHBOARD_TOPIC_DELETE,
          },
        ],
      },
    ],
  }

  const navMain = data.navMain
    .map((item) => {
      const children = item.items?.filter((child) =>
        userHasPermission(currentUser, child.permission)
      )
      const visible =
        userHasPermission(currentUser, item.permission) ||
        Boolean(children?.length)
      if (!visible) return null
      return {
        ...item,
        items: children,
      }
    })
    .filter((item): item is NonNullable<typeof item> => Boolean(item))

  return (
    <Sidebar
      collapsible="icon"
      className="border-r bg-[var(--dashboard-panel)]"
      {...props}
    >
      <SidebarHeader>
        <SidebarBrand {...data.brand} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navMain} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
