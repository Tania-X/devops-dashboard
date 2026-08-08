# Gitee + GitHub 双仓库同步指南

> 解决国内访问 GitHub 慢的问题：日常用 Gitee，自动同步回 GitHub。

## 远程仓库配置

本地仓库配置了两个远程地址：

```bash
origin  → Gitee（默认推送目标）
github  → GitHub（备用，有 VPN 时才推）
```

查看方式：

```bash
git remote -v
```

## 日常操作

```bash
# 推送代码（默认到 Gitee，国内秒完成）
git push

# 等同于
git push origin main

# 如果需要在本地拉取 GitHub 的更新
git pull github main
```

## Gitee 镜像同步（自动同步到 GitHub）

Gitee 提供仓库镜像功能，推送到 Gitee 后自动同步到 GitHub。

### 配置步骤

1. 浏览器打开 Gitee 仓库 → **管理** → **仓库镜像管理**
2. 点击 **添加镜像**
3. 目标地址：`https://github.com/Tania-X/devops-dashboard.git`
4. 用户名：GitHub 用户名
5. 密码：**GitHub Personal Access Token**（不是登录密码）
   - 生成 PAT：GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens
   - 权限选 `Contents: Read and write`（仓库内容读写）
6. 点击保存

### 镜像同步限制

- **方向**：原仅配置 Gitee → GitHub（单向）。Gitee 推代码后自动同步到 GitHub。
- **反向可行**：Gitee 官方支持 Pull 方向镜像（GitHub → Gitee），见下文「GitHub → Gitee 反向同步」。
- **单人开发无影响**：因为只有你一个人提交代码，全程推 Gitee 即可，GitHub 自动保持同步。

## 为什么这么做

| 痛点 | 解决 |
|------|------|
| GitHub push 经常断连/超时 | 改用 Gitee 推送，国内秒完成 |
| 代码还是要备份到 GitHub | Gitee 镜像自动同步 |
| 双仓库操作麻烦 | 默认推 `origin`（Gitee），`git push` 一键完成 |

## 如果推送到 GitHub 需要两边同步

```bash
# 手动双推（需要 VPN 能连 GitHub）
git push && git push github main
```

---

## GitHub 优先工作流(2026-08-08 起)

> 背景：GitHub 关联了 CodeRabbit(AI 审核 PR)。日常开发切换到 GitHub 为主，
> 采用 feature 分支 + PR 流程，合并后再同步 Gitee。

### 一次性切换(当前目录直接加 remote，无需新建目录)

```bash
# 1. 添加 GitHub remote
git remote add github https://github.com/Tania-X/devops-dashboard.git

# 2. 先把本地 main 同步到 GitHub(保证 GitHub main 最新，含未推送 commit)
git push github main

# 3. 基于 GitHub main 建开发分支
git fetch github
git checkout -b feat/xxx github/main
```

### 日常开发流程

```bash
git push github feat/xxx       # 推送开发分支
# → GitHub 上开 PR → CodeRabbit 自动 AI 审核 → 人工确认合并

git checkout main
git pull github main           # 合并后更新本地 main
git push origin main           # 可选：同步回 Gitee(镜像/备份)
```

### Remote 约定

| remote | 地址 | 角色 |
|--------|------|------|
| `origin` | Gitee | 备份 + 国内访问(保留) |
| `github` | GitHub | **开发主线**，feature 分支与 PR 都在这里 |

- 开发分支显式推 `github`；`main` 合并后按需回推 `origin` 保持 Gitee 同步
- 原"Gitee 镜像自动同步"可保留，方向不变(Gitee → GitHub)，开发分支不受镜像影响

---

## GitHub → Gitee 反向同步(Pull 方向镜像)

> 2026-08-08 探索确认：Gitee 官方支持 Pull 方向镜像，可让 GitHub 直接覆盖同步到 Gitee。
> 来源：Gitee 帮助中心「仓库镜像管理」+ 官方博客。

### 两种方式

| 方式 | 操作 | 特点 |
|------|------|------|
| **仓库主页「同步更新」按钮** | Gitee 仓库主页直接点按钮 | 强制同步(新代码直接覆盖 Gitee)，适合一次性手动 |
| **Pull 方向镜像(推荐)** | 镜像配置 → 方向选 Pull | 可选手动点更新 / 自动(webhook 触发)，适合日常 |

### Pull 方向镜像配置步骤

1. Gitee 仓库 → **管理** → **仓库镜像管理** → **添加镜像**
2. **镜像方向**选择 **Pull**(GitHub → Gitee)
3. **镜像仓库**下拉选择 GitHub 仓库
4. **个人令牌**填 GitHub PAT(必须含 `repo` 访问授权)
5. 可选勾选「**自动从 GitHub 同步仓库**」：
   - 勾选 → Gitee 自动生成 webhook 实现自动镜像，**需要 PAT 含 `admin:repo_hook` 授权**，否则添加失败
   - 不勾选 → 手动点「更新」触发同步
6. 保存后，触发方式：推送代码到 GitHub 镜像仓库 / 手动更新镜像(**最短间隔 5 分钟**)
7. 同步内容：分支(Branches)、标签(Tags)、提交记录(Commits)

### ⚠️ 覆盖式同步风险

- Pull 镜像为**强制覆盖、单向**(以 GitHub 为准)，Gitee 上的**独有提交会被覆盖丢失**
- 适用前提：**GitHub 为唯一真实源，Gitee 只做国内只读备份**
- 纪律：不要直接往 Gitee push 独有代码；反向同步前确认 Gitee 无本地独有改动
- Gitee 帮助中心提示：**双向镜像存在代码丢失风险，请谨慎使用**（本项目不需要双向，单向 Pull 即可）
