# Phase 2 工程化基建 — 软件设计文档 (SDD)

> **方向**: C — 工程化基建（代码质量）  
> **目标**: 建立可维护、可测试、可自动化的前端工程体系，消除一期快速迭代中积累的技术债务。  
> **日期**: 2026-07-19

---

## 1. 引言

### 1.1 编写目的

Phase 1 完成了 DevOps Dashboard 的 4 个核心功能页面（Dashboard、服务器管理、日志查询、部署状态），验证了产品概念和 UI 设计方向。但在快速迭代过程中，项目积累了显著的技术债务：

- 零测试覆盖，回归风险极高
- 类型安全薄弱，大量使用 `any` 绕过检查
- 重复代码遍布各页面（数据获取、卡片样式、状态标签）
- Mock/数据层、API 客户端、构建配置均未优化
- 无任何自动化门禁（CI、pre-commit）

本 SDD 定义 Phase 2 的工程化改造范围、技术方案和实施路线，确保代码库从"原型级"演进至"生产级"。

### 1.2 项目现状速览

| 维度 | 当前状态 | 风险等级 |
|---|---|---|
| 测试 | 无测试框架、无测试文件 | 🔴 高 |
| 类型安全 | `strict: false`，大量 `any` | 🔴 高 |
| 代码规范 | 仅 Oxlint（2 条规则），无 Prettier | 🟡 中 |
| CI/CD | 无 GitHub Actions | 🔴 高 |
| Git Hooks | 无 Husky / lint-staged | 🟡 中 |
| 构建优化 | 无代码分割、无路径别名、vendor 未拆包 | 🟡 中 |
| API 层 | Orval 生成的 client 未接入 customInstance，含 mock 代码 | 🔴 高 |
| 错误处理 | 页面级缺失，失败即白屏/无限 loading | 🔴 高 |
| 状态管理 | Zustand 已安装但完全未使用 | 🟢 低 |
| Mock 层 | 414 行 god file，职责混杂 | 🟡 中 |

---

## 2. 现状分析（As-Is）

### 2.1 类型与编译体系

- `tsconfig.app.json` 未启用 `strict: true`，`noImplicitAny` 等核心检查关闭
- `noUnusedLocals` / `noUnusedParameters` 已开启，但无法拦截显式 `any` 的使用
- Oxlint 仅配置了 `react/rules-of-hooks` 和 `react/only-export-components`，对 TypeScript 类型和通用代码规范无覆盖

### 2.2 API 与数据层

- `orval.config.ts` 配置了 `mock.output: './src/api/mocks'`，但实际生成位置未确认，且 `client.ts` 直接内联了 MSW mock factory 代码
- `src/api/axios-instance.ts` 定义了带 `baseURL` 的实例，但 Orval 生成的代码使用 `axios.default.get(...)`，custom instance 完全悬空
- `getDevOpsDashboardAPI()` 在每个组件的 `useEffect` 中重复实例化

### 2.3 前端页面层

- **重复的数据获取模式**：`ServerListPage`、`LogQueryPage`、`DeploymentPage` 均独立实现了 `useState(loading/data/pagination)` + `useEffect(fetch)` + `.then().finally()` 的相同模板
- **重复的视觉样式**：Dark Card 样式（`background: '#1f1f1f'`、`border: 'none'` 等）在 7 个文件中硬编码
- **重复的状态标签逻辑**：`StatusTag`、`LevelTag`、`EnvTag` 各自为政，颜色映射未统一
- **错误处理真空**：除 `ChartPanel.tsx` 有 `catch(console.error)` 外，其余页面无错误分支，API 失败即无限 loading

### 2.4 Mock 层

- `src/mocks/browser.ts` 414 行，混合了：常量定义、数据生成、业务过滤逻辑、MSW handler 注册
- 自定义 handler 与 Orval 生成的 `getDevOpsDashboardAPIMock()` 同时存在，后者因优先级问题实际为 dead code

### 2.5 构建与工程配置

