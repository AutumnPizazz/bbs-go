# Repository Guidelines

## 项目结构与模块组织

Go 应用从 `main.go` 启动。后端按职责组织在 `internal/` 下，包括 HTTP 处理器、服务、仓储、模型、中间件、权限和公共包。数据库迁移位于 `migrations/`，服务端翻译位于 `locales/`，部署文件位于 `deploy/` 和 `docker/`。React Router/Vite 前端位于 `web/`：路由模块在 `web/app/routes`，可复用 UI 和后台组件在 `web/components`，共享客户端逻辑在 `web/lib`，浏览器资源在 `web/public`。Go 测试与被测包放在同级目录，前端专项检查脚本位于 `web/scripts/test-*.mjs`。

## 构建、测试与开发命令

使用 `make web-install` 或 `cd web && corepack pnpm install --frozen-lockfile` 安装前端依赖。

- `make dev` 同时启动 Go 开发服务器和前端开发服务器。

- `make build` 构建 SPA，并将其嵌入 Go 二进制文件。

- `make test` 确认 SPA 资源存在后运行 `go test ./...`。

- `make check` 运行后端测试、前端类型检查和 ESLint。

- `cd web && corepack pnpm lint` 运行 ESLint；`corepack pnpm typecheck` 生成 React Router 类型并运行 `tsc --noEmit`。

- `cd web && corepack pnpm build:ssr` 或 `corepack pnpm build:spa` 构建对应的前端产物。

## 编码风格与命名约定

Go 代码使用 `gofmt` 格式化、制表符缩进，并遵循 Go 命名规范：导出标识符使用 `PascalCase`，私有标识符使用 `camelCase`。前端 TypeScript/TSX 使用两空格缩进、Prettier、ESLint 和现有 Tailwind 类名约定。组件使用 `PascalCase`，Hook 使用 `use...`，路由文件遵循 React Router 的命名方式。涉及用户界面的文案时，应同时维护 `web/lib/i18n/messages/en-US.ts` 和 `zh-CN.ts`。

## 测试约定

Go 测试文件使用 `*_test.go`，放在实现文件所在包内，并运行 `go test ./...`。前端改动应通过 `pnpm lint` 和 `pnpm typecheck`；涉及特定功能时运行对应脚本，例如 `node web/scripts/test-dashboard-routes.mjs`。仓库目前没有统一的覆盖率门槛。
