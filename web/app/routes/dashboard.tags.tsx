"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardTagsRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "tags"),
    description: dashboardData.desc(t, "tags"),
    listEndpoint: "/api/admin/tag/list",
    viewPermission: PERMISSIONS.DASHBOARD_TAG_VIEW,
    detailEndpoint: (id) => `/api/admin/tag/${id}`,
    createEndpoint: "/api/admin/tag/create",
    createPermission: PERMISSIONS.DASHBOARD_TAG_CREATE,
    updateEndpoint: "/api/admin/tag/update",
    updatePermission: PERMISSIONS.DASHBOARD_TAG_UPDATE,
    filters: [
      { name: "id", label: dashboardData.label(t, "id") },
      { name: "name", label: dashboardData.label(t, "name") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: dashboardData.normalDeletedOptions(t),
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "name", label: dashboardData.label(t, "name") },
      {
        key: "description",
        label: dashboardData.label(t, "description"),
        className: "min-w-72",
      },
      {
        key: "status",
        label: dashboardData.label(t, "status"),
        render: (record) => dashboardData.statusCell(t, record.status),
      },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      { name: "id", label: dashboardData.label(t, "id"), type: "number" },
      { name: "name", label: dashboardData.label(t, "name"), required: true },
      {
        name: "description",
        label: dashboardData.label(t, "description"),
        type: "textarea",
      },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        required: true,
        options: dashboardData.normalDeletedOptions(t),
      },
    ],
  }

  return <DashboardDataPage config={config} />
}
