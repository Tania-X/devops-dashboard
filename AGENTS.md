# DevOps Dashboard — 项目约定

> 供 AI 审查与开发 Agent 遵循的代码约定(硬性标记为不可违反项)。
> 工作流约定(git 提交策略/事故教训)见 docs/ 文档,不属于代码约定。

## 项目概览

- 技术栈: Vite + React 19 + TypeScript + Ant Design 5
- 架构: Feature-Based,按页面/功能模块组织代码
- 开发模式: SDD(Spec-Driven)— OpenAPI 契约优先;改契约走 `spec/v1-api.yaml` → orval 重新生成

## 目录与命名

- 新页面放 `src/features/{name}/`;页面组件 `{Feature}Page.tsx`;配置驱动文件 `{feature}-config.ts`
- 布局组件在 `src/components/layout/`;API 客户端在 `src/api/`(生成产物);Mock 在 `src/mocks/`

## API 与 Mock

- 所有 API 调用必须走 `src/api/client.ts`(生成的客户端),禁止直接 `fetch` 或手写 axios
- 自定义逻辑(拦截器等)放独立文件(如 `src/api/auth-interceptor.ts`),由 main.tsx 导入
- Mock 增强:自定义 handler 放 `setupWorker(...)` 之前覆盖生成逻辑;数据用**固定数据池**(保证筛选/分页可验证)

## UI

- 使用 Ant Design 5,遵循 24 列栅格
- 状态标签色值: running `#73bf69` / stopped `#aaaaaa` / maintenance `#f2c94c`
- 等宽字体优先 `Roboto Mono`(IP/主机名/MAC 地址等)

## TypeScript

- 严格模式开启
- 优先使用生成的 API Model 类型,避免重复定义接口
- UI 层类型绕过可接受(如 `columns as any`),数据层必须类型安全

## 开发克制(硬性)

- 不在 UI 加多余的提示/解释文案/说明组件;权限差异靠**控件本身状态**体现(禁用/隐藏/不可达)
- 禁止"当前角色仅有查看权限…"类能力说明;确有必要说明的写 docs/,不留 UI

## 权限建模(硬性)

- 权限点关系: 互斥 / 独立 / 依赖;**操作类权限点**(update/create/delete/deploy 等)必须依赖
  本模块入口(read),用 `authz.permissionRequires` 建模,禁止平铺成独立开关
- UpdateRolePermissions 提交时**自动补全**依赖点(隐式继承,不报错);前端矩阵联动(勾依赖点自动勾被依赖点,取消级联)
- 新增操作类权限点必须声明依赖,否则视为建模错误(会产生"仅操作无入口"的死配置)

## 生成文件(硬性)

- `src/api/client.ts`、`src/api/model/*` 是 orval 生成产物,重新生成全量覆盖,**禁止手工修改**
- orval 重生成后大量文件显示 M 但无内容 diff = CRLF 换行差异,git 自动规范化,无需处理

## 测试约定

- 分类从被测代码特征推导(表驱动/独立函数/并发);**表驱动优先**,一份输入输出 + 一个 for 循环,断言只写一次
- 并发测试只验"会不会崩"(不 panic/不 race 即通过);测试文件 `*_test.go`,位于实现代码同包
- 测试优先级: 复杂逻辑函数 → 有内部状态的函数 → 返回值语义易误解的函数;纯胶水代码不测
- E2E: 浏览器通道由 `playwright.config` 自动探测,**spec 禁止硬编码 channel**;权限功能必须做**组合矩阵测试**(全量组合,暴露死配置);登录后禁止 `waitForLoadState('networkidle')`(dashboard 有轮询 API),改等菜单/表单渲染
- 环境隔离: 测试用独立端口(后端 8081 / 前端 5174)+ 独立测试库,不动用户的 dev server

## Git 工作流

- 已合并的本地/远端分支**保留不删**(2026-08-16 用户约定),禁止主动 `git branch -d/-D` 清理;
  分支历史是资产,避免误删找回成本
- 其他 git 提交策略/事故教训见 `docs/` 文档

## 文档

- 环境/工具问题 → `docs/env-setup-macos.md`;开发经验与前端模式 → `docs/development-guide.md`
- 新增文档记录**根因分析**与**经验教训**
