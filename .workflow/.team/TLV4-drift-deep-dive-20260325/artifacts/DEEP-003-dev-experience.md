# DEEP-003: 本地开发体验对比分析

## 执行摘要

本文档对比 repo-cn（Python/Streamlit）与 paperbanana-clean（Go/React）的本地开发体验，识别所有差距并提出改进方案。

---

## 1. 启动体验对比

### 1.1 repo-cn 启动流程

| 特性 | 实现方式 | 用户操作 |
|-----|---------|---------|
| **一键启动** | `win-start.bat` / `mac-start.command` | 双击启动脚本 |
| **Python 自动安装** | winget / Homebrew / GitHub 便携版下载 | 无需预装 |
| **虚拟环境自动创建** | `.venv` 目录自动生成 | 无需手动操作 |
| **依赖自动安装** | pip install from requirements.txt | 首次运行等待 |
| **数据目录自动创建** | `data/PaperBananaBench/{diagram,plot}` | 无需手动操作 |
| **端口冲突自动清理** | 检测并 kill 占用 8501 的进程 | 无需手动操作 |
| **浏览器自动打开** | 3 秒延迟后自动打开 localhost:8501 | 无需手动操作 |
| **中文界面** | 全中文提示信息 | 零门槛 |

**启动脚本特性详解**：

```
win-start.bat 关键特性：
- chcp 65001 设置 UTF-8 编码
- EnableDelayedExpansion 支持复杂变量
- 多路径 Python 搜索（runtime/ → 系统 → winget → 自动下载）
- Python 版本校验 >= 3.10
- GitHub API 查询最新 python-build-standalone 版本
- PowerShell 下载 + tar 解压
- 中国镜像源 pip (pypi.tuna.tsinghua.edu.cn)
- 详细进度提示（[OK]/[..]/[!!]）
```

### 1.2 paperbanana-clean 启动流程

| 特性 | 实现方式 | 用户操作 |
|-----|---------|---------|
| **启动方式** | 手动命令行 | `go run ./cmd/server` |
| **Go 环境** | 无自动安装 | 需预装 Go 1.23+ |
| **前端启动** | 单独命令 | `cd web && npm install && npm run dev` |
| **配置文件** | 需手动创建 | `cp .env.example .env` + 编辑 |
| **数据目录** | 代码自动创建 | 无需手动操作 |
| **端口冲突** | 无自动处理 | 需手动检查 8080/5173 |
| **浏览器打开** | 无自动打开 | 手动访问 localhost:5173 |

**README.md 指引**：

```bash
# Backend:
go run ./cmd/server --config ./configs/config.yaml

# Frontend:
cd web
npm install
npm run dev

# Optional:
set PAPERBANANA_BENCH_ROOT=D:\datasets\PaperBananaBench
```

### 1.3 启动体验差距汇总

| 维度 | repo-cn | paperbanana-clean | 差距评估 |
|-----|---------|-------------------|---------|
| 一键启动 | 完整 | 无 | **严重缺失** |
| 环境自动安装 | Python 自动安装 | Go 需预装 | **中等差距** |
| 双平台支持 | win + mac 脚本 | 无脚本 | **严重缺失** |
| 中文提示 | 全中文 | 英文/无提示 | **中等差距** |
| 浏览器自动打开 | 是 | 否 | **小差距** |
| 端口冲突处理 | 自动清理 | 无 | **小差距** |

---

## 2. 热重载支持对比

### 2.1 repo-cn 热重载

**Streamlit 内置热重载**：
- 代码修改后自动刷新浏览器
- 无需重启服务
- 实时预览 UI 变更

**Hot-Reload Prompts 特性**（`visualize/show_referenced_eval.py`）：
```python
if st.sidebar.button("Re-run Eval (Hot-Reload Prompts)", type="primary"):
    # 重新加载 prompt 文件并运行
```

### 2.2 paperbanana-clean 热重载

**前端 (Vite)**：
- `npm run dev` 启动 Vite 开发服务器
- HMR (Hot Module Replacement) 自动刷新
- API 代理到后端 8080

