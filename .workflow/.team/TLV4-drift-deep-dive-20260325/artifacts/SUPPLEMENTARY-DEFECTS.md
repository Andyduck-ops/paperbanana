# 补充缺陷报告 - 数据与启动问题

**生成时间**: 2026-03-25 23:00
**状态**: 紧急发现

---

## 🔴 CRITICAL-NEW-001: PaperBananaBench 数据缺失

**严重程度**: Critical
**影响**: 无法进行参考图像检索，生成质量严重下降

### 发现

paperbanana-clean 的 `data/PaperBananaBench/` 目录数据为空：

```
paperbanana-clean/data/PaperBananaBench/diagram/
├── agent_selected_12.json  (3 bytes - 内容: "[]")
└── ref.json                (3 bytes - 内容: "[]")
```

而 repo-cn 的同样目录包含：
```
repo-cn/data/PaperBananaBench/diagram/
├── ref.json                (4,496,771 bytes - 4.5MB!)
├── test.json               (4,653,548 bytes)
├── agent_selected_12.json  (3 bytes)
└── images/                 (241+ images)
    ├── xxx_diagram.jpg     (每个 100-700KB)
    └── ...
```

### 影响

1. **无参考图像**: Planner 无法加载示例图片
2. **无基准数据**: Retriever 无法检索相似案例
3. **生成质量下降**: 没有 few-shot 示例，LLM 生成质量大打折扣

### 根本原因

`.gitignore` 可能排除了大型数据文件，但没有提供数据获取脚本。

### 修复方案

1. 添加数据下载脚本 `scripts/download-benchmark.sh`
2. 在 README 中说明数据获取方式
3. 或使用 Git LFS 管理大文件

---

## 🔴 CRITICAL-NEW-002: 无一键启动脚本

**严重程度**: High
**影响**: 开发者体验差，上手成本高

### 对比

**repo-cn (Python)**:
- `win-start.bat` - Windows 双击启动
- `mac-start.command` - macOS 双击启动
- 自动检测/安装 Python
- 自动创建虚拟环境
- 自动安装依赖
- 零配置启动

**paperbanana-clean (Go)**:
- `docker-compose.yml` - 需要 Docker
- 没有本地启动脚本
- 需要手动：
  - 安装 Go
  - 编译后端
  - 编译前端
  - 配置数据库

### 修复方案

创建 `start.bat` / `start.sh`:
```batch
@echo off
echo Starting PaperBanana...
REM Check Go
REM Build backend
REM Build frontend
REM Start server
```

---

## 🔴 CRITICAL-NEW-003: 失败后无继续执行逻辑

**严重程度**: High
**影响**: 任何阶段失败导致整个任务失败

### 当前代码 (runner.go:214-224)

```go
output, err := stageAgent.Execute(stageCtx, stageInput)
if err != nil {
    cancel()
    if cleanupErr := stageAgent.Cleanup(ctx); cleanupErr != nil {
        err = errors.Join(err, cleanupErr)
    }
    // 直接返回错误，无 fallback
    return r.finishStageError(ctx, tracker, publisher, stage, stageInput, startedAt, err, stageAgent)
}
```

### 对比 repo-cn

```python
# repo-cn 有降级逻辑
try:
    result = agent.execute(input)
except Exception as e:
    if can_fallback:
        result = fallback_result  # 使用降级结果
        continue_pipeline = True
    else:
        raise
```

### 修复方案

添加 Fallback 策略：
```go
type FallbackStrategy int

const (
    FallbackAbort FallbackStrategy = iota
    FallbackContinue
    FallbackRetry
)

func (r *Runner) handleStageError(stage StageName, err error) FallbackStrategy {
    category := ClassifyError(err)
    switch category {
    case ErrCodeLLMTimeout:
        return FallbackRetry
    case ErrCodeImageGeneration:
        return FallbackContinue  // 使用文本输出
    default:
        return FallbackAbort
    }
}
```

---

## 🟠 HIGH-NEW-001: Reference Image 加载可能失败静默

**严重程度**: High
**影响**: Planner 可能无示例图片

### 代码位置

`internal/application/agents/planner/agent.go:199-220`

```go
func (a *Agent) loadExampleImage(mode domainagent.VisualMode, path string) ([]byte, string, error) {
    resolved, err := a.resolveExamplePath(mode, path)
    if err != nil {
        return nil, "", err  // 错误直接返回
    }
    // ...
}
```

如果 `PaperBananaBench` 数据缺失，图片加载失败，但可能被静默处理。

---

## 🟠 HIGH-NEW-002: Context 超时后资源未清理

**严重程度**: High
**影响**: Plot 执行时 Python 进程可能泄露

### 代码位置

`internal/application/agents/visualizer/plot_executor.go`

当 context 超时后，`cmd.Wait()` 可能不会被调用，导致 Python 进程变成孤儿进程。

---

## 功能对比补充

| 功能 | repo-cn | paperbanana-clean | 状态 |
|------|---------|-------------------|------|
| **基准数据** | ✅ 4.5MB ref.json + images | ❌ 空文件 | **严重缺失** |
| **一键启动** | ✅ 双击启动 | ❌ 无脚本 | **缺失** |
| **零配置** | ✅ 自动安装依赖 | ❌ 手动配置 | **缺失** |
| **失败降级** | ✅ 继续/降级 | ❌ 直接失败 | **需移植** |
| **孤儿进程清理** | ⚠️ 基础 | ⚠️ 可能有泄露 | 都需改进 |

---

## 修复优先级更新

### 立即修复 (新增)

| 任务 | 工作量 |
|------|--------|
| 下载基准数据脚本 | 0.5 天 |
| 创建一键启动脚本 | 0.5 天 |
| 添加 Fallback 逻辑 | 2 天 |

### 更新后的总工作量

| 阶段 | 原估算 | 新增 | 总计 |
|------|--------|------|------|
| 止血 | 3-5 天 | 1 天 | 4-6 天 |
| 核心修复 | 5-7 天 | 2 天 | 7-9 天 |
| 加固完善 | 5-6 天 | - | 5-6 天 |
| **总计** | **13-18 天** | **3 天** | **16-21 天** |

---

## 建议行动

### 立即

1. **复制基准数据**: 从 repo-cn 复制 `PaperBananaBench` 数据
2. **创建启动脚本**: 参考 repo-cn 的 `win-start.bat`

### 短期

1. **添加 Fallback 逻辑**: 在 runner.go 中实现
2. **数据下载脚本**: 自动获取基准数据

---

*报告生成: 补充调查*
