# DevOps Dashboard — Agent Guidelines

> 本项目级指令供 AI Agent（Claude Code）在协助开发时遵循。

---

## 一、工作流约束（硬性）

### 1.1 提交策略：自动 Commit，不 Push

**每次代码修改完成后，Agent 应主动执行 `git add` + `git commit`（本地提交，遵循 1.2 的格式），但严禁执行 `git push`。**

- ✅ 修改完成后主动 commit，commit message 遵循 1.2 节格式
- ✅ 可拆分为多个语义独立的 commit（如 fix / docs / test 分开）
- ❌ **禁止 `git push`** — push 由用户本人检查确认后自行执行
- 若 commit 有误或需要调整，用户会告知，继续排查修复即可（可追加新的 commit，不强制 amend）
- 变更背景：2026-08-08 用户调整策略（原为"未经确认不得 commit"）

### 1.2 Commit Message 格式

本项目遵循 [Conventional Commits](https://www.conventionalcommits.org/) 约定。格式：`<type>(scope): <message>`

| type | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(backend): Agent 化（Pull 模型）` |
| `fix` | Bug 修复 | `fix(frontend): 分页翻页后数据不更新` |
| `docs` | 文档变更 | `docs: 构建与发布指南` |
| `refactor` | 代码重构（不改行为） | `refactor: Service 层抽取` |
| `test` | 测试相关 | `test(backend): 告警引擎测试` |
| `chore` | 工程事务（不改源代码） | `chore: .gitignore 修复屏蔽规则` |
| `style` | 格式调整（空格、分号等） | `style: gofmt` |

`scope` 可省略，常用值：`backend` / `frontend` / `docs`。

### 1.3 Git 操作禁止项（事故教训，2026-08-04）

> 背景：一次 `git rebase -i` 中途报 `not a valid object`，随后执行 `git reset --hard` 导致仓库对象库损坏、全部未推送提交丢失。

**禁止：**

- ❌ **禁止对未推送的本地提交直接 `git rebase -i`**（压缩历史前必须先推送远端备份）
- ❌ **禁止在 rebase/操作失败后直接 `git reset --hard`**——先确认对象库完好
- ❌ **禁止使用 `git clean -fd` 等带强制清理参数的命令**（可能连带删除未跟踪的源码文件）

**必须：**

- ✅ rebase / squash 前先 `git push` 到远端备份
- ✅ 任何 git 操作失败后，第一时间执行 `git reflog` 和 `git fsck --full` 检查状态
- ✅ rebase 中途出错立即 `git rebase --abort`，不要继续其他破坏性命令
- ✅ 把握不准时把输出贴给用户，不要自行决定恢复手段

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
- UI 层 `columns as any` 等类型绕过是允许的，但数��层必须类型安全

### 3.5 开发克制原则（硬性，2026-08-12 用户明确）

**写代码/加功能/修 bug 时，不要在 UI 上留多余的提示、解释文案或说明性组件。** 有什么权限就显示什么，没有就不显示；用户能自己看出状态差异，不需要界面"告诉他"。

- 禁止加"当前角色仅有查看权限…"之类的能力说明 Alert/文案
- 禁止加"此处功能需要 XX 权限"等解释性提示
- 权限差异通过**控件本身的状态**体现（禁用/隐藏/不可达）
- 确有必要的设计说明写进 docs/ 文档，不留在 UI
- 违反案例：webhook 设置页曾加"仅查看权限"只读提示，被用户要求移除（设计过度）

### 3.6 权限建模：前置依赖（Prerequisite Permission，硬性）

**权限点之间的关系有三种：互斥 / 独立 / 依赖。依赖关系必须建模为依赖，绝不能平铺成独立开关。**

- 操作类权限点（update/create/delete/deploy 等）通常依赖本模块的入口权限（read）
- 建模方式：后端 authz.permissionRequires 映射声明 操作点→入口点；UpdateRolePermissions 提交时**自动补全**依赖点（隐式继承，不报错）
- 前端配置矩阵据此联动：勾依赖点自动勾被依赖点；取消被依赖点级联取消依赖点
- 新增"操作类"权限点时，**必须**在 permissionRequires 声明其依赖，否则视为建模错误
- 理由：平铺依赖关系会产生"仅操作无入口"的死配置（可达性为 0），组合空间从 2^n 收敛到合法子集

### 3.7 orval 生成文件不可手工修改（硬性）

- src/api/client.ts、src/api/model/* 是 orval 生成的**产物**，重新生成会全量覆盖
- **禁止手工修改生成文件**（历史教训：JWT 拦截器手工加进 client.ts，orval 重生成后被覆盖丢失 → 全部受保护接口 401）
- 需要自定义逻辑（拦截器、请求封装）时，放独立文件（如 src/api/auth-interceptor.ts），由 main.tsx 导入
- 改 API 契约 → 改 spec/v1-api.yaml → 跑 orval 重新生成（命令见 3.2）
- orval 重生成后若大量文件显示 M 但无内容 diff，是 CRLF 换行符差异，提交时 git 自动规范化，无需处理

---

## 四、测试约定

> 详细编写规则见 `docs/testing-conventions.md`。

### 核心原则

1. **分类方式**：从被测代码特征推导测试模式 → 表驱动 / 独立函数 / 并发测试
2. **表驱动优先**：一份输入输出数据 + 一个 for 循环，断言只写一次
3. **并发测试只验"会不会崩"**：多个 goroutine 同时读写共享结构体，不 panic / 不 race 即通过
4. **测试文件名**：`*_test.go`，位于实现代码所在的同一包目录下

### 何时测

按优先级：复杂逻辑函数 → 有内部状态的函数 → 返回值语义易误解的函数。纯胶水代码不测。

### Playwright E2E：浏览器通道自动探测 + 权限组合矩阵（2026-08-12）

- **浏览器通道**：由 `playwright.config.js` 的 `detectBrowserChannel()` 自动探测（msedge → chrome → chromium），**spec 内禁止硬编码 `test.use({ channel })`**——多电脑（本机 Edge/他机 Chrome）同一份配置通用
- **权限类功能必须做组合矩阵测试**：权限组合空间是 2ⁿ（n=权限点数），**全量覆盖所有组合**，不能只测"有/无"两端。组合矩阵能暴露死配置（如"仅 update 无 read"菜单不可达）等依赖设计缺陷
- 矩阵测试写法：`setViewerPerms(perms)` 通过 admin API 预置角色权限 → 登录目标用户 → 断言每个组合下的 UI 状态（菜单可见性/控件禁用/按钮显隐）
- 登录后**不要用 `waitForLoadState('networkidle')`**（dashboard 有轮询 API 永远等不到）→ 改为等待 `.ant-menu-item` 或 `form` 渲染
- 测试间相互隔离：每个用例开头重置角色权限（避免前一个用例的权限改动影响后一个）

### 测试环境隔离（2026-08-12）

- 用户自起的 dev server（8080/5173）**不要动**；测试用独立端口 + 环境变量：
  - 后端：`PORT=8081 ./devops-api-xxx.exe`（独立 DB_PATH 测试库）
  - 前端：`VITE_API_PROXY=http://localhost:8081 VITE_PORT=5174 node node_modules/vite/bin/vite.js`
- `vite.config.ts` 已支持 `VITE_API_PROXY` / `VITE_PORT` 环境变量（默认 8080/5173）
- vite 在 config 文件变化后会退出（Re-optimizing dependencies 后进程消失）→ 改配置后重启
- 后台启动 + 端口占用排查：启动前先 `netstat` 确认目标端口空闲

---

## 五、文档维护

- 环境/工具问题 → `docs/env-setup-macos.md`
- 开发经验与问题排查 → `docs/development-guide.md`（第三章常见问题）
- 编码实践心得 → `docs/development-guide.md`（第二章前端模式）
- 新增文档后，在对应文件中记录**根因分析**和**经验教训**

---

*本文档随项目迭代持续更新。*