**后端 (Go)**：
- **无热重载**：代码修改需重启服务
- Config Watcher 仅支持配置热重载：
```go
// cmd/server/main.go
configWatcher := configservice.NewWatcher()
configSvc := configservice.NewServiceWithWatcher(providerRepo, apiKeyRepo, configWatcher)
```

**配置热重载架构**：
```
internal/application/config/watcher.go:
- Watcher.Subscribe() 返回事件通道
- ConfigSSEHandler 推送 SSE 到前端
- 前端订阅 /api/v1/config/events
```

### 2.3 热重载差距汇总

| 维度 | repo-cn | paperbanana-clean | 差距评估 |
|-----|---------|-------------------|---------|
| 前端热重载 | Streamlit 内置 | Vite HMR | **相当** |
| 后端热重载 | Python 自动重载 | 无 | **中等差距** |
| 配置热重载 | 需重启 | SSE 实时推送 | **paperbanana 更优** |
| Prompt 热重载 | 专门实现 | 无 | **小差距** |

---

## 3. 调试便利性对比

### 3.1 repo-cn 调试特性

**Streamlit 内置调试工具**：
- `st.write()` / `st.json()` / `st.code()` 实时输出
- 侧边栏调试控件
- Session State 可视化

**调试界面示例**：
```python
# visualize/show_referenced_eval.py
st.sidebar.subheader("Debug Target")
st.sidebar.info(f"Active: {identifier}\nIndex: {st.session_state.debug_idx}")
if st.sidebar.button("Clear Debug State"):
    st.session_state.clear()
```

**Python 调试优势**：
- `print()` 直接输出到终端
- `breakpoint()` 交互式调试
- IDE 断点调试支持良好

### 3.2 paperbanana-clean 调试特性

**日志系统**：
- Zap structured logging (JSON 格式)
- 中间件请求日志：
```go
// internal/api/middleware/logger.go
func Logger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        logger.Info("request",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", time.Since(start)),
        )
    }
}
```

**Go 调试方式**：
- Delve 调试器 (`dlv debug`)
- IDE 断点支持
- pprof 性能分析

**前端调试**：
- React DevTools
- 浏览器开发者工具
- Vite 错误覆盖层

### 3.3 调试便利性差距汇总

| 维度 | repo-cn | paperbanana-clean | 差距评估 |
|-----|---------|-------------------|---------|
| 即时输出 | st.write() 实时 | 需刷新/查看日志 | **repo-cn 更优** |
| 调试界面 | UI 内置调试控件 | 无内置 UI 调试 | **中等差距** |
| 日志格式 | 文本输出 | JSON 结构化 | **paperbanana 更规范** |
| 断点调试 | pdb/IDE | Delve/IDE | **相当** |

---

## 4. 日志输出对比

### 4.1 repo-cn 日志

**特点**：
- 文本格式，人类可读
- `print()` 直接输出到 Streamlit 界面
- 控制台输出 + 界面输出并行

**示例**：
```python
print("调试：正在导入代理模块...")
print("调试：已导入 PlannerAgent")
print(f"调试：导入错误：{e}")
```

### 4.2 paperbanana-clean 日志

**特点**：
- Zap JSON 结构化日志
- 可解析、可聚合
- 生产环境友好

**示例输出**：
```json
{"level":"info","ts":...,"msg":"request","method":"POST","path":"/api/v1/generate","status":200,"duration":"1.234s"}
```

### 4.3 日志对比汇总

| 维度 | repo-cn | paperbanana-clean | 评估 |
|-----|---------|-------------------|------|
| 格式 | 文本 | JSON 结构化 | **paperbanana 生产更优** |
| 可读性 | 高（中文+人类友好） | 中（需解析） | **repo-cn 开发更友好** |
| 可解析性 | 低 | 高 | **paperbanana 更优** |
| 界面集成 | 直接输出到 UI | 无 | **repo-cn 更便利** |

---

## 5. 缺失功能清单

### 5.1 一键启动脚本（优先级：P0）

**缺失**：
- `win-start.bat` Windows 启动脚本
- `mac-start.command` macOS 启动脚本

**影响**：
- 新用户上手门槛高
- 需要阅读 README 理解启动流程
- 需要手动检查环境依赖

### 5.2 环境自动检测与安装（优先级：P1）

