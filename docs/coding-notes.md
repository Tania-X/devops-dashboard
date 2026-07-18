# 编码心得记录

> 记录 Server List 页面实现及 Mock 增强过程中的实践体会与技术决策。

---

## 一、Mock 数据策略：固定池 > 完全随机

**背景**：Orval 生成的 MSW handler 每次请求都会返回全新的随机数据，导致筛选和分页完全不可验证。

**决策**：在 `mocks/browser.ts` 中构建 35 条固定数据池 `ALL_SERVERS`，页面刷新前数据完全一致。

**好处**：
- 筛选结果可预期（点"运行中"一定只看到绿色标签）
- 分页可验证（第 1 页和第 2 页内容不同且稳定）
- 详情页可与列表行一一对应（`id` 一致）

**经验**：Mock 不是越随机越好，前端开发阶段需要的是**可复现的行为**。固定池 + 可控变化（如每次打开详情生成不同的磁盘使用率）是更好的平衡。

---

## 二、MSW Handler 优先级：自定义覆盖生成

**背景**：Orval 生成的 `getDevOpsDashboardAPIMock()` 包含了所有 API 的 mock handler，但它对请求参数不敏感。

**技巧**：将自定义 handler 放在 `setupWorker(...)` 的**前面**，利用 MSW "先匹配先响应"的规则实现覆盖：

```typescript
export const worker = setupWorker(
  customServerListHandler,   // ← 先匹配，拦截 /api/servers
  customServerDetailHandler, // ← 先匹配，拦截 /api/servers/:id
  ...getDevOpsDashboardAPIMock(), // 兜底其他接口
);
```

**经验**：不必 fork 或修改 Orval 的生成模板，通过在注册时"插队"即可优雅地增强特定接口的 mock 逻辑，保持 spec-first 的代码生成流程不被破坏。

---

## 三、筛选与分页的状态协同

**背景**：用户切换状态筛选时，如果当前在第 3 页，而筛选后的总数据不足 3 页，会出现空白页。

**决策**：任何筛选条件变化时，强制将页码重置为 1：

```typescript
const handleStatusChange = (value: string | undefined) => {
  setStatusFilter(value);
  fetchList(1, pagination.pageSize, value); // ← page = 1
};
```

**经验**：筛选和分页不是独立事件。筛选变化意味着数据全集发生了改变，继续停留在原页码是糟糕的体验。这是 AntD Table 的 `onChange` 和自定义筛选控件之间需要手动同步的点。

---

## 四、Drawer 详情页的信息组织

**背景**：服务器详情包含基本信息、磁盘分区、网卡三类信息，需要合理组织避免信息过载。

**决策**：
- **基本信息**：用 `Descriptions` 两列布局，密度高但整齐
- **磁盘分区**：用 `Row + Col + Progress`，视觉化使用率（>90% 红 / >75% 黄 / 正常绿）
- **网卡**：用内嵌 `Table`，结构统一且可扩展

**经验**：详情页不要一味堆砌 Descriptions，混合使用进度条、小型表格等多种呈现方式，能让用户更快抓住重点（磁盘告警一眼可见）。

---

## 五、TypeScript 与 AntD 的摩擦点

### 5.1 Table `columns` 类型推导

AntD Table 的 `columns` 在配合 `render` 函数时，TypeScript 推导容易断裂。本次采用 `columns as any` 绕过，原因：
- 列数多、render 逻辑复杂，完整类型标注会极度冗长
- 数据类型已通过 `ServerItem` 保证，列定义本身的类型安全收益不高

**经验**：在 UI 层和数据层类型已经隔离清晰时，对 presentation 层的 `any` 不必过度洁癖。

### 5.2 Select `value` 与枚举

`statusFilter` 是 `string | undefined`，但 API 参数类型是枚举 `ServerItemStatus`。使用 `as any` 进行桥接：

```typescript
.getServerList({ page, pageSize, status: status as any })
```

**经验**：Orval 生成的枚举类型在运行时就是字符串，TypeScript 层面的严格区分在前端筛选场景下反而成了摩擦。`as any` 在此处是务实的选择。

---

## 六、颜色决策：状态标签

**背景**：AntD 默认的 `success`、`error` 等预设色在深色主题下对比度不足或语义不符。

**决策**：自定义色值映射：

| 状态 | 色值 | 语义 |
|------|------|------|
| running | `#73bf69` | 健康绿，比 AntD 默认 `green` 更柔和 |
| stopped | `#aaaaaa` | 中性灰，表示无活性而非错误 |
| maintenance | `#f2c94c` | 警示黄，温和提醒而非紧急告警 |

**经验**：深色主题下，避免使用纯饱和色（如 `#ff0000`）。降低明度、提高柔和度能让界面看起来更专业，也更护眼。

---

## 七、性能与体验细节

### 7.1 Mock 延迟

为所有自定义 handler 添加了 `delay(200~300ms)`，原因：
- 无延迟的 Mock 会让本地开发时的加载状态难以测试
- 300ms 足以让 Spin/loading 效果显形，又不至于拖慢开发节奏

### 7.2 点击行打开 Drawer

为 Table 的 `onRow` 添加 `cursor: 'pointer'`，给用户明确的可点击暗示。

---

## 附录：SDD 流程验证

本次开发再次验证了 spec-first 流程：

1. **Spec 先行**：`spec/v1-api.yaml` 已定义 `GET /api/servers` 和 `GET /api/servers/{id}`
2. **Mock 基于 Spec**：Orval 生成基础 handler，再自定义增强行为
3. **UI 最后绑定**：前端直接调用 Orval 生成的 `getDevOpsDashboardAPI().getServerList()`，类型安全

全程未修改 API 契约，仅在前端和 Mock 层实现，前后端边界清晰。

---

*文档维护：随项目迭代持续更新。*