- `vite.config.ts` 完全 stock，无 `resolve.alias`、无 `build.rollupOptions.manualChunks`、无 `server.proxy`
- `package.json` 的 `name` 仍为 `vite-init`
- `src/App.css` 为 Vite 模板遗留文件，疑似 dead code

---

## 3. 目标架构（To-Be）

### 3.1 架构原则

1. **防御性构建**：任何代码进入主分支前必须通过类型检查、Lint、测试三层门禁。
2. **DRY & 单一职责**：提取通用 Hook、组件、工具函数，消除跨页面重复。
3. **显式优于隐式**：消除 `any`，消除隐式依赖（如 `dayjs`），消除静默失败（如未处理的 Promise reject）。
4. **分层清晰**：API 层、Mock 层、状态层、UI 层职责分离，禁止跨层直接操作。

### 3.2 目标技术栈增补

| 新增工具 | 用途 | 替代/补充 |
|---|---|---|
| **Vitest** | 单元/集成测试 | 新增 |
| **@vitest/coverage-v8** | 代码覆盖率 | 新增 |
| **@testing-library/react** | 组件测试 | 新增 |
| **@testing-library/jest-dom** | DOM 断言辅助 | 新增 |
| **Prettier** | 代码格式化 | 补充 Oxlint |
| **Husky** | Git hooks | 新增 |
| **lint-staged** | 暂存区增量检查 | 新增 |
| **GitHub Actions** | CI 自动化 | 新增 |
| **TanStack Query (React Query)** | 服务端状态管理 | 替代手动 useEffect 请求 |

> **关于 Zustand 的决策**：当前 Zustand 已安装但完全未使用。Phase 2 先引入 TanStack Query 解决服务端状态（数据获取、缓存、分页）问题；客户端纯 UI 状态（如 sidebar 折叠、theme）暂继续使用 `useState` + Context，待有跨页面共享 UI 状态的明确需求后再启用 Zustand，避免过度设计。

---

## 4. 详细设计

### 4.1 类型安全强化（TS001）

**范围**: `tsconfig.app.json`、全源码  
**方案**:

1. 在 `tsconfig.app.json` 中启用 `strict: true`（等价于同时开启 `noImplicitAny`、`strictNullChecks`、`strictFunctionTypes`、`strictBindCallApply`、`strictPropertyInitialization`、`noImplicitThis`、`alwaysStrict`）
2. 增量修复编译错误：
   - 将页面中的 `as any` 替换为正确的生成类型（如 `GetServerListParams['status']`）
   - 将 Table 的 `columns={columns as any}` 改为 `TableProps<DataType>['columns']`
   - 将 `newPagination: any` 改为 `TablePaginationConfig`
   - 将 `dates: any` 改为 `[Dayjs, Dayjs] | null`
3. 在 `package.json` 中显式添加 `dayjs` 依赖（当前为 Ant Design 的隐式传递依赖）

**涉及文件**:
- `tsconfig.app.json`
- `src/features/servers/ServerListPage.tsx`
- `src/features/logs/LogQueryPage.tsx`
- `src/features/deployments/DeploymentPage.tsx`
- `src/features/dashboard/StatPanel.tsx`
- `src/mocks/browser.ts`
- `package.json`

---

### 4.2 工具链完善（TOOL001）

#### 4.2.1 Prettier 配置

创建 `.prettierrc`：

```json
{
  "semi": true,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5",
  "printWidth": 100
}
```

添加 `format` 和 `format:check` scripts。

#### 4.2.2 Oxlint 规则增强

在 `.oxlintrc.json` 中追加：

```json
{
  "rules": {
    "no-console": ["warn", { "allow": ["error"] }],
    "no-debugger": "error",
    "prefer-const": "error",
    "typescript/no-explicit-any": "warn"
  }
}
```

> 注：若 Oxlint 对某些规则支持不完整，评估迁移至 ESLint（`typescript-eslint` + `eslint-plugin-react-hooks`）。

#### 4.2.3 Git Hooks

安装 `husky` + `lint-staged`：

```bash
npx husky-init && npm install
npm install -D lint-staged
```

配置 `.lintstagedrc.json`：

