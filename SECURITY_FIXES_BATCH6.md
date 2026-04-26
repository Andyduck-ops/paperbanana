# Batch 6: 安全与持久化修复说明

## 修复概述

本次修复包含三个关键安全问题/缺陷的修复：

1. **CRITICAL-SECURITY-001**: Plot RCE 漏洞修复
2. **CRITICAL-BATCH-001**: 批处理结果持久化
3. **CRITICAL-EXPORT-001**: Export 字段名一致性检查

---

## 1. CRITICAL-SECURITY-001: Plot RCE 漏洞修复

### 问题描述
`plot_executor.go` 中的 Python 代码执行存在远程代码执行 (RCE) 漏洞。原始代码直接执行用户输入的 Python 代码：

```python
namespace = {}
exec(code, namespace)  # 危险！直接执行用户代码
```

### 修复措施

#### 1.1 输入验证 (白名单策略)
- 添加 `validatePlotCode()` 函数，使用白名单策略验证代码
- 只允许 `matplotlib` 及其子模块、`numpy` 的导入
- 禁止危险模块：`os`, `sys`, `subprocess`, `importlib`, `socket`, `urllib` 等
- 禁止危险函数：`exec()`, `eval()`, `compile()`, `__import__()`
- 禁止文件操作：`open()`, `file()`
- 禁止魔术命令：`!`, `%`

```go
var allowedModules = map[string]bool{
    "matplotlib":       true,
    "matplotlib.pyplot": true,
    "numpy":             true,
    // ...
}
```

#### 1.2 超时控制
- 默认执行超时：30 秒
- 使用 Go context 进行超时控制
- Python 端同时设置信号超时作为双重保护

```go
const defaultExecutionTimeout = 30 * time.Second
```

#### 1.3 资源限制
- 内存限制：512MB（通过环境变量传递）
- Python 端使用 `resource.setrlimit()` 设置内存上限
- 在 Unix 系统上通过 `SysProcAttr` 设置进程资源限制

```go
const maxMemoryMB = 512
```

#### 1.4 禁用危险模块
Python 执行脚本使用受限的 `__builtins__`：

```python
safe_globals = {
    '__builtins__': {
        'abs': abs, 'all': all, 'any': any,  # 仅允许安全函数
        # ... 危险函数如 exec, eval 被完全移除
    },
    'plt': plt,
    'np': np,
    # ...
}
```

#### 1.5 沙箱化执行
- 创建隔离的命名空间执行用户代码
- 仅暴露必要的 matplotlib 和 numpy API
- 使用 `matplotlib.use('Agg')` 确保无头模式运行

### 安全关键点
- ✅ 绝不直接执行用户输入的代码
- ✅ 所有输入都需要验证（白名单而非黑名单）
- ✅ 限制执行时间和资源
- ✅ 使用受限的 Python 环境

---

## 2. CRITICAL-BATCH-001: 批处理结果持久化

### 问题描述
批处理结果仅存储在内存中，在以下情况下会丢失：
- 应用程序重启
- 内存不足
- 进程崩溃

### 修复措施

#### 2.1 创建 SQLite 持久化存储
新建文件：`internal/infrastructure/persistence/batch_result_store.go`

```go
type BatchResultStore struct {
    db *sql.DB
}
```

#### 2.2 数据库表结构
```sql
CREATE TABLE batch_results (
    batch_id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    successful INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    results_json TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    duration_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 2.3 实现的功能
- `Save()`: 保存/更新批处理结果（支持 UPSERT）
- `Get()`: 按 ID 获取批处理结果
- `List()`: 分页列出批处理结果
- `Delete()`: 删除批处理结果
- `CleanupOldResults()`: 清理旧数据

#### 2.4 进度保存机制
在 `batch_runner.go` 中添加：
- `BatchProgress` 结构体跟踪中间进度
- 每个候选完成后立即更新进度
- 支持从持久化存储恢复进度

```go
type BatchProgress struct {
    BatchID     string                        
    Status      domainagent.RunStatus         
    Total       int                           
    Completed   int                           
    Failed      int                           
    Results     []domainagent.CandidateResult 
    StartedAt   time.Time                     
    UpdatedAt   time.Time                     
}
```

#### 2.5 集成到 BatchRunner
```go
func (r *BatchRunner) updateProgress(progress *BatchProgress) {
    // 内存更新
    r.progress[progress.BatchID] = progress
    
    // 持久化到数据库
    if r.resultStore != nil {
        r.resultStore.Save(tempResult)
    }
}
```

### 容错改进
- ✅ 中间进度持久化，支持故障恢复
- ✅ 结果自动备份到数据库
- ✅ 支持查询历史批处理结果
- ✅ 自动清理过期数据

---

## 3. CRITICAL-EXPORT-001: Export 字段名一致性检查

### 检查结果
经过对以下文件的检查：
- `web/src/components/ExportModal.tsx`
- `web/src/i18n/locales/en.json`
- `web/src/i18n/locales/zh.json`
- `web/src/lib/export.ts`

**结论**：Export 字段名前后端**已经一致**，无需修改。

### 字段映射关系
| 字段用途 | 翻译键 | 英文 | 中文 |
|---------|--------|------|------|
| 标题 | export.title | Export | 导出 |
| 格式 | export.format | Format | 格式 |
| DPI | export.dpi | Resolution (DPI) | 分辨率 (DPI) |
| DPI范围 | export.dpiRange | Range | 范围 |
| 下载 | export.download | Download | 下载 |
| 导出中 | export.exporting | Exporting... | 导出中... |
| 取消 | common.cancel | Cancel | 取消 |

---

## 文件变更列表

### 修改的文件
1. `internal/application/agents/visualizer/plot_executor.go` - RCE 漏洞修复
2. `internal/application/orchestrator/batch_runner.go` - 添加进度持久化
3. `internal/domain/agent/types.go` - 添加 UpdatedAt 字段

### 新增的文件
1. `internal/infrastructure/persistence/batch_result_store.go` - SQLite 持久化实现

---

## 测试建议

### RCE 漏洞修复测试
```python
# 以下代码应该被拒绝执行
code1 = "import os; os.system('ls')"  # 应该失败
code2 = "exec('print(1)')"              # 应该失败
code3 = "open('/etc/passwd')"          # 应该失败

# 以下代码应该正常执行
code4 = "import matplotlib.pyplot as plt; plt.plot([1,2,3])"  # 应该成功
```

### 批处理持久化测试
1. 启动批量生成任务
2. 强制重启应用程序
3. 验证可以从数据库恢复结果

---

## 后续建议

1. **定期安全审计**：对所有执行用户代码的地方进行审计
2. **日志监控**：记录所有代码执行的安全事件
3. **资源限制调优**：根据实际使用情况调整内存限制
4. **数据库备份**：定期备份 SQLite 数据库文件
