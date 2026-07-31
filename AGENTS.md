# Repository Guidelines

## 项目概览与模块边界

- Go 程序从 `main.go` 启动：`server.Init()` 完成配置、日志、翻译、数据库和迁移初始化，随后 `server.NewServer()` 注册 Gin 路由。
- 后端代码位于 `internal/`：`handlers/api`、`handlers/admin`、`handlers/render` 负责 HTTP 层；`services` 负责业务逻辑；`repositories` 负责数据访问；`models` 存放领域模型和 DTO；`middleware`、`permissions`、`cache`、`scheduler`、`spam` 和 `pkg` 提供横向能力。
- API 边界主要在 `internal/server/router.go`：公共接口位于 `/api`，管理接口位于 `/api/admin`；管理接口还经过管理员权限和 CSRF 中间件。修改接口时同时检查对应的前端 API 调用和路由测试。
- 数据库初始化和升级由 `internal/install/` 与 `migrations/` 负责。迁移是 Go 函数注册机制，不是独立 SQL 文件：新增迁移应创建编号文件，并在 `migrations/migration.go` 的 `init()` 中用严格递增的版本号调用 `register`。
- 服务端翻译在 `locales/en-US.yml` 和 `locales/zh-CN.yml`；前端翻译在 `web/lib/i18n/messages/en-US.ts` 和 `zh-CN.ts`。新增用户可见文案时两端涉及的语言文件都要同步维护。
- `cmd/` 存放辅助程序和生成器，包括 `generator`、`package-release`、密码检查/重置等工具；不要把一次性工具逻辑混入业务包。
- React Router/Vite 前端位于 `web/`：`web/app/routes` 使用文件系统路由，路由配置在 `web/app/routes.ts`；共享组件在 `web/components`，客户端/服务端共享逻辑在 `web/lib`，静态资源在 `web/public`。
- 非 `dev` 的 Go 构建通过 `web/embed.go` 嵌入 `web/build/spa`；`dev` 构建通过 `web/embed_dev.go` 直接读取 `web/`。SSR 入口是 `web/scripts/serve-ssr.mjs`，会将 `/api/`、`/res/` 和 `/sitemap.xml` 代理到 `BBSGO_SERVER_URL`。
- 部署相关文件集中在 `Dockerfile`、根目录 `docker-compose.yml`、`deploy/` 和 `docker/`；文档和示例配置分别位于 `docs/`、`README*.md` 和 `bbs-go.example.yaml`。根目录 `tests/`、`tools/` 当前没有可运行的测试入口，测试以各包内文件和 `web/scripts/test-*.mjs` 为主。

## 环境与依赖

- Go 版本按 `go.mod` 使用 `go 1.26.0`；容器构建使用 Go 1.26 和 Node 24。
- 前端包管理器固定为 `pnpm@10.30.2`（见 `web/package.json`）。优先在 `web/` 目录使用 `corepack pnpm`，不要混用未锁定版本的全局 `pnpm`。
- 安装前端依赖：

  ```bash
  cd web && corepack pnpm install --frozen-lockfile
  ```

- Vite 和 SSR 构建要求 `BBSGO_SERVER_URL`。本地开发通常在 `web/.env` 中设置为 `http://localhost:8082`；改动端口或后端地址时同步修改该变量。
- `Makefile` 的命令使用 Unix shell 语法。在 Windows 且没有可用的 `make` 时，使用下方列出的直接命令，或在已配置 POSIX shell 的环境中执行 `make`。

## 构建、测试与开发命令

根目录 `Makefile` 是命令编排的权威来源：

- `make dev`：以 `-tags dev` 启动 Go 服务，并启动 Vite 开发服务器。Vite 默认使用 `3000`，Go 默认使用 `8082`。没有 `make` 时，在两个终端分别运行：
  `go run -tags dev ./main.go` 和 `cd web && corepack pnpm dev`。
- `make build`：构建 SPA 并将其嵌入 Go 二进制。直接执行 Go 的非 `dev` 构建前，必须先执行 `cd web && corepack pnpm build:spa`，否则 `web/build/spa` 不存在时会导致 `embed` 构建失败。
- `make run` / `make run-go`：构建或确保 SPA 产物存在后运行 Go 服务。
- `make test`：先通过 `ensure-spa` 确保 SPA 产物存在，再运行 `go test ./...`。
- `make check`：运行后端测试、前端类型检查和 ESLint。等价的分步命令为 `go test ./...`、`cd web && corepack pnpm typecheck`、`cd web && corepack pnpm lint`；直接运行 `go test ./...` 前仍需先构建 SPA。
- `make web-build-spa` / `make web-build-ssr`：分别生成嵌入用 SPA 和 SSR 构建产物。SSR 本地运行顺序为 `cd web && corepack pnpm build:ssr`，然后 `corepack pnpm start`。
- `make web-install`、`make web-dev`、`make web-typecheck`、`make web-lint` 分别对应前端依赖安装、开发服务器、类型检查和 ESLint。
- 前端专项断言脚本不由统一测试 runner 托管，按需单独执行，例如 `node web/scripts/test-dashboard-routes.mjs`。涉及某个功能时，搜索并运行对应的 `web/scripts/test-*.mjs`。
- 生成辅助代码：`make generator` 运行 `cmd/generator/generator.go`；`make generate-permissions` 运行 `cmd/generator/permissions` 并更新 `web/lib/auth/permissions.generated.ts`。生成文件不要手工编辑，应修改源定义后重新生成。
- 提交前至少运行与改动范围匹配的测试；涉及后端和前端边界的改动运行 `make check`，并补跑相关的前端专项脚本。

## 编码风格与命名约定

- Go 使用 `gofmt`，保持制表符缩进；导出标识符使用 `PascalCase`，私有标识符使用 `camelCase`。Go 测试文件命名为 `*_test.go`，放在被测包同级，测试包按现有包结构组织。
- 前端 TypeScript/TSX 使用两空格缩进、无分号、双引号和 80 列宽等仓库现有 Prettier 规则；格式化使用 `cd web && corepack pnpm format`。组件使用 `PascalCase`，Hook 使用 `use...`，路由文件遵循 React Router 的命名规则。
- 优先复用现有的 `web/components/ui`、`web/lib` 和后端公共包，不要在业务路由中重复实现通用 API、权限、分页或本地化逻辑。
- 修改权限、接口、数据库字段或配置结构时，同时检查生成文件、迁移、前后端类型/API 契约和相关测试。

## 验证约定

- Go 改动运行 `go test ./...`；需要格式化的文件运行 `gofmt`。仓库没有统一的覆盖率门槛，也没有在 CI 中自动运行 `make check` 的工作流。
- 前端改动至少运行 `cd web && corepack pnpm lint` 和 `corepack pnpm typecheck`；`typecheck` 会先生成 React Router 类型。涉及特定行为时，再运行对应的 `node web/scripts/test-*.mjs`。
- 只修改文档或配置时仍应检查命令、路径和示例是否与 `Makefile`、`package.json`、`go.mod` 及实际目录一致。
- 当前 GitHub Actions 工作流 `.github/workflows/docker-image.yml` 仅在推送 `v*` 标签时构建并推送 Docker 镜像，因此本地验证不能以 CI 会自动执行测试为前提。
