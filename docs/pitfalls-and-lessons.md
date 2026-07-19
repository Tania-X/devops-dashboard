# Sprint 1 踩坑记录与经验总结

> 记录从纯前端（MSW Mock）演进到 Go + Gin + SQLite 真实后端过程中遇到的所有问题，以及对应的解决方案。供后续 Sprint 参考。

---

## 1. Go 模块下载超时

**症状**
```text
go mod tidy
go: finding module for package github.com/gin-gonic/gin
go: module github.com/gin-gonic/gin: Get "https://proxy.golang.org/...": dial tcp ... connectex: A connection attempt failed...
```

**原因**
Go 默认使用 `proxy.golang.org` 作为模块代理，该服务在国内极不稳定，即使挂了 VPN 也可能无法连通（Go 的 HTTP Client 不自动走系统代理）。

**解决**
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off
```
`goproxy.cn` 是七牛云维护的国内镜像，一劳永逸，所有项目通用。

---

## 2. C 盘空间不足导致 Go 编译失败

**症状**
```text
compile: writing output: write $WORK\b160\_pkg_.a: There is not enough space on the disk.
```

**原因**
Go 编译时会将大量中间文件写入系统临时目录（`C:\Users\...\AppData\Local\Temp`），C 盘空间不足时直接编译失败。

**解决**
将 Go 编译临时目录改到空间充足的 D 盘：
```bash
mkdir D:\temp
go env -w GOTMPDIR=D:\temp
```

---

## 2.5 Go 目录全面搬离 C 盘（模块缓存 / 编译缓存 / 二进制安装）

**症状（连锁反应）**
1. `go install github.com/go-delve/delve/cmd/dlv@latest` 报错：`There is not enough space on the disk`（模块下载到 C 盘）
2. Delve 装好后，VSCode 按 `F5` Debug 再次报错：`write C:\Users\...\AppData\Local\go-build\...: There is not enough space on the disk`（编译缓存写到 C 盘）

**原因**
Go 默认会把以下三类文件全部写到 C 盘（Windows 默认 GOPATH 在 `C:\Users\<用户>\go`）：
- **模块下载缓存**（`GOMODCACHE`）：`go install` / `go mod tidy` 时下载的依赖包
- **编译缓存**（`GOCACHE`）：`go build` 时产生的中间产物，位于 `AppData\Local\go-build`
- **二进制安装目录**（`GOBIN`）：`go install` 安装的可执行文件（如 `dlv.exe`）

**解决**
逐一改到 D 盘，并确保目录存在：

```powershell
# 模块下载缓存
mkdir D:\tools\go-mod-cache -ErrorAction SilentlyContinue
go env -w GOMODCACHE=D:\tools\go-mod-cache

# 编译缓存
mkdir D:\tools\go-cache -ErrorAction SilentlyContinue
go env -w GOCACHE=D:\tools\go-cache

# go install 安装的二进制（dlv.exe 等）
mkdir D:\tools\go-workspace\bin -ErrorAction SilentlyContinue
go env -w GOBIN=D:\tools\go-workspace\bin
```

**注意**
- `go env -w GOPATH=...` 如果提示 `does not override conflicting OS environment variable`，说明 `GOPATH` 已在系统环境变量中设置，直接去系统环境变量里改即可，或保持现状只改上面三个变量。
- 改完后**必须重启 VSCode / Terminal**，新配置才会生效。
- 如果 C 盘已有旧的 `go-build` 缓存，可删除释放空间：`Remove-Item -Recurse -Force C:\Users\Administrator\AppData\Local\go-build`

**验证**
```powershell
go env GOMODCACHE    # 应输出 D:\tools\go-mod-cache
go env GOCACHE       # 应输出 D:\tools\go-cache
go env GOBIN         # 应输出 D:\tools\go-workspace\bin
dlv version          # 应正常输出版本号
```

---

## 3. Go 未使用的 import 编译错误

**症状**
```text
internal\api\server.go:6:2: "time" imported and not used
```

**原因**
Go 是严格编译型语言，不允许存在未使用的 import。VS Code 自动导入后如果后续删除了相关代码，import 不会自动移除。

**解决**
手动删除未使用的 import，或开启 VS Code 的 `"gopls": { "ui.diagnostic.annotations": { "bounds": true } }` 自动提示。

---

## 4. 前端 502 Bad Gateway

**症状**
浏览器 Network 面板显示 `/api/servers` 返回 `502 Bad Gateway`。

**原因**
Vite dev server 的 proxy 配置已将 `/api` 转发到 `localhost:8080`，但后端进程没有运行，proxy 找不到目标服务。

**解决**
确保后端服务已启动：
```bash
cd backend
go run cmd/api/main.go
```

**排查命令**
```bash
# 检查端口是否监听
netstat -ano | grep 8080