```json
{
  "*.{ts,tsx}": ["oxlint", "prettier --write"],
  "*.{json,md,yaml,yml}": ["prettier --write"]
}
```

**涉及文件**:
- `.prettierrc`
- `.oxlintrc.json`
- `.husky/pre-commit`
- `.lintstagedrc.json`
- `package.json`

---

### 4.3 Vite 构建优化（BUILD001）

**范围**: `vite.config.ts`、`src/routes/index.tsx`  
**方案**:

1. **路径别名**：配置 `@/` → `./src`，消除深层相对路径 `../../../api/client`
2. **代码分割**：
   - 页面级：`React.lazy()` + `Suspense` 按路由懒加载
   - Vendor 级：`manualChunks` 拆分 `echarts`、`antd`、`react-router-dom`
3. **开发代理**：`server.proxy` 将 `/api` 转发到后端，消除 MSW 是开发唯一选择的限制

```typescript
// vite.config.ts 目标示例
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  server: {
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          antd: ['antd'],
          echarts: ['echarts'],
        },
      },
    },
  },
});
```

**涉及文件**:
- `vite.config.ts`
- `tsconfig.app.json` (compilerOptions.paths)
- `src/routes/index.tsx`
- `src/App.tsx` (Suspense wrapper)

---

### 4.4 API 层重构（API001）

**范围**: `orval.config.ts`、`src/api/*`  
**方案**:

1. **修复 Orval mutator**：在 `orval.config.ts` 中配置 `mutator: './src/api/axios-instance.ts'`，重新生成 client，使所有 API 调用走带 `baseURL` 的 custom instance
2. **Mock 代码分离**：确保 Orval 的 mock 生成到独立文件（`src/api/mocks/`），不与生产代码打包。若 Orval 的 `mock.output` 配置行为异常，则通过 `vite.config.ts` 的 `define` 或条件导入确保 mock factory 不被打入生产包
3. **Client 单例化**：在 `src/api/client.ts`（或新建 `src/api/index.ts`）中导出单例：`export const api = getDevOpsDashboardAPI()`，禁止组件内重复实例化
4. **统一错误拦截**：在 `axios-instance.ts` 中添加 response interceptor，对 4xx/5xx 统一做 `message.error()` 或 Toast 提示，避免每个页面自己处理

**涉及文件**:
- `orval.config.ts`
- `src/api/axios-instance.ts`
- `src/api/client.ts`
- `src/api/index.ts` (新建)

---

### 4.5 状态管理与数据获取重构（DATA001）

**范围**: 所有 `*Page.tsx`、`src/hooks/`（新建）  
**方案**:

引入 **TanStack Query (React Query)** 替代手动 `useState + useEffect` 请求模式。

**动机**：
- 当前 3 个列表页 + Dashboard 的 3 个 Panel 均重复实现了 loading/data/error/pagination/refetch 逻辑
- 无缓存，页面切换即重新请求
- 无错误重试、无后台刷新

**改造步骤**：

1. 安装依赖：`@tanstack/react-query`、`@tanstack/react-query-devtools`
2. 在 `main.tsx` 中挂载 `QueryClientProvider`
3. 为每个 API endpoint 定义 Query Key 策略：
   - `['dashboard', 'metrics']`
   - `['dashboard', 'trend']`
   - `['servers', { status, page, pageSize }]`
   - `['logs', { level, service, keyword, page, pageSize }]`
   - `['deployments']`
   - `['deploymentHistory', id]`
4. 创建通用 Hook：
   - `useDashboardMetrics()`
   - `useDashboardTrend()`
   - `useServerList(params)`
   - `useLogList(params)`
   - `useDeploymentList()`
5. 页面组件中直接消费这些 Hook，删除重复的 `useState/useEffect`

**涉及文件**:
- `package.json`
- `src/main.tsx`
- `src/hooks/useDashboard.ts` (新建)
- `src/hooks/useServers.ts` (新建)
- `src/hooks/useLogs.ts` (新建)
- `src/hooks/useDeployments.ts` (新建)
- `src/features/servers/ServerListPage.tsx`
- `src/features/logs/LogQueryPage.tsx`
- `src/features/deployments/DeploymentPage.tsx`
- `src/features/dashboard/StatPanel.tsx`
- `src/features/dashboard/ChartPanel.tsx`
- `src/features/dashboard/AlertListPanel.tsx`

