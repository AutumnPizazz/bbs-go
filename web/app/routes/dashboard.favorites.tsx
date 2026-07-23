"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardFavoritesRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "favorites"),
    description: dashboardData.desc(t, "favorites"),
    listEndpoint: "/api/admin/favorite/list",
    viewPermission: PERMISSIONS.DASHBOARD_FAVORITE_VIEW,
    detailEndpoint: (id) => `/api/admin/favorite/${id}`,
    filters: [
      { name: "userId", label: dashboardData.label(t, "userId") },
      { name: "entityType", label: dashboardData.label(t, "entityType") },
      { name: "entityId", label: dashboardData.label(t, "entityId") },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "entityType", label: dashboardData.label(t, "entityType") },
      { key: "entityId", label: dashboardData.label(t, "entityId") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    detailFields: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "entityType", label: dashboardData.label(t, "entityType") },
      { key: "entityId", label: dashboardData.label(t, "entityId") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
  }

  return <DashboardDataPage config={config} />
}
