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

- **方向**：仅 Gitee → GitHub（单向）。Gitee 推代码后自动同步到 GitHub。
- **不反向**：GitHub 的改动不会自动回到 Gitee。
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