---

### 4.6 UI 组件层重构（UI001）

**范围**: `src/components/`（新建/扩充）、各 Panel/Page  
**方案**:

1. **PageCard 通用容器**：提取 Dark Card 样式
   ```tsx
   <PageCard title="服务器管理">...</PageCard>
   ```
2. **StatusBadge 统一标签**：合并 `StatusTag`、`LevelTag`、`EnvTag` 为一个组件，通过 `variant` + `value` 驱动颜色
   ```tsx
   <StatusBadge variant="server" value="running" />
   <StatusBadge variant="level" value="ERROR" />
   <StatusBadge variant="env" value="prod" />
   ```
3. **Breadcrumb 自动推导**：将 `AppHeader.tsx` 中硬编码的 `breadcrumbMap` 移除，改为基于路由配置或当前 pathname 自动生成
4. **主题 Token 集中化**：将硬编码颜色 `#141414`、`#1f1f1f`、`#73bf69` 等提取到 `src/theme/tokens.ts`，或复用 Ant Design `ConfigProvider` 的 token

**涉及文件**:
- `src/components/ui/PageCard.tsx` (新建)
- `src/components/ui/StatusBadge.tsx` (新建)
- `src/components/layout/AppHeader.tsx`
- `src/components/layout/AppSidebar.tsx`
- `src/features/**/*.tsx`
- `src/theme/tokens.ts` (新建)

---

### 4.7 Mock 层重构（MOCK001）

**范围**: `src/mocks/*`  
**方案**:

将 `browser.ts` 按职责拆分为：

```
src/mocks/
├── browser.ts          # 仅负责 setupWorker 注册
├── server.ts           # 新增：Node 环境用（测试）
├── data/
│   ├── logStore.ts     # ALL_LOGS + generateLogList
│   ├── serverStore.ts  # ALL_SERVERS + generateServerList
│   ├── deploymentStore.ts
│   └── dashboardStore.ts
├── handlers/
│   ├── logHandlers.ts
│   ├── serverHandlers.ts
│   ├── deploymentHandlers.ts
│   └── dashboardHandlers.ts
└── utils/
    └── pagination.ts   # 通用 paginateArray, filterByKeyword
```

**涉及文件**:
- `src/mocks/browser.ts`（重构）
- `src/mocks/server.ts`（新建）
- `src/mocks/data/*`（新建）
- `src/mocks/handlers/*`（新建）
- `src/mocks/utils/*`（新建）

---

### 4.8 测试体系建设（TEST001）

**范围**: `src/**/*.test.{ts,tsx}`、`vitest.config.ts`  
**方案**:

1. **安装依赖**：`vitest`、`@vitest/coverage-v8`、`@testing-library/react`、`@testing-library/jest-dom`、`jsdom`
2. **配置 Vitest**：创建 `vitest.config.ts`，配置 `environment: 'jsdom'`、`globals: true`、`setupFiles: ['./src/test/setup.ts']`
3. **测试辅助**：创建 `src/test/setup.ts`（导入 `@testing-library/jest-dom`）、`src/test/render.tsx`（包装 QueryClientProvider + ConfigProvider + Router）
4. **分层测试策略**：
   - **单元测试**：`StatusBadge`、`PageCard`、工具函数（`validateRow`、`paginateArray`）
   - **集成测试**：Hook 测试（MSW + React Query）、Panel 测试（MSW 拦截 API + render）
   - **E2E（可选）**：Phase 2 暂不实施，留待 Phase 3
5. **覆盖率门禁**：初期设定 `statements: 60, branches: 50, functions: 60, lines: 60`，随迭代逐步提高

**涉及文件**:
- `vitest.config.ts` (新建)
- `package.json`
- `src/test/setup.ts` (新建)
- `src/test/render.tsx` (新建)
- `src/mocks/server.ts` (新建，用于测试环境 MSW)
- `src/**/*.test.ts` / `src/**/*.test.tsx`

