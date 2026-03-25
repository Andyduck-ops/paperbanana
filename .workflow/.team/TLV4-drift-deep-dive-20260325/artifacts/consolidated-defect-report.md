# 综合缺陷报告 - paperbanana-clean vs repo-cn

**生成时间**: 2026-03-25 22:30
**分析方法**: 多视角代码审查 + 架构对比
**状态**: 完成

---

## 执行摘要

**发现总数**: 12 个缺陷
**P0 Critical**: 4 个
**P1 High**: 5 个
**P2 Medium**: 3 个

**核心问题**: paperbanana-clean 的架构设计优于 repo-cn，但关键运行逻辑存在断层，特别是**图片展示流程**完全断裂。

---

## P0 - Critical (阻塞性缺陷)

### 🔴 DEFECT-001: 图片展示流程断裂 [新发现]

**严重程度**: Critical
**影响**: 100% 用户无法看到生成结果

**根本原因**:
1. 后端 `Artifact` 结构体没有 `asset_id` 字段
2. Visualizer 生成的图片使用 `memory://` URI，未持久化
3. Runner 没有将 artifacts 存储到 AssetService
4. 前端 SSE 接口缺少 `asset_id` 和 `project_id`
5. 前端 URL 构建使用错误的路径格式

**代码位置**:
| 文件 | 行号 | 问题 |
|------|------|------|
| `internal/domain/agent/types.go` | 87-95 | 缺少 asset_id |
| `internal/application/agents/visualizer/agent.go` | 420-431 | memory:// URI |
| `internal/application/orchestrator/runner.go` | - | 无持久化逻辑 |
| `web/src/lib/sse.ts` | 34-42 | 接口不完整 |
| `web/src/hooks/useGenerate.ts` | 170-176 | 类型断言失败 |
| `web/src/components/ArtifactPreview.tsx` | 21-25 | URL 格式错误 |

**数据流断裂点**:
```
Visualizer 生 Bytes
    ↓
Artifact.Bytes (内存中)
    ↓
❌ 断裂点: 没有调用 AssetService.Store()
    ↓
SSE 发送 JSON (无 asset_id)
    ↓
前端收到 artifact.asset_id = undefined
    ↓
imageUrl = null
    ↓
图片不显示
```

**修复方案**:
```go
// 在 runner.go 完成后添加:
func (r *Runner) persistArtifacts(ctx context.Context, output AgentOutput, projectID string) error {
    for _, artifact := range output.GeneratedArtifacts {
        if len(artifact.Bytes) > 0 {
            assetID, err := r.assetStore.Store(ctx, projectID, artifact.Bytes, artifact.MIMEType)
            if err != nil {
                return err
            }
            artifact.ID = assetID  // 或添加 AssetID 字段
        }
    }
    return nil
}
```

**工作量**: 2-3 天

---

### 🔴 DEFECT-002: 无优雅降级机制

**严重程度**: Critical
**影响**: 任何 LLM 调用失败导致整个 session 失败

**对比 repo-cn**:
```python
# repo-cn 的降级逻辑
if image_result == "Error":
    # 继续使用文本描述
    return text_fallback
```

**代码位置**: `internal/application/orchestrator/runner.go`

**修复方案**: 实现 fallback 逻辑，部分失败时继续执行

**工作量**: 2-3 天

---

### 🔴 DEFECT-003: Plot 执行任意代码漏洞

**严重程度**: Critical (安全)
**影响**: 远程代码执行风险

**代码位置**: `internal/application/agents/visualizer/plot_executor.go`

**当前代码**:
```go
cmd := exec.Command("python3", "-c", code)  // 危险！
```

**修复方案**:
1. Docker 容器隔离
2. 或使用 RestrictedPython
3. 或禁用 plot 模式，仅保留 diagram 模式

**工作量**: 2 天

---

### 🔴 DEFECT-004: 批处理结果内存存储

**严重程度**: Critical
**影响**: 服务器重启丢失所有批量生成结果

**代码位置**: `internal/api/handlers/batch.go`

**修复方案**: 持久化到 SQLite，关联 session

**工作量**: 1 天

---

## P1 - High (重要缺陷)

### 🟠 DEFECT-005: 缺少 Lite Retrieval Mode

