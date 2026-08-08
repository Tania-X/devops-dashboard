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

---

## 五、文档维护

- 环境/工具问题 → `docs/env-setup-macos.md`
- 开发经验与问题排查 → `docs/development-guide.md`（第三章常见问题）
- 编码实践心得 → `docs/development-guide.md`（第二章前端模式）
- 新增文档后，在对应文件中记录**根因分析**和**经验教训**

---

*本文档随项目迭代持续更新。*