---

### 4.9 CI/CD 流水线（CI001）

**范围**: `.github/workflows/ci.yml`  
**方案**:

创建 GitHub Actions workflow，在 PR 和 push 到 `main` 时执行：

```yaml
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci
      - run: npm run lint
      - run: npm run typecheck   # tsc --noEmit
      - run: npm run test
      - run: npm run build
```

并添加 `package.json` scripts：

```json
{
  "typecheck": "tsc --noEmit",
  "test": "vitest run",
  "test:watch": "vitest",
  "test:coverage": "vitest run --coverage"
}
```

**涉及文件**:
- `.github/workflows/ci.yml` (新建)
- `package.json`

---

## 5. 实施路线图

### Sprint 1：工具链与门禁（1 周）

> **目标**：建立"代码不通过检查就无法入库"的底线。

| ID | 任务 | 涉及文件 | 产出 |
|---|---|---|---|
| S1-1 | 启用 TS `strict: true` 并修复编译错误 | `tsconfig.app.json`, 全源码 | 零 `any`，零编译错误 |
| S1-2 | 安装并配置 Prettier | `.prettierrc`, `package.json` | 统一代码格式 |
| S1-3 | 增强 Oxlint 规则（或迁移 ESLint） | `.oxlintrc.json` | 拦截 `no-console`、显式 `any` |
| S1-4 | 安装 Husky + lint-staged | `.husky/pre-commit`, `.lintstagedrc.json` | 提交前自动 lint + format |
| S1-5 | 配置 Vite 路径别名 | `vite.config.ts`, `tsconfig.app.json` | `@/` 可用 |
| S1-6 | 清理 dead code | `src/App.css`, `package.json` name | 删除无用文件 |
| S1-7 | 创建 GitHub Actions CI | `.github/workflows/ci.yml` | PR 自动检查 |

**Sprint 1 门禁**：
- `npm run lint` 通过
- `npm run typecheck` 通过
- `npm run build` 通过
- CI 绿

---

### Sprint 2：架构与复用（1.5 周）

> **目标**：消除重复代码，修复 API 层和数据获取层的结构性缺陷。

| ID | 任务 | 涉及文件 | 产出 |
|---|---|---|---|
| S2-1 | 重构 Orval client：接入 customInstance + 单例导出 | `orval.config.ts`, `src/api/axios-instance.ts`, `src/api/index.ts` | API 层可维护 |
| S2-2 | 引入 TanStack Query 并替换手动请求 | `src/main.tsx`, `src/hooks/*`, 各 Page/Panel | 删除 80% 的重复 useEffect |
| S2-3 | 统一错误处理（axios interceptor + query error handling） | `src/api/axios-instance.ts`, `src/main.tsx` | 全局错误提示 |
| S2-4 | 提取 UI 公共组件（PageCard、StatusBadge） | `src/components/ui/*`, 各 Page/Panel | 消除硬编码样式 |
| S2-5 | 自动推导 Breadcrumb | `src/components/layout/AppHeader.tsx` | 消除手动同步风险 |
| S2-6 | 拆分 Mock god file | `src/mocks/*` | 职责清晰，可测试 |
| S2-7 | 添加 `dayjs` 显式依赖 | `package.json` | 消除隐式传递依赖 |

**Sprint 2 门禁**：
- 所有页面功能与 Phase 1 完全一致（回归测试通过）
- `npm run lint` / `typecheck` / `build` 通过
- 无新的 `any` 引入

---

### Sprint 3：测试体系与性能（1.5 周）

> **目标**：建立自动化测试体系，优化构建产物。

