# DevOps Dashboard — Agent Guidelines

> 本项目级指令供 AI Agent（Claude Code）在协助开发时遵循。

---

## 一、工作流约束（硬性）

### 1.1 禁止自动 Commit / Push

**除非用户显式说明"commit"、"push"或"提交"等指令，否则 Agent 不得执行任何 `git add`、`git commit`、`git push` 操作。**

- 可以编写 commit message 草稿供用户审阅
- 可以提示用户"变更已就绪，是否需要提交？"
- 不得在未经确认的情况下将代码推送到远程仓库

---

## 二、项目概述

- **名称**：DevOps Dashboard
- **技术栈**：Vite + React 19 + TypeScript + Ant Design 5
- **架构**：Feature-Based，按页面/功能模块组织代码
- **开发模式**：SDD（Spec-Driven Development）— OpenAPI 契约优先

---

## 三、编码规范

### 3.1 目录结构

```
src/
├── components/layout/    # 全局布局组件
├── features/
│   ├── dashboard/        # Dashboard 页面及配置
│   ├── server/           # 服务器管理页面
│   ├── logs/             # 日志查询页面
│   └── deployments/      # 部署状态页面
├── api/                  # Orval 生成的 API 客户端 + Model
├── mocks/                # MSW Mock Handlers
└── main.tsx
```

- 新页面必须放入 `features/{name}/`
- 页面级组件命名：`{Feature}Page.tsx`
- 配置驱动文件：`{feature}-config.ts`

### 3.2 API 与 Mock

- 所有 API 调用必须通过 `src/api/client.ts` 生成的客户端
- 禁止直接写 `fetch` 或手写 axios 调用
- Mock 增强时，自定义 handler 必须放在 `setupWorker(...)` 的前面以覆盖生成逻辑
- Mock 数据应使用**固定数据池**（保证筛选、分页可验证），而非完全随机

### 3.3 UI 规范

- 使用 Ant Design 5.x 组件，遵循其 24 列栅格系统
- 深色主题色值参考 `spec/ui-theme.md`
- 状态标签自定义色值：
  - running: `#73bf69`
  - stopped: `#aaaaaa`
  - maintenance: `#f2c94c`
- 等宽字体优先使用 `Roboto Mono`（IP、主机名、MAC 地址等）

### 3.4 TypeScript

- 严格模式开启
- 优先使用生成的 API Model 类型，避免重复定义接口
- UI 层 `columns as any` 等类型绕过是允许的，但数据层必须类型安全

---

## 四、文档维护

- 环境/工具问题 → `docs/env-setup.md`
- 项目构建踩坑 → `docs/project-issues.md`
- 编码实践心得 → `docs/coding-notes.md`
- 新增文档后，在对应文件中记录**根因分析**和**经验教训**

---

*本文档随项目迭代持续更新。*
