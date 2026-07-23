"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardCommentsRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "comments"),
    description: dashboardData.desc(t, "comments"),
    listEndpoint: "/api/admin/comment/list",
    viewPermission: PERMISSIONS.DASHBOARD_COMMENT_VIEW,
    detailEndpoint: (id) => `/api/admin/comment/${id}`,
    deleteEndpoint: "/api/admin/comment/delete",
    deletePermission: PERMISSIONS.DASHBOARD_COMMENT_DELETE,
    filters: [
      { name: "id", label: dashboardData.label(t, "id") },
      { name: "userId", label: dashboardData.label(t, "userId") },
      { name: "entityType", label: dashboardData.label(t, "entityType") },
      { name: "entityId", label: dashboardData.label(t, "entityId") },
      { name: "content", label: dashboardData.label(t, "content") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: dashboardData.normalDeletedOptions(t),
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "entityType", label: dashboardData.label(t, "entityType") },
      { key: "entityId", label: dashboardData.label(t, "entityId") },
      {
        key: "content",
        label: dashboardData.label(t, "content"),
        className: "min-w-96",
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
    detailFields: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "entityType", label: dashboardData.label(t, "entityType") },
      { key: "entityId", label: dashboardData.label(t, "entityId") },
      { key: "quoteId", label: dashboardData.label(t, "quoteId") },
      {
        key: "content",
        label: dashboardData.label(t, "content"),
        render: (record) => dashboardData.codeBlock(record.content),
      },
      {
        key: "topic",
        label: dashboardData.label(t, "topic"),
        render: (record) => dashboardData.codeBlock(record.topic),
      },
      {
        key: "parent",
        label: dashboardData.label(t, "parentComment"),
        render: (record) => dashboardData.codeBlock(record.parent),
      },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
  }

  return <DashboardDataPage config={config} />
}
