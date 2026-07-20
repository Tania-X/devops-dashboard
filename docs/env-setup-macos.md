# macOS 环境搭建记录

> 本记录面向 macOS（Apple Silicon / Intel）开发者，覆盖 Go + Node.js 全栈环境配置，以及国内网络下的依赖加速方案。

---

## 环境信息

| 项目 | 版本/路径 | 说明 |
|------|----------|------|
| 操作系统 | macOS 15.x (Darwin) | Apple Silicon (arm64) 或 Intel (amd64) |
| Shell | zsh | macOS 默认 Shell |
| Homebrew | 最新版 | 建议作为 macOS 首选包管理器 |
| Go 版本 | 1.22+ | 后端开发必需 |
| Node.js 版本 | 20 LTS | 前端开发必需 |

---

## 一、Go 环境配置

### 1.1 安装 Go

推荐通过 Homebrew 安装：

```bash
brew install go
```

验证安装：

```bash
go version   # 应输出 go1.22.x 或更高
```

> **Apple Silicon 注意**：Homebrew 默认安装 arm64 版本，与 Apple Silicon 原生兼容。如需交叉编译到 Linux AMD64，Go 天然支持：`GOOS=linux GOARCH=amd64 go build`。

### 1.2 配置国内模块代理（关键步骤）

Go 默认使用 `proxy.golang.org`，在国内下载依赖极慢或超时。

**已配置的代理（七牛云镜像）：**

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off
```

验证配置：

```bash
go env GOPROXY    # 应输出 https://goproxy.cn,direct
go env GOSUMDB    # 应输出 off
```

**备选代理（阿里云）：**

```bash
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
```

> `GOSUMDB=off` 仅建议在开发环境使用，关闭模块校验和数据库检查，加速下载。生产环境 CI 建议开启。

### 1.3 Go 环境变量速查（macOS 默认值）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `GOPATH` | `~/go` | Go 工作区，存放下载的模块和工具 |
| `GOMODCACHE` | `~/go/pkg/mod` | 模块下载缓存 |
| `GOCACHE` | `~/Library/Caches/go-build` | 编译缓存 |
| `GOBIN` | `~/go/bin` | `go install` 安装的二进制目录 |

macOS 下通常**不需要**像 Windows 那样手动搬移目录，因为 `~/go` 和 `~/Library/Caches` 都在用户目录下，不会导致系统盘空间问题。但如果你使用容量较小的 Mac，可以考虑外接 SSD 并修改缓存路径：

```bash
# 示例：将模块缓存改到外接 SSD（可选）
mkdir -p /Volumes/External/go-mod-cache
go env -w GOMODCACHE=/Volumes/External/go-mod-cache
```

### 1.4 将 GOBIN 加入 PATH

确保 `go install` 安装的工具（如 `dlv`、`air`）可以在终端直接调用：

```bash
# 写入 ~/.zshrc
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

---

## 二、Node.js 环境配置

### 2.1 安装 fnm（推荐）

[fnm](https://github.com/Schniz/fnm) 是跨平台的 Node 版本管理器，支持 macOS：

```bash
brew install fnm
```

配置 Shell 集成（zsh）：

```bash
echo 'eval "$(fnm env --use-on-cd)"' >> ~/.zshrc
source ~/.zshrc
```

安装 Node 20 LTS：

```bash
fnm install 20
fnm default 20
```

验证：

```bash
node -v   # v20.x.x
npm -v    # 10.x.x
```

### 2.2 配置国内 npm 镜像（可选）

```bash
# 使用淘宝镜像
npm config set registry https://registry.npmmirror.com

# 或使用腾讯云镜像
npm config set registry https://mirrors.cloud.tencent.com/npm/

# 验证
npm config get registry
```

---

## 三、项目启动流程

### 3.1 首次克隆后初始化

```bash
# 1. 进入项目根目录
cd devops-dashboard

# 2. 后端依赖下载（已配置国内代理，应很快）
cd backend
go mod tidy

# 3. 前端依赖安装
cd ../frontend
npm install
```

### 3.2 日常开发启动

**终端 1 — 启动后端：**

```bash
cd backend
go run cmd/api/main.go

# 预期输出：
# 数据库已连接 path=storage/devops.db
# 数据初始化完成 servers=35 logs=200 deployments=15
# 服务启动 address=http://localhost:8080
```

**终端 2 — 启动前端（联调模式，直连 Go 后端）：**

```bash
cd frontend
npm run dev

# 浏览器访问 http://localhost:5173
# /api/* 请求自动通过 Vite proxy 转发到 localhost:8080
```

**终端 2 — 启动前端（Mock 模式，独立前端开发）：**

```bash
cd frontend
VITE_USE_MSW=true npm run dev
```

---

## 四、macOS 特有注意事项

### 4.1 Gatekeeper / 权限问题

如果运行 `go build` 输出的二进制文件或 `fnm` 安装的 Node 时遇到 "无法打开，因为无法验证开发者"：

```bash
# 方式 1：在系统设置 → 隐私与安全性 → 安全性 中点击"仍要打开"
# 方式 2：使用 xattr 移除隔离属性（仅限你信任的程序）
xattr -d com.apple.quarantine /path/to/binary
```

### 4.2 VS Code 命令行启动

建议通过命令行启动 VS Code，确保继承正确的 PATH：

```bash
code /path/to/devops-dashboard
```

如果 `code` 命令不可用：

```bash
# 在 VS Code 中按 Cmd+Shift+P，输入 "Shell Command: Install 'code' command in PATH"
```

### 4.3 端口占用排查

macOS 下检查端口占用：

```bash
# 查看 8080 端口是否被占用
lsof -i :8080

# 结束占用进程（谨慎使用）
kill -9 <PID>
```

### 4.4 SQLite 文件权限

后端 SQLite 数据库文件 `backend/storage/devops.db` 需要写权限。如果项目目录是从 Windows 磁盘或网络驱动器复制过来的，可能需要修复权限：

```bash
cd backend
mkdir -p storage
chmod 755 storage
```

---

## 五、环境速查表

```bash
# Go 代理（已配置）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# npm 镜像（可选）
npm config set registry https://registry.npmmirror.com

# 启动后端
cd backend && go run cmd/api/main.go

# 启动前端（联调）
cd frontend && npm run dev

# 启动前端（Mock）
cd frontend && VITE_USE_MSW=true npm run dev
```

---

## 六、与 Windows 文档的差异对照

| 场景 | Windows（[env-setup.md](./env-setup.md)） | macOS（本文档） |
|------|------------------------------------------|----------------|
| Go 安装 | 官网下载 `.msi` | `brew install go` |
| PATH 配置 | PowerShell `[Environment]::SetEnvironmentVariable` | 写入 `~/.zshrc` |
| C 盘空间 | 需手动搬移 `GOMODCACHE`/`GOCACHE` | 通常无需处理 |
| 端口排查 | `netstat -ano \| findstr 8080` | `lsof -i :8080` |
| Shell | PowerShell / Git Bash | zsh |
| 编译缓存 | `D:\temp` | `~/Library/Caches/go-build` |

---

*本文档随项目迭代持续更新。*