| ID | 任务 | 涉及文件 | 产出 |
|---|---|---|---|
| S3-1 | 安装并配置 Vitest + RTL + jsdom | `vitest.config.ts`, `package.json` | 测试框架就绪 |
| S3-2 | 创建测试辅助（render wrapper、MSW server） | `src/test/setup.ts`, `src/test/render.tsx`, `src/mocks/server.ts` | 测试环境完备 |
| S3-3 | 编写单元测试：工具函数 + 纯组件 | `src/**/*.test.ts` / `*.test.tsx` | 核心逻辑有覆盖 |
| S3-4 | 编写集成测试：Hook + Panel（MSW 拦截） | `src/hooks/*.test.ts`, `src/features/dashboard/*.test.tsx` | API 交互有覆盖 |
| S3-5 | 配置代码覆盖率与门禁 | `vitest.config.ts`, CI workflow | 覆盖率可追踪 |
| S3-6 | 实现路由懒加载 + Vendor 拆包 | `src/routes/index.tsx`, `vite.config.ts` | 首屏加载优化 |
| S3-7 | 验证生产构建产物体积 | `dist/` | 确认 vendor 分离生效 |

**Sprint 3 门禁**：
- `npm run test` 通过
- 覆盖率 ≥ 60%（statements）
- 生产构建产物无 mock/faker 代码
- Lighthouse / 网络面板验证首屏 JS 减小

---

## 6. 验收标准

### 6.1 功能回归

- Dashboard、服务器管理、日志查询、部署状态 4 个页面的全部功能与 Phase 1 末端完全一致
- 筛选、分页、搜索、详情展开等交互行为正常

### 6.2 代码质量指标

| 指标 | 当前值 | 目标值 |
|---|---|---|
| `any` 使用次数 | ~15 处 | 0 处 |
| TS `strict` 模式 | ❌ 关闭 | ✅ 开启且无错误 |
| 重复代码块（硬编码 Card 样式） | 7 处 | 0 处（统一组件） |
| 重复数据获取模板 | 6 处 | 0 处（统一 Hook） |
| 未处理的 Promise reject | 5 处 | 0 处 |
| 测试覆盖率 | 0% | ≥ 60% |

### 6.3 工程指标

- PR 合并前必须通过：Lint ✅、Type Check ✅、Test ✅、Build ✅
- 生产构建产物无 `faker`、`msw` 相关代码（tree-shaking 验证）
- 首屏加载 JS 体积下降 ≥ 20%（通过代码分割）

---

## 7. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|---|---|---|
| `strict: true` 导致大量编译错误，修复工作量超预期 | 延期 | 按文件分批修复，优先核心页面；非关键文件可暂时 `// @ts-expect-error` 标记 |
| TanStack Query 引入改变数据获取心智模型，团队成员学习成本 | 质量 | Sprint 2 第一周先搭建 1 个示例 Hook（如 `useServerList`），其余照模板迁移 |
| Orval mock 分离后，开发环境行为变化 | 回归 | 保留 `browser.ts` 的自定义 handlers 不变，仅调整文件结构；开发环境验证后再合并 |
| 测试编写拖慢交付 | 进度 | 优先覆盖工具函数和纯组件（快、稳），再覆盖复杂 Panel；不追求 100% 覆盖 |
| CI 运行时间变长 | 体验 | 使用 `actions/cache` 缓存 `node_modules`；Vitest 单线程足够快，无需并行优化 |

---

## 8. 附录

### 8.1 术语表

| 术语 | 说明 |
|---|---|
| SDD | Software Design Document，软件设计文档 |
| MSW | Mock Service Worker，用于拦截 HTTP 请求的 mock 工具 |
| Orval | 基于 OpenAPI 自动生成 TypeScript API Client 的工具 |
| TanStack Query | 原 React Query，用于服务端状态管理的库 |
| Oxlint | Oxc 项目下的 JavaScript/TypeScript Linter，Rust 编写，性能极高 |
| DRY | Don't Repeat Yourself，消除重复代码原则 |

### 8.2 参考文档

- [Orval Documentation — Mocking](https://orval.dev/reference/configuration/output#mock)
- [TanStack Query React Quick Start](https://tanstack.com/query/latest/docs/framework/react/quick-start)
- [Vitest Getting Started](https://vitest.dev/guide/)
- [MSW v2 API](https://mswjs.io/docs/api/setup-worker)

---

*文档版本: v1.0*  
*维护人: DevOps Dashboard Team*  
*下次评审: Sprint 3 结束后*
