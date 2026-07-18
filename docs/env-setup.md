# 环境搭建记录

> 记录 DevOps Dashboard 前端项目在 Windows 环境下初始化时遇到的问题及解决方案，供后续参考和团队复现。

---

## 环境信息

| 项目 | 版本/路径 |
|------|----------|
| 操作系统 | Windows 10 Pro |
| Shell | Git Bash (MSYS2) |
| Node.js 版本 | 20.20.2 LTS |
| Node.js 安装路径 | `D:\tools\fnm\node-versions\node-versions\v20.20.2\installation` |
| fnm 路径 | `D:\tools\fnm\fnm.exe` |
| 项目路径 | `D:\my_projects\devops-dashboard` |

---

## 问题 1：npm / node 命令不可用

### 现象

在 Git Bash 中执行 `npm` 或 `node` 命令，提示：

```bash
/usr/bin/bash: line 1: npm: command not found
/usr/bin/bash: line 1: node: command not found
```

### 根因

当前 Shell 的 `PATH` 环境变量中没有包含 Node.js 的可执行文件目录。Windows 上安装 Node 后，如果未自动写入 PATH，或使用的是便携版/版本管理器（如 fnm），需要手动配置。

### 解决方案

将 Node.js 的实际安装路径追加到**用户级 PATH** 环境变量：

```powershell
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "User") + ";D:\tools\fnm\node-versions\node-versions\v20.20.2\installation",
    "User"
)
```

> ⚠️ 执行后必须**关闭所有终端窗口并重新打开**，新 PATH 才会生效。

### 验证

```bash
node -v   # v20.20.2
npm -v    # 10.7.0
```

---

## 问题 2：C 盘空间紧张，需将 Node.js 安装到其他盘

### 现象

默认 Node.js 安装到 `C:\Program Files\nodejs` 或 `C:\Users\<用户名>\AppData\Local\fnm`，C 盘空间不足。

### 根因

Node.js 官方 `.msi` 安装包默认往 C 盘写；fnm 默认也将版本存放在 `%LOCALAPPDATA%\fnm`（即 C 盘用户目录下）。

### 解决方案

使用 **fnm**（Fast Node Manager）并自定义 `FNM_DIR` 环境变量，将 Node 版本存到 D 盘。

#### 安装步骤

1. 下载 `fnm-windows.zip` 从 [GitHub Releases](https://github.com/Schniz/fnm/releases)
2. 解压 `fnm.exe` 到 `D:\tools\fnm\`
3. 将 `D:\tools\fnm` 加入系统 PATH
4. 设置 `FNM_DIR` 环境变量指向 D 盘：

```powershell
[Environment]::SetEnvironmentVariable("FNM_DIR", "D:\tools\fnm", "User")
```

5. 安装 Node 20 LTS：

```bash
fnm install 20
fnm default 20
```

---

## 问题 3：fnm 出现双重 `node-versions` 嵌套路径

### 现象

安装后目录结构异常：

```
D:/tools/fnm/node-versions/node-versions/v20.20.2/installation
         ↑ FNM_DIR 设置       ↑ fnm 又自建了一层
```

### 根因

首次配置时，`FNM_DIR` 被误设为 `D:\tools\fnm\node-versions`，而 fnm 内部逻辑会在 `FNM_DIR` 下再创建 `node-versions` 子目录，导致双重嵌套。

### 解决方案

将 `FNM_DIR` 修正为 `D:\tools\fnm`（父目录），让 fnm 自己管理 `node-versions` 子目录：

```powershell
[Environment]::SetEnvironmentVariable("FNM_DIR", "D:\tools\fnm", "User")
```

如果已产生嵌套，可清理重装：

```bash
fnm uninstall 20
# 手动删除 D:\tools\fnm\node-versions 目录
fnm install 20
fnm default 20
```

### 正确路径结构

```
D:/tools/fnm/
├── fnm.exe
└── node-versions/          ← fnm 自动创建
    ├── aliases/
    │   └── default -> D:/tools/fnm/node-versions/v20.20.2/installation
    └── v20.20.2/
        └── installation/
            ├── node.exe
            ├── npm
            └── ...
```

---

## 最终可用性检查清单

在新终端中依次执行：

```bash
fnm --version        # 应输出 fnm 版本号
node -v              # 应输出 v20.20.2
npm -v               # 应输出 10.7.0
where.exe node       # 应指向 D:\tools\fnm\...\installation\node.exe
```

全部通过后即可进入 Vite 项目初始化阶段。

---

## 相关文档

- [项目 SDD 规范](../spec/) — OpenAPI、UI 主题、Dashboard 配置规范
