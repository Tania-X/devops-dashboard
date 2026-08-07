# DevOps Dashboard

> 运维监控仪表盘 — 个人学习项目，用于探索 Go + React 全栈开发与系统监控。

![Tech Stack](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)
![Tech Stack](https://img.shields.io/badge/React-19-61DAFB?logo=react)
![Tech Stack](https://img.shields.io/badge/Vite-8-646CFF?logo=vite)
![Tech Stack](https://img.shields.io/badge/Ant%20Design-6-1677FF?logo=antdesign)

---

## 技术栈

| 层 | 技术 |
|------|------|
| 前端框架 | React 19 + TypeScript 6 |
| 构建工具 | Vite 8 |
| UI 组件库 | Ant Design 6 |
| 图表 | ECharts 6 |
| API 客户端 | Orval（从 OpenAPI 自动生成） |
| Mock | MSW (Mock Service Worker) |
| 后端框架 | Go 1.24+ / Gin |
| ORM | GORM |
| 数据库 | SQLite |
| 系统监控 | gopsutil (CPU/内存/磁盘/进程) |
| 开发模式 | SDD (Spec-Driven Development) — OpenAPI 契约优先 |

---

## 快速开始

### 前置条件

- **Node.js** >= 20（推荐 [fnm](https://github.com/Schniz/fnm) 管理版本）
- **Go** >= 1.24
- **pnpm**（推荐）或 npm

### 1. 启动后端

```bash
cd backend

# 安装依赖（首次）
go mod tidy

# 启动 API 服务（默认端口 8080）
go run cmd/api/main.go

# 指定端口
PORT=9090 go run cmd/api/main.go
```

验证：`curl http://localhost:8080/api/dashboard/metrics`

### 2. 启动前端

```bash
cd frontend

# 安装依赖（首次）
npm install

# 启动开发服务器（默认 http://localhost:5173）
npm run dev
```

前端默认通过 `VITE_API_BASE_URL` 连接后端，无特殊配置时使用相对路径 `/`（需后端运行在 5173 同端口或用反向代理）。

> **调试提示**：VSCode 中可按 F5 使用 `.vscode/launch.json` 中的配置一键启动后端调试。

---

## 项目结构

```
devops-dashboard/
├── spec/                          # OpenAPI 规范（SDD 核心）
│   └── v1-api.yaml                # 所有 API 接口定义
├── frontend/                      # React 前端
│   └── src/
│       ├── api/                   # Orval 自动生成的客户端与类型
│       ├── components/
│       │   └── layout/            # 全局布局
│       ├── features/
│       │   ├── dashboard/         # 系统概览
│       │   ├── server/            # 服务器管理
│       │   ├── log/               # 日志查询
│       │   ├── deployment/        # 部署状态
│       │   └── monitor/           # 实时监控（进程列表 + 主机信息）
│       └── mocks/                 # MSW Mock Handler
├── docker-compose.yml              # Docker 编排
├── backend/                       # Go 后端
│   ├── Dockerfile                  # 后端容器构建
│   ├── cmd/api/main.go            # 入口
│   └── internal/
│       ├── api/                   # HTTP Handler（路由 + 控制器）
│       ├── model/                 # 数据模型
│       ├── monitor/               # gopsutil 采集器 + 历史缓存
│       ├── repository/            # GORM 数据库操作
│       └── pkg/seed/              # 模拟数据填充
├── docs/                          # 开发文档
│   ├── architecture.md            # 架构总览与 API 实现状态
│   ├── development-guide.md       # 开发指南、常见问题与路线图
│   └── env-setup-macos.md         # macOS 环境搭建
```

---

## API 列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/dashboard/metrics` | 仪表盘实时指标（CPU / 内存 / 磁盘） |
| GET | `/api/dashboard/trend` | 历史趋势数据（支持 ?hours=6/12/24） |
| GET | `/api/dashboard/alerts` | 告警列表 |
| GET | `/api/servers` | 服务器列表 |
| GET | `/api/servers/:id` | 服务器详情 |
| GET | `/api/monitor/processes` | 进程列表（支持排序/搜索/条数限制） |
| GET | `/api/monitor/processes/:pid` | 进程详情 |
| GET | `/api/monitor/host` | 主机信息 |
| GET | `/api/logs` | 日志查询 |
| GET | `/api/deployments` | 部署记录 |
| GET | `/api/deployments/:id/history` | 部署历史 |

---

## 构建与部署

### 前端构建

```bash
cd frontend
npm run build

# 构建产物输出到 frontend/dist/
# 可用任意静态服务器托管：
npx serve dist          # 开发测试
# 或部署到 Nginx / CDN / Vercel / Netlify
```

### 后端构建

```bash
cd backend
go build -o server cmd/api/main.go

# 运行二进制（独立部署，无需 Go 环境）
./server
```

### 全量部署（后端 + 前端）

#### 方式 A：Docker 部署（推荐）

**开发模式（本地构建，默认）**：

```bash
# 启动（本地 build 镜像）
docker compose up -d

# 改了代码要重新构建
docker compose up -d --build

# 查看日志
docker compose logs -f

# 停止
docker compose down
```

**生产模式（拉取 ghcr.io 发布镜像）**：

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

> 说明：`docker-compose.yml` 默认本地构建（开发用）；`docker-compose.prod.yml`
> 用 `!reset null` 覆盖掉 build，改用 ghcr.io 发布的镜像（`build: !reset null` 是
> Compose Spec 覆盖语法）。发布新版本时更新 `docker-compose.prod.yml` 里的镜像 tag。

如果配置了 Agent 采集，启动时传入地址：

```bash
AGENT_HOSTS=192.168.1.100:9100 docker compose up -d
```

**说明**：
- 后端使用多阶段构建，运行镜像 ~15MB
- SQLite 数据通过 `volumes` 挂载到 `backend/storage/`，容器销毁数据不丢
- Agent **不放入 Docker**——Agent 需直接访问宿主机系统指标，应编译为裸 exe 在目标机器运行（见 `docs/build-deploy.md`）

### CI/CD 发布流程

代码推送后自动触发 CI（测试 + 构建产物），打 tag 触发 Docker 镜像构建并推送到 ghcr.io：

```bash
# 1. 推送代码（触发 CI 测试 + 构建，产出二进制与前端 dist）
git push origin main

# 2. 打版本 tag 并推送（触发 docker-publish，构建 server/web/agent 三个镜像推 ghcr.io）
git tag v0.2.1
git push origin v0.2.1
```

**版本号规则**：tag 一旦推送即不可变，失败后需递增小版本号（如 v0.2.1 → v0.2.2），不要删 tag 重打。

**镜像产物**（ghcr.io/tania-x/devops-dashboard/）：
- `server`：后端 API（Go + SQLite）
- `web`：前端（Nginx + 静态产物 + /api 反代）
- `agent`：监控采集代理

#### 方式 B：直接运行

```bash
# 前端
cd frontend && npm run build   # 产出 dist/

# 后端
cd backend && go build -o bin/devops-api.exe ./cmd/api
.\bin\devops-api.exe
```

前端产物由 Nginx 或后端 Gin 托管静态文件即可。

---

## 开发模式：SDD

本项目采用 **Spec-Driven Development**，流程为：

1. **契约先行** — 在 `spec/v1-api.yaml` 定义 OpenAPI 接口
2. **生成客户端** — `cd frontend && npx orval` 自动生成 TypeScript 类型 + API 函数
3. **后端实现** — 按接口定义实现 Go Handler
4. **前端消费** — 使用 Orval 生成的 Client 开发页面

如需修改 API，请先从 `spec/v1-api.yaml` 开始，然后重新生成客户端。

---

## 配置文件参考

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `PORT` | `8080` | 后端监听端口 |
| `DB_PATH` | `storage/devops.db` | SQLite 数据库路径 |
| `GIN_MODE` | `debug` | Gin 运行模式 |
| `VITE_API_BASE_URL` | `/` | 前端 API 代理地址 |

---

## License

[MIT](LICENSE)
