"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardOperateLogsRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "operateLogs"),
    description: dashboardData.desc(t, "operateLogs"),
    listEndpoint: "/api/admin/operate-log/list",
    viewPermission: PERMISSIONS.DASHBOARD_OPERATE_LOG_VIEW,
    detailEndpoint: (id) => `/api/admin/operate-log/${id}`,
    filters: [
      { name: "userId", label: dashboardData.label(t, "userId") },
      { name: "opType", label: dashboardData.label(t, "opType") },
      { name: "dataType", label: dashboardData.label(t, "dataType") },
      { name: "dataId", label: dashboardData.label(t, "dataId") },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "opType", label: dashboardData.label(t, "opType") },
      { key: "dataType", label: dashboardData.label(t, "dataType") },
      { key: "dataId", label: dashboardData.label(t, "dataId") },
      {
        key: "description",
        label: dashboardData.label(t, "description"),
        className: "min-w-80",
      },
      { key: "ip", label: dashboardData.label(t, "ip") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    detailFields: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "opType", label: dashboardData.label(t, "opType") },
      { key: "dataType", label: dashboardData.label(t, "dataType") },
      { key: "dataId", label: dashboardData.label(t, "dataId") },
      { key: "description", label: dashboardData.label(t, "description") },
      { key: "ip", label: dashboardData.label(t, "ip") },
      { key: "userAgent", label: dashboardData.label(t, "userAgent") },
      { key: "referer", label: dashboardData.label(t, "referer") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
  }

  return <DashboardDataPage config={config} />
}
