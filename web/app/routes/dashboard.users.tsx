"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardUsersRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "users"),
    description: dashboardData.desc(t, "users"),
    listEndpoint: "/api/admin/user/list",
    viewPermission: PERMISSIONS.DASHBOARD_USER_VIEW,
    detailEndpoint: (id) => `/api/admin/user/${id}`,
    createEndpoint: "/api/admin/user/create",
    createPermission: PERMISSIONS.DASHBOARD_USER_CREATE,
    updateEndpoint: "/api/admin/user/update",
    updatePermission: PERMISSIONS.DASHBOARD_USER_UPDATE,
    filters: [
      { name: "id", label: dashboardData.label(t, "id") },
      { name: "username", label: dashboardData.label(t, "username") },
      { name: "nickname", label: dashboardData.label(t, "nickname") },
      {
        name: "forbidden",
        label: dashboardData.label(t, "forbidden"),
        type: "select",
        options: [
          { label: t("dashboard.boolean.yes"), value: "true" },
          { label: t("dashboard.boolean.no"), value: "false" },
        ],
      },
    ],
    columns: [
      {
        key: "id",
        label: dashboardData.label(t, "id"),
        render: (record) => dashboardData.userLinkCell(record, record.id),
      },
      {
        key: "idEncode",
        label: dashboardData.label(t, "idEncode"),
        render: (record) => dashboardData.userLinkCell(record, record.idEncode),
      },
      {
        key: "avatar",
        label: dashboardData.label(t, "avatar"),
        render: (record) =>
          dashboardData.imageCell(
            record.avatar || record.smallAvatar,
            String(record.nickname || record.username || "")
          ),
      },
      {
        key: "username",
        label: dashboardData.label(t, "username"),
        render: (record) => dashboardData.userLinkCell(record, record.username),
      },
      {
        key: "nickname",
        label: dashboardData.label(t, "nickname"),
        render: (record) => dashboardData.userLinkCell(record, record.nickname),
      },
      { key: "phone", label: dashboardData.label(t, "phone") },
      {
        key: "contentAccessMode",
        label: dashboardData.label(t, "contentAccessMode"),
        render: (record) =>
          record.contentAccessMode === "all"
            ? t("dashboard.userAccess.all")
            : t("dashboard.userAccess.assignedCategories"),
      },
      {
        key: "forbidden",
        label: dashboardData.label(t, "forbidden"),
        render: (record) =>
          record.forbidden
            ? t("dashboard.boolean.yes")
            : t("dashboard.boolean.no"),
      },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      {
        name: "username",
        label: dashboardData.label(t, "username"),
      },
      { name: "phone", label: dashboardData.label(t, "phone") },
      {
        name: "nickname",
        label: dashboardData.label(t, "nickname"),
        required: true,
      },
      {
        name: "avatar",
        label: dashboardData.label(t, "avatar"),
        type: "image",
      },
      {
        name: "gender",
        label: dashboardData.label(t, "gender"),
        type: "select",
        options: [
          { label: t("dashboard.gender.male"), value: "Male" },
          { label: t("dashboard.gender.female"), value: "Female" },
        ],
      },
      { name: "homePage", label: dashboardData.label(t, "homePage") },
      {
        name: "description",
        label: dashboardData.label(t, "description"),
        type: "textarea",
      },
      {
        name: "roleIds",
        label: dashboardData.label(t, "roles"),
        type: "multiselect",
        optionsEndpoint: "/api/admin/role/roles",
        optionLabel: (record) =>
          String(record.name || record.code || record.id),
        optionValue: (record) => record.id as number,
        valueFromRecord: (record) =>
          Array.isArray(record.roleIds)
            ? record.roleIds.map((item) => String(item))
            : [],
      },
      {
        name: "password",
        label: dashboardData.label(t, "password"),
        type: "password",
      },
      {
        name: "contentAccessMode",
        label: dashboardData.label(t, "contentAccessMode"),
        type: "select",
        required: true,
        options: [
          { label: t("dashboard.userAccess.all"), value: "all" },
          {
            label: t("dashboard.userAccess.assignedCategories"),
            value: "assigned_categories",
          },
        ],
      },
      {
        name: "categoryIds",
        label: dashboardData.label(t, "categoryIds"),
        type: "multiselect",
        optionsEndpoint: "/api/admin/category/options",
        optionLabel: dashboardData.treeOptionLabel,
        optionValue: (record) => record.id as number,
        valueFromRecord: (record) =>
          Array.isArray(record.categoryIds)
            ? record.categoryIds.map((item) => String(item))
            : [],
      },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        required: true,
        options: dashboardData.normalDeletedOptions(t),
      },
    ],
    rowActions: [
      {
        label: t("dashboard.actions.resetPassword"),
        endpoint: "/api/admin/user/reset_password",
        permission: PERMISSIONS.DASHBOARD_USER_RESET_PASSWORD,
        payload: (record) => ({ userId: record.id as number }),
        confirm: t("dashboard.confirmResetPassword"),
      },
    ],
  }

  return <DashboardDataPage config={config} />
}
