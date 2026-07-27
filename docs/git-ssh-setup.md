# Git SSH 配置指南

> 解决 `git push` 时反复输入密码的问题，一次配置永久使用。

## 1. 检查是否已有 SSH key

```bash
ls ~/.ssh/id_ed25519.pub
```

如果文件存在说明以前配过，跳到第 3 步。

## 2. 生成 SSH key

```bash
ssh-keygen -t ed25519 -C "你的GitHub邮箱@example.com"
```

一路 Enter 用默认路径，**密码直接留空**（回车两次）。

## 3. 复制公钥内容

```bash
# Windows PowerShell 用完整路径
cat C:\Users\你的用户名\.ssh\id_ed25519.pub
# 或者用记事本打开
notepad C:\Users\你的用户名\.ssh\id_ed25519.pub
```

会看到一行以 `ssh-ed25519` 开头的文本，**全部复制**。

## 4. 添加到 GitHub

1. 浏览器打开 https://github.com/settings/ssh/new
2. Title 随便填，比如 `"my-pc"` 或 `"办公电脑"`
3. Key 粘贴刚才复制的那一行
4. 点 **Add SSH key**

## 5. 验证连通性

```bash
ssh -T git@github.com
```

应该看到：`Hi 你的用户名! You've successfully authenticated`

## 6. 切换远程仓库地址到 SSH

如果之前用的 HTTPS 地址，改为 SSH：

```bash
# 查看当前远程地址
git remote -v

# 改为 SSH（只需第一次）
git remote set-url origin git@github.com:你的用户名/你的仓库名.git

# 验证
git remote -v
# 应该显示: git@github.com:xxx/xxx.git (fetch)
#          git@github.com:xxx/xxx.git (push)
```

## 7. 推送验证

```bash
git push origin main
```

配置完毕后，后续所有 `git push` / `git pull` 都不再需要密码。

## 常见问题

### Q: `Permission denied (publickey)`
SSH key 没有添加到 GitHub，或者添加的不是当前机器的 key。

### Q: `Could not connect to server` 
网络问题，检查 VPN 是否开启、能否访问 github.com。

### Q: 有多台电脑怎么办？
每台电脑各自生成 key，全部添加到同一个 GitHub 账号下（最多可加 100 个）。