**缺失**：
- Go 环境检测与版本校验
- Node.js 环境检测
- 依赖完整性检查

**影响**：
- 环境问题难以诊断
- 新用户可能遇到晦涩错误

### 5.3 开发调试 UI（优先级：P2）

**缺失**：
- 前端调试面板
- API 请求/响应查看器
- 流水线执行状态可视化

**影响**：
- 调试效率低
- 问题定位困难

### 5.4 中文本地化（优先级：P2）

**缺失**：
- 启动脚本中文提示
- 错误消息中文翻译
- 开发文档中文版

**影响**：
- 国内用户门槛高

### 5.5 端口冲突自动处理（优先级：P3）

**缺失**：
- 端口占用检测
- 自动清理或提示

**影响**：
- 启动失败需手动排查

### 5.6 后端热重载（优先级：P3）

**缺失**：
- Go 代码修改后自动重启
- 类似 Air / CompileDaemon 工具集成

**影响**：
- 开发迭代速度慢

---

## 6. 改进方案

### 6.1 一键启动脚本实现

**Windows (`win-start.bat`)**：

```batch
@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion

echo ==========================================
echo   PaperBanana 开发服务器
echo ==========================================
echo.

REM 检查 Go 环境
where go >nul 2>&1
if errorlevel 1 (
    echo [!!] 未检测到 Go，请安装 Go 1.23+
    echo     下载地址: https://go.dev/dl/
    pause
    exit /b 1
)

REM 检查 Go 版本
for /f "tokens=3" %%v in ('go version 2^>^&1') do set GO_VER=%%v
echo [OK] Go %GO_VER% 已安装

REM 检查 Node.js
where node >nul 2>&1
if errorlevel 1 (
    echo [!!] 未检测到 Node.js，请安装 Node.js 18+
    echo     下载地址: https://nodejs.org/
    pause
    exit /b 1
)
echo [OK] Node.js 已安装

REM 检查配置文件
if not exist ".env" (
    echo [..] 创建 .env 配置文件...
    copy .env.example .env >nul
    echo [OK] 已创建 .env，请编辑配置 API Key
)

REM 创建数据目录
if not exist "data\PaperBananaBench\diagram" mkdir "data\PaperBananaBench\diagram"
if not exist "data\PaperBananaBench\plot" mkdir "data\PaperBananaBench\plot"

REM 检查端口
set PORT=8080
for /f "tokens=5" %%A in ('netstat -ano 2^>nul ^| findstr ":%PORT% " ^| findstr "LISTENING"') do (
    echo [!!] 端口 %PORT% 被占用 (PID: %%A)
    choice /C YN /M "是否终止占用进程"
    if errorlevel 2 exit /b 1
    taskkill /F /PID %%A >nul 2>&1
    timeout /t 1 /nobreak >nul
)

REM 启动后端
echo.
echo [..] 启动后端服务 (端口 8080)...
start "PaperBanana Backend" cmd /c "go run ./cmd/server"

REM 等待后端启动
timeout /t 3 /nobreak >nul

REM 启动前端
echo [..] 启动前端开发服务器...
cd web
if not exist "node_modules" (
    echo [..] 安装前端依赖...
    npm install
)
start "PaperBanana Frontend" cmd /c "npm run dev"

REM 打开浏览器
timeout /t 3 /nobreak >nul
start http://localhost:5173

echo.
echo ==========================================
echo   服务已启动！
echo   前端: http://localhost:5173
echo   后端: http://localhost:8080
echo   关闭此窗口不会停止服务
echo ==========================================
pause
```

**macOS (`mac-start.command`)**：

