"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardMessagesRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "messages"),
    description: dashboardData.desc(t, "messages"),
    listEndpoint: "/api/admin/message/list",
    viewPermission: PERMISSIONS.DASHBOARD_MESSAGE_VIEW,
    detailEndpoint: (id) => `/api/admin/message/${id}`,
    createEndpoint: "/api/admin/message/create",
    createPermission: PERMISSIONS.DASHBOARD_MESSAGE_SEND,
    updateEndpoint: "/api/admin/message/update",
    updatePermission: PERMISSIONS.DASHBOARD_MESSAGE_SEND,
    deleteEndpoint: "/api/admin/message/delete",
    deletePermission: PERMISSIONS.DASHBOARD_MESSAGE_DELETE,
    deleteMode: "formIds",
    filters: [
      { name: "userId", label: dashboardData.label(t, "userId") },
      { name: "type", label: dashboardData.label(t, "type") },
      { name: "title", label: dashboardData.label(t, "title") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: [
          { label: dashboardData.label(t, "unread"), value: 0 },
          { label: dashboardData.label(t, "read"), value: 1 },
          { label: dashboardData.label(t, "deleted"), value: 2 },
        ],
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "title", label: dashboardData.label(t, "title"), className: "min-w-56" },
      { key: "type", label: dashboardData.label(t, "type") },
      { key: "status", label: dashboardData.label(t, "status") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    detailFields: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "fromId", label: dashboardData.label(t, "fromId") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "title", label: dashboardData.label(t, "title") },
      { key: "content", label: dashboardData.label(t, "content") },
      { key: "type", label: dashboardData.label(t, "type") },
      { key: "status", label: dashboardData.label(t, "status") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      {
        name: "userId",
        label: dashboardData.label(t, "userId"),
        required: true,
        type: "number",
        min: 1,
      },
      { name: "title", label: dashboardData.label(t, "title"), required: true },
      {
        name: "content",
        label: dashboardData.label(t, "content"),
        required: true,
        type: "textarea",
        colSpan: 2,
      },
    ],
    canDelete: (record) => record.status !== 2,
  }

  return <DashboardDataPage config={config} />
}
