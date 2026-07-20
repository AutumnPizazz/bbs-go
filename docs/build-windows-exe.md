# Windows EXE 构建与故障复盘

本文记录在 Windows 上构建 BBS-GO 可执行文件时的关键步骤，以及一次前端静态资源构建错误的复盘。

## 正确的构建前提

- Go 版本必须满足 `go.mod` 中的 `go 1.26.0`。
- 前端使用 `pnpm@10.30.2`，通过 Node.js Corepack 安装依赖。
- Go 服务通过 `//go:embed build/spa` 嵌入前端产物，因此 `web/build/spa/index.html` 必须存在。
- Windows 运行时配置文件名是 `bbs-go.yaml`，缺失时会使用默认配置；首次启动会进入安装向导。

## 这次犯过的错误

### 1. 没有先确认本机工具链

Windows 环境最初没有 `go` 命令，直接构建只能失败。以后应先执行：

```powershell
go version
node --version
corepack pnpm --version
```

### 2. 把 Unix 环境变量写法直接用于 Windows

`package.json` 中的脚本包含：

```text
BBSGO_WEB_SPA=true react-router build
```

这在 Windows CMD 中会被当成命令，报错“`BBSGO_WEB_SPA` 不是命令”。PowerShell 的等价写法是：

```powershell
$env:BBSGO_WEB_SPA = "true"
pnpm exec react-router build
```

长期维护时应让项目脚本使用跨平台环境变量工具，或提供 Windows 专用脚本。

### 3. 把 SSR 的 `build/client` 直接复制成 SPA 产物

SSR 构建可以生成大量静态 assets，但不一定生成可独立运行的 `index.html`。只复制 `build/client` 会导致 Go 服务找不到首页，访问 `/` 返回 404。

### 4. 用 SSR fallback 冒充 SPA fallback

把 SSR 请求生成的 HTML 直接嵌入后，页面可能看起来正常，但 React Router 的 context 仍然是：

```json
{"ssr": true, "isSpaMode": false}
```

Go 服务只提供静态文件，无法继续提供 SSR hydration 所需的服务端上下文，结果就是页面能显示、JavaScript 资源也能加载，但按钮没有事件响应。

### 5. 只嵌入当前页面的局部 route manifest

从 `/` 生成的 HTML 只会携带当前匹配的部分 routes。缺少完整 manifest 时，点击进入 `/install` 等页面会落到 React Router 的 404 页面，即使对应的 JS 文件实际已经存在。

## 正确的前端产物流程

1. 安装依赖：

   ```powershell
   cd web
   corepack pnpm install --frozen-lockfile
   ```

2. 生成 client/server 构建：

   ```powershell
   Remove-Item Env:BBSGO_WEB_SPA -ErrorAction SilentlyContinue
   corepack pnpm exec react-router build
   ```

3. 使用 React Router 的 `createRequestHandler` 生成真正的 SPA fallback。构建对象必须使用：

   ```text
   ssr: false
   isSpaMode: true
   ```

   生成时需要设置 `BBSGO_SERVER_URL`，因为部分 route loader 会请求 BBS-GO API。API 服务应临时启动，生成完成后再停止。

4. 将完整的 `build/client/assets/manifest-*.js` 保留在 `build/spa/assets` 中。首页 HTML 必须同时满足：

   - 包含 `"ssr":false`
   - 包含 `"isSpaMode":true`
   - 包含 `window.__reactRouterContext.streamController.close()`
   - 包含 `entry.client-*.js`
   - 通过 manifest 提供全部 routes，包括 `routes/install`

5. 最后编译 Go：

   ```powershell
   $env:GOPROXY = "https://goproxy.cn,direct"
   go build -trimpath -ldflags="-s -w" -o bbs-go-windows-amd64.exe ./main.go
   ```

## 发布前检查

- 关闭旧的 exe 进程，避免误测旧产物。
- 删除或重新加载浏览器缓存，使用 `Ctrl + Shift + R`。
- 访问 `/` 后确认浏览器请求了 `entry.client-*.js`。
- 访问 `/install`，确认请求了 `assets/install-*.js`，且页面不是 React Router 404。
- 检查 API 请求能返回 200。
- 执行服务器静态资源测试：

  ```powershell
  go test ./internal/server ./internal/pkg/ginx
  ```

- 计算 exe 的 SHA-256，记录最终交付文件的校验值。

## 结论

Go 编译成功不代表嵌入的前端应用可用。BBS-GO 的 Windows 单文件产物必须同时验证 Go 二进制、SPA fallback、完整 route manifest、客户端 hydration 和安装流程。