```bash
#!/bin/bash
set -euo pipefail

# 颜色定义
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

banner() { echo -e "\n${CYAN}${BOLD}$1${NC}"; }
ok() { echo -e "  ${GREEN}✔${NC} $1"; }
fail() { echo -e "  ${RED}✖${NC} $1"; }

cd "$(dirname "$0")"

banner "=========================================="
banner "  PaperBanana 开发服务器"
banner "=========================================="
echo ""

# 检查 Go
if ! command -v go &>/dev/null; then
    fail "未检测到 Go，请安装 Go 1.23+"
    echo "  下载地址: https://go.dev/dl/"
    exit 1
fi
ok "Go $(go version | awk '{print $3}') 已安装"

# 检查 Node.js
if ! command -v node &>/dev/null; then
    fail "未检测到 Node.js，请安装 Node.js 18+"
    echo "  下载地址: https://nodejs.org/"
    exit 1
fi
ok "Node.js $(node -v) 已安装"

# 检查配置
if [ ! -f ".env" ]; then
    echo "  创建 .env 配置文件..."
    cp .env.example .env
    ok "已创建 .env，请编辑配置 API Key"
fi

# 创建数据目录
mkdir -p data/PaperBananaBench/{diagram,plot}

# 检查端口
PORT=8080
if lsof -i:$PORT &>/dev/null; then
    PID=$(lsof -ti:$PORT)
    echo "  端口 $PORT 被占用 (PID: $PID)"
    read -p "  是否终止占用进程? [y/N] " choice
    [[ "$choice" =~ ^[Yy]$ ]] && kill -9 $PID
fi

# 启动后端
banner "启动后端服务 (端口 8080)..."
osascript -e 'tell app "Terminal" to do script "cd \"'$(pwd)'\" && go run ./cmd/server"'

# 等待后端
sleep 3

# 启动前端
banner "启动前端开发服务器..."
cd web
[ ! -d "node_modules" ] && npm install
osascript -e 'tell app "Terminal" to do script "cd \"'$(pwd)'\" && npm run dev"'

# 打开浏览器
sleep 3
open http://localhost:5173

banner "=========================================="
banner "  服务已启动！"
banner "  前端: http://localhost:5173"
banner "  后端: http://localhost:8080"
banner "=========================================="
```

### 6.2 后端热重载集成

**使用 Air 实现热重载**：

```toml
# .air.toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["web", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html", "yaml"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = true
```

**安装与运行**：
```bash
# 安装 Air
go install github.com/air-verse/air@latest

# 使用热重载启动
air
```

### 6.3 开发调试面板

**前端调试组件建议**：

```tsx
// web/src/components/DevTools.tsx
import { useState } from 'react';

export function DevTools() {
  const [isOpen, setIsOpen] = useState(false);

  if (process.env.NODE_ENV !== 'development') {
    return null;
  }

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="bg-gray-800 text-white px-4 py-2 rounded-lg shadow-lg"
      >
        Dev Tools
      </button>
      {isOpen && (
        <div className="mt-2 bg-white border rounded-lg shadow-xl p-4 w-96">
          <h3 className="font-bold mb-2">API Requests</h3>
          {/* 请求日志显示 */}
          <h3 className="font-bold mb-2 mt-4">State</h3>
          {/* 状态树显示 */}
        </div>
      )}
    </div>
  );
}
```

---

## 7. 改进优先级排序

| 优先级 | 改进项 | 工作量 | 收益 |
|--------|--------|--------|------|
| P0 | 一键启动脚本 (win-start.bat, mac-start.command) | 2-4h | 高 |
| P1 | 环境检测与提示 | 1-2h | 中 |
| P2 | 中文本地化（启动脚本） | 1h | 中 |
| P2 | 开发调试面板 | 4-8h | 中 |
| P3 | 后端热重载 (Air) | 1h | 低 |
| P3 | 端口冲突自动处理 | 1h | 低 |

---

## 8. 结论

paperbanana-clean 相比 repo-cn 在本地开发体验上存在明显差距，主要集中在：

1. **启动便利性**：repo-cn 提供一键启动脚本，paperbanana-clean 需要手动执行多个命令
2. **环境友好度**：repo-cn 自动安装 Python，paperbanana-clean 需预装 Go/Node.js
3. **调试便利性**：Streamlit 内置调试 UI，paperbanana-clean 缺乏前端调试工具
4. **中文本地化**：repo-cn 全中文，paperbanana-clean 以英文为主

**建议优先实施**：
1. 创建 `win-start.bat` 和 `mac-start.command` 一键启动脚本
2. 集成 Air 实现后端热重载
3. 添加前端开发调试面板

这些改进将显著降低新用户上手门槛，提升开发效率。
