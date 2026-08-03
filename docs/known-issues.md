# 已知问题记录

> 本文件记录已确认但尚未修复的问题，避免后续重复排查。

## 1. 匿名访问内容页返回 500（登录墙 SSR 未适配）

**状态**：待修复（2026-08-03 记录，决策：保留登录墙，修复 SSR 降级）

### 现象

- 未登录访问 `/`（首页）、`/topic/:id`（话题详情）等**内容页**返回 HTTP 500
- 登录/注册/搜索页正常（200）
- API 层返回业务错误 `{"errorCode":1,"message":"请先登录","success":false}`（HTTP 200）

### 根因

`tsn_fork_pre` 分支实现了「权限隔离机制」（提交 `fd10c7d0`，2026-07-22）：

- `internal/server/router.go` 中 `topic/tag/comment/favorite/like/attachment/search/vote` 等路由组全部挂载了
  `middleware.ContentAccessMiddleware`（`internal/middleware/content_access_middleware.go`），形成**全站登录墙**：
  匿名用户无法访问任何内容类接口；`sitemap.xml` 同样受控；注册/重置密码接口被移除
- 配套的 `internal/services/content_access_service.go` 按用户授权分类（`GetAllowedCategoryIds`），
  匿名用户返回空权限
- **前端 SSR 未适配**：`web/app/route-helpers/loaders.ts` 等 loader 直接 `apiFetch`，
  未处理 `errs.NotLogin`（errorCode 1），react-router SSR loader 抛 `ApiError` → 整页 500

### 决策

**保留登录墙**（这是私有论坛的产品设计），修复前端 SSR 降级：

- 未登录访问内容页时，SSR 应渲染登录引导页（HTTP 200），而不是 500
- 登录后行为不变

### 涉及文件（修复时参考）

- `web/app/route-helpers/loaders.ts`：`loadTopicListRouteData`、`loadCategoryRouteData`、
  `loadTopicTagRouteData`、`loadTopicDetail` 等
- `web/app/routes/_index.tsx`、`web/app/routes/topic.$id.tsx` 等路由 loader
- 可参考 `web/lib/api/client.ts` 中 `handleInstallRequired` 抛 `Response` 的模式
  （对未安装返回 428），对未登录可类似处理或返回降级数据 + 登录引导 UI

### 备注

- 该问题随 2026-07-30 部署（提交 `212ad325` 之后）即存在，本次 2026-08-03 部署再次确认
- 验证命令（服务器上）：
  `docker exec bbs-go-bbs-go-1 wget -qO- 'http://127.0.0.1:8082/api/topic/topics?categoryId=0'`
  → 返回 `{"errorCode":1,"message":"请先登录","data":null,"success":false}`
