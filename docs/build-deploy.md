# 构建与发布指南

## 项目结构

```
devops-dashboard/
├── backend/                 # Go 后端（Gin + GORM + SQLite）
│   ├── cmd/
│   │   ├── api/main.go      # Server 入口
│   │   └── agent/main.go    # Agent 入口
│   ├── bin/                 # 构建产物目录
│   └── storage/             # SQLite 数据库文件
└── frontend/                # React 前端（Vite + Ant Design）
    └── dist/                # 前端构建产物
```

## 后端构建

### 编译

```powershell
cd backend

# Windows
go build -o bin/devops-api.exe ./cmd/api
go build -o bin/devops-agent.exe ./cmd/agent

# Linux（在 Windows 上交叉编译）
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o bin/devops-api-linux ./cmd/api
go build -o bin/devops-agent-linux ./cmd/agent
```

### 验证

```powershell
# 启动 Server
cd backend
.\bin\devops-api.exe

# 新开终端验证
curl http://localhost:8080/api/health
# → {"db":"connected","status":"ok"}

curl http://localhost:8080/api/dashboard/metrics
# → {"cpu":{"current":17,"status":"normal"},...}

curl http://localhost:8080/swagger/index.html
# → Swagger UI 页面
```

**注意**：exe 启动时会在当前目录找 `storage/devops.db`。如果在其它目录运行，加上 `DB_PATH` 环境变量：

```powershell
$env:DB_PATH="C:\Users\001\WorkBuddy\2026-07-20-17-50-43\devops-dashboard\backend\storage\devops.db"
.\bin\devops-api.exe
```

## 前端构建

### 编译

```powershell
cd frontend
npm run build
# 产出 dist/ 目录，包含 index.html + JS/CSS 静态文件
```

### 前后端联动的两种方式

#### 方式 A：开发期分离（当前状态）

```
前端 Vite Dev Server :5173 ──→ proxy /api → 后端 :8080
```

不需要构建前端。`npm run dev` 然后访问 `http://localhost:5173`。

#### 方式 B：生产态合一（推荐发布方式）

后端 Gin **直接托管前端的静态文件**。启动一个端口，访问 `http://localhost:8080` 直接看到前端页面。

**做法**：在 `router.go` 加一行：

```go
// 生产模式：前端静态文件
r.Use(static.Serve("/", static.LocalFile("./frontend/dist", true)))
```

或更简单——在后端 `main.go` 所在目录放一个 `public/` 目录，Gin 自动托管：

```go
r.Static("/", "./frontend/dist")
```

**部署时的目录结构**：

```
发布包/
├── devops-api.exe
├── storage/
│   └── devops.db
└── frontend/
    └── dist/
        ├── index.html
        └── assets/
```

启动命令不变，浏览器打开 `http://localhost:8080` 直接看到完整系统。

#### 方式 C：嵌入 Go 二进制（最干净）

用 Go 1.16+ 的 `embed` 特性把前端 dist 直接塞进 exe：

```go
//go:embed frontend/dist/*
var staticFiles embed.FS

// 在 router.go 里
r.GET("/*filepath", func(c *gin.Context) {
    c.FileFromFS(c.Request.URL.Path, http.FS(staticFiles))
})
```

编译后一个 `devops-api.exe` 就包含了前后端全部代码——拷贝到任何机器双击即可。

## 一键构建脚本

```powershell
# build.ps1 — 放到项目根目录
param([string]$target = "windows")

Write-Host "=== 构建前端 ==="
cd frontend
npm run build
cd ..

Write-Host "=== 构建后端 ==="
cd backend
if ($target -eq "linux") {
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -o bin/devops-api-linux ./cmd/api
    go build -o bin/devops-agent-linux ./cmd/agent
} else {
    go build -o bin/devops-api.exe ./cmd/api
    go build -o bin/devops-agent.exe ./cmd/agent
}
cd ..

Write-Host "=== 完成 ==="
Write-Host "后端: backend/bin/"
Write-Host "前端: frontend/dist/"
```

运行：

```powershell
.\build.ps1              # Windows
.\build.ps1 -target linux  # Linux
```

## 最小发布包

拷贝以下内容到目标机器：

```
发布包/
├── devops-api.exe      ← 后端
├── devops-agent.exe    ← Agent
└── storage/
    └── devops.db       ← 数据库
```

目标机器不需要 Go、Node.js、npm——一个 exe 搞定。

## 验证清单

| 检查项 | 状态 |
|--------|------|
| `go build -o bin/devops-api.exe ./cmd/api` 无报错 | |
| `.\bin\devops-api.exe` 启动成功 | |
| `curl http://localhost:8080/api/health` 返回 `{"db":"connected"}` | |
| `curl http://localhost:8080/api/dashboard/metrics` 返回含 CPU/Memory/Disk 的 JSON | |
| `curl http://localhost:8080/swagger/index.html` 返回 Swagger UI | |
| `go build -o bin/devops-agent.exe ./cmd/agent` 无报错 | |
| `.\bin\devops-agent.exe` 启动成功 | |
| `curl http://localhost:9100/api/metrics` 返回系统指标 JSON | |