# 检查 Go 进程是否存在
tasklist | grep -i go
```

---

## 5. PowerShell 的 curl 不是真正的 curl

**症状**
PowerShell 里执行 `curl http://localhost:8080/api/servers` 报错或行为异常。

**原因**
PowerShell 的 `curl` 是 `Invoke-WebRequest` 的别名，语法和参数与系统自带的 `curl.exe` 完全不同。

**解决**
```powershell
# 方式 1：使用系统原生 curl
curl.exe http://localhost:8080/api/servers

# 方式 2：使用 Invoke-WebRequest
Invoke-WebRequest -Uri http://localhost:8080/api/servers
```

---

## 6. MSW 无条件拦截所有请求

**症状**
前端即使配置了 Vite proxy，请求仍然被 MSW 拦截，返回 mock 数据。

**原因**
`frontend/src/main.tsx` 中 `import.meta.env.DEV` 判断在开发模式下无条件启动 MSW，优先级高于浏览器真实网络请求。

**解决**
将 MSW 启动条件改为环境变量控制：
```tsx
if (import.meta.env.VITE_USE_MSW === 'true') {
  const { worker } = await import('./mocks/browser')
  await worker.start()
}
```
- 默认启动（走真实后端）：`npm run dev`
- 显式启用 MSW：`VITE_USE_MSW=true npm run dev`

---

## 7. 后端接口实现遗漏

**症状**
- 日志页面的「服务筛选」下拉框无效
- 服务器/部署列表刷新后顺序不一致
- 告警时间显示为 `2026-07-19T19:44:08+08:00` 这类不友好的格式

**原因**
从 OpenAPI Spec 到 Go Handler 的实现过程中，遗漏了部分查询参数和展示细节。

**解决**
| 问题 | 修复位置 | 修复内容 |
|------|---------|---------|
| 日志缺 service 筛选 | `log.go` | 添加 `if service != "" { query.Where("service = ?", service) }` |
| 服务器列表无排序 | `server.go` | 添加 `Order("CASE status WHEN 'running' THEN 1 ... END, id ASC")` |
| 部署列表无排序 | `deployment.go` | 添加 `Order("last_deployed_at DESC")` |
| 告警时间格式 | `dashboard.go` | `time.RFC3339` → `"01-02 15:04"` |

**经验**
后续新增接口时，建议对照 `spec/v1-api.yaml` 逐项检查：参数 → 筛选 → 排序 → 返回格式。

---

## 8. 目录重构时的嵌套错误

**症状**
执行 `mkdir -p frontend/backend/...` 时意外创建了 `frontend/backend/` 嵌套目录。

**原因**
命令是在 `devops-dashboard/` 根目录执行的，但理解有误，把 `backend/` 创建成了 `frontend/` 的子目录。

**解决**
```bash
rm -rf frontend   # 删除错误的嵌套结构
mkdir -p frontend backend/cmd/api backend/internal/api backend/internal/model backend/internal/repository backend/pkg/seed backend/storage
```

---

## 环境速查表

| 变量 | 值 | 作用 |
|------|-----|------|
| `GOPROXY` | `https://goproxy.cn,direct` | Go 模块国内加速下载 |
| `GOSUMDB` | `off` | 关闭模块校验和检查（仅开发环境） |
| `GOTMPDIR` | `D:\temp` | Go 编译临时文件目录（避免 C 盘满） |
| `GOMODCACHE` | `D:\tools\go-mod-cache` | 模块下载缓存目录（避免 C 盘满） |
| `GOCACHE` | `D:\tools\go-cache` | Go 编译缓存目录（避免 C 盘满） |
| `GOBIN` | `D:\tools\go-workspace\bin` | `go install` 二进制安装目录 |
| `GOPATH` | `D:\tools\go-workspace` | Go 工作区（依赖缓存位置） |
| `VITE_USE_MSW` | `true` / 不设置 | 控制前端是否启用 MSW |

---

## 启动命令速查

```bash
# 后端
cd backend
go run cmd/api/main.go          # 开发启动
go build -o devops-api cmd/api/main.go  # 编译可执行文件

# 前端（走真实后端）
cd frontend
npm run dev

# 前端（走 MSW Mock）
cd frontend
VITE_USE_MSW=true npm run dev
```