**对比**: repo-cn 有此功能，节省 96% token

**代码位置**: `internal/application/agents/retriever/`

**修复方案**: 移植 repo-cn 的精简检索逻辑

**工作量**: 1 天

---

### 🟠 DEFECT-006: SSE 接口不完整

**问题**:
- 前后端接口定义不一致
- 缺少 `asset_id`, `project_id`
- `summary` 字段从未填充

**修复方案**: 统一接口，使用共享类型定义

**工作量**: 1 天

---

### 🟠 DEFECT-007: 健康检查不完整

**代码位置**: `internal/api/router.go:56-61`

**当前实现**:
```go
router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})  // 总是返回 ok
})
```

**修复方案**: 检查 DB、LLM、Redis 连接

**工作量**: 0.5 天

---

### 🟠 DEFECT-008: Session 状态膨胀

**问题**: 每次迭代保存完整 artifacts，导致 DB 膨胀

**修复方案**: 实现压缩策略，只保留最新 artifacts

**工作量**: 1 天

---

### 🟠 DEFECT-009: 可观测性断点

**问题**:
- Prometheus 端点定义但未连接
- 无结构化 LLM 请求日志
- 无分布式追踪

**工作量**: 2 天

---

## P2 - Medium (中等缺陷)

### 🟡 DEFECT-010: main.go 紧耦合

**问题**: 375 行手动 DI，难以测试

**工作量**: 2 天

---

### 🟡 DEFECT-011: 缺少请求验证 Schema

**问题**: 无 JSON Schema 验证，错误消息不一致

**工作量**: 1 天

---

### 🟡 DEFECT-012: 缺少成本透明度

**问题**: UI 不显示 token 成本估算

**对比**: repo-cn 无此功能，但用户期望有

**工作量**: 0.5 天

---

## 功能对比矩阵

| 功能 | repo-cn | paperbanana-clean | 状态 |
|------|---------|-------------------|------|
| **图片展示** | ✅ Base64 内联 | ❌ 断裂 | **需修复** |
| **优雅降级** | ✅ | ❌ | **需移植** |
| **Lite Retrieval** | ✅ 96% 节省 | ❌ | **需移植** |
| **Critic 早停** | ✅ | ⚠️ 部分 | 需完善 |
| **Plot 执行** | ✅ 无沙箱 | ❌ 无沙箱 | **都有风险** |
| **状态恢复** | ❌ | ✅ | Go 优势 |
| **多 Provider** | 2 个 | 16+ | Go 优势 |
| **类型安全** | ❌ | ✅ | Go 优势 |

---

## 修复优先级建议

### 阶段 1: 止血 (P0, 3-5 天)

| 任务 | 工作量 | 依赖 |
|------|--------|------|
| DEFECT-001 图片展示 | 2-3 天 | - |
| DEFECT-004 批处理持久化 | 1 天 | - |

### 阶段 2: 稳定 (P0-P1, 3-4 天)

| 任务 | 工作量 | 依赖 |
|------|--------|------|
| DEFECT-002 优雅降级 | 2-3 天 | - |
| DEFECT-003 Plot 安全 | 2 天 | - |
| DEFECT-005 Lite Retrieval | 1 天 | - |

### 阶段 3: 增强 (P1-P2, 3-4 天)

| 任务 | 工作量 | 依赖 |
|------|--------|------|
| DEFECT-006 SSE 接口 | 1 天 | DEFECT-001 |
| DEFECT-007 健康检查 | 0.5 天 | - |
| DEFECT-008 状态压缩 | 1 天 | - |
| DEFECT-009 可观测性 | 2 天 | - |

**总工作量**: 9-13 天

---

## 结论

paperbanana-clean 的架构基础良好（类型安全、多 Provider 支持、状态恢复），但存在一个**阻塞性缺陷**：图片无法展示。这是数据流设计问题，需要系统性地修复:

1. **立即**: 修复图片展示流程 (DEFECT-001)
2. **短期**: 移植 repo-cn 的运行逻辑 (降级、Lite Retrieval)
3. **中期**: 完善安全性和可观测性

不建议替换 Go 实现，因为其架构优势明显。修复工作量可控，预计 2 周内可完成关键修复。

---

*报告生成: Team Lifecycle v4 深度分析*
