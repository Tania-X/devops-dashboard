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
├── backend/                       # Go 后端
│   ├── cmd/api/main.go            # 入口
│   └── internal/
│       ├── api/                   # HTTP Handler（路由 + 控制器）
│       ├── model/                 # 数据模型
│       ├── monitor/               # gopsutil 采集器 + 历史缓存
│       ├── repository/            # GORM 数据库操作
│       └── pkg/seed/              # 模拟数据填充
└── docs/                          # 开发文档与踩坑记录
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

当前项目**尚无 Docker 化配置**。基础部署方式：

1. 后端编译为二进制，运行在目标服务器
2. 前端构建产物（`dist/`）由 Nginx 托管
3. Nginx 反代 `/api/*` 到后端端口

> 后续可添加 `Dockerfile` 和 `docker-compose.yml` 实现容器化部署。

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
