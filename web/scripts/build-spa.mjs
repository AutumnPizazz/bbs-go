import { spawn } from "node:child_process"
import path from "node:path"
import { fileURLToPath } from "node:url"

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..")
const isWindows = process.platform === "win32"
const reactRouterCommand = isWindows ? process.env.ComSpec || "cmd.exe" : "react-router"
const reactRouterArgs = isWindows ? ["/d", "/s", "/c", "react-router.cmd build"] : ["build"]

await new Promise((resolve, reject) => {
  const child = spawn(reactRouterCommand, reactRouterArgs, {
    cwd: root,
    env: { ...process.env, BBSGO_WEB_SPA: "true" },
    stdio: "inherit",
    shell: false,
  })

  child.once("error", reject)
  child.once("exit", (code, signal) => {
    if (code === 0) {
      resolve()
      return
    }
    reject(new Error(`react-router build failed (${signal || `exit code ${code}`})`))
  })
})

await import("./prepare-spa-build.mjs")
