# 开发环境安装指南

本文档说明在运行 `make setup` 之前需要安装的前置条件。

## 前置条件

| 工具 | 必须性 | 说明 |
|------|--------|------|
| **Go** | ✅ 必须 | Go 编程语言（版本 1.22+） |
| **make** | ✅ 必须 | 构建工具 |
| **Git** | ✅ 必须 | 版本控制 |
| **uv** | ✅ 必须 | Python 包管理工具（用于安装 pre-commit） |

---

## Windows 安装步骤

### 方式一：使用 winget（推荐）

```powershell
# 1. 安装 Go
winget install GoLang.Go

# 2. 安装 make
winget install GnuWin32.Make

# 3. 安装 Git
winget install Git.Git

# 4. 安装 uv（Python 包管理器）
winget install astral-sh.uv

# 5. 重启终端，使环境变量生效

# 6. 验证安装
go version
make --version
git --version
uv --version
```

### 方式二：使用 PowerShell 脚本

```powershell
# 安装 uv（推荐方式）
irm https://astral.sh/uv/install.ps1 | iex

# 安装 Go
# 下载：https://go.dev/dl/

# 安装 make
# 下载：http://gnuwin32.sourceforge.net/packages/make.htm

# 安装 Git
# 下载：https://git-scm.com/download/win
```

### 方式三：使用 Scoop

```powershell
# 安装 Scoop（如果未安装）
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex

# 安装工具
scoop install go make git uv

# 验证安装
go version
make --version
git --version
uv --version
```

---

## macOS 安装步骤

```bash
# 1. 安装 Homebrew（如果未安装）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 2. 安装工具
brew install go make git uv

# 3. 验证安装
go version
make --version
git --version
uv --version
```

---

## Linux (Ubuntu/Debian) 安装步骤

```bash
# 1. 更新包列表
sudo apt update

# 2. 安装基础工具
sudo apt install -y golang-go make git

# 3. 安装 uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# 4. 重新加载 shell 配置
source ~/.bashrc

# 5. 验证安装
go version
make --version
git --version
uv --version
```

### 安装最新版 Go（推荐）

Ubuntu 默认的 Go 版本可能较旧，建议安装最新版：

```bash
# 下载并安装 Go 1.22
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# 添加到 PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证
go version
```

---

## 关于 uv

uv 是一个现代化的 Python 包管理器，比 pip 快 10-100 倍。

### uv 安装方式

| 系统 | 安装命令 |
|------|----------|
| Windows (winget) | `winget install astral-sh.uv` |
| Windows (PowerShell) | `irm https://astral.sh/uv/install.ps1 \| iex` |
| macOS/Linux | `curl -LsSf https://astral.sh/uv/install.sh \| sh` |

### uv 常用命令

```bash
# 安装工具（全局）
uv tool install pre-commit

# 查看已安装工具
uv tool list

# 更新工具
uv tool upgrade pre-commit

# 卸载工具
uv tool uninstall pre-commit
```

---

## 验证前置条件

安装完成后，运行以下命令验证：

```bash
# 检查所有前置条件是否满足
make check-prerequisites
```

如果输出 `✅ 所有前置条件已安装`，则可以继续运行 `make setup`。

---

## 运行 make setup

前置条件安装完成后，运行：

```bash
make setup
```

这将自动安装：
- golangci-lint（代码检查）
- goimports（import 排序）
- gosec（安全扫描）
- pre-commit（Git 钩子，使用 uv 安装）
- 项目依赖

---

## 常见问题

### Q: Windows 上 make 命令找不到？

**方案一**：使用完整路径
```powershell
C:\Program Files (x86)\GnuWin32\bin\make.exe setup
```

**方案二**：添加到 PATH
1. 打开「系统属性」→「环境变量」
2. 在「系统变量」中找到 `Path`，点击「编辑」
3. 添加 `C:\Program Files (x86)\GnuWin32\bin`
4. 重启终端

### Q: uv 命令找不到？

Windows 上安装 uv 后需要重启终端。如果仍然找不到，手动添加到 PATH：

```
%USERPROFILE%\.local\bin
%USERPROFILE%\.cargo\bin
```

### Q: Go 版本太旧？

检查 Go 版本：
```bash
go version
```

如果版本低于 1.22，请从 [Go 官网](https://go.dev/dl/) 下载最新版。

### Q: pre-commit 安装失败？

确保 uv 已正确安装：
```bash
uv --version
```

手动安装 pre-commit：
```bash
uv tool install pre-commit
```

---

## 快速开始清单

- [ ] 安装 Go
- [ ] 安装 make
- [ ] 安装 Git
- [ ] 安装 uv
- [ ] 运行 `make check-prerequisites` 验证
- [ ] 运行 `make setup` 安装开发工具
- [ ] 运行 `make run` 启动项目
