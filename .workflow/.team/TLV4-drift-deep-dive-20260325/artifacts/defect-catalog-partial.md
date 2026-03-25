# 深度缺陷清单 (初步)

**生成时间**: 2026-03-25
**状态**: 初步分析 (后台 agents 仍在运行)

---

## P0 - Critical (阻塞性缺陷)

### DEFECT-001: 图片无法正确展示

**位置**:
- `internal/domain/agent/types.go:87-95` (Artifact 缺少 asset_id)
- `internal/application/agents/visualizer/agent.go:420-431` (使用 memory:// URI)
- `web/src/hooks/useGenerate.ts:170-176` (前端尝试获取不存在的 asset_id)

**描述**:
生成的图片存储在 `Artifact.Bytes` 中，但:
1. 没有持久化到 Asset Store
2. SSE 结果中没有 asset_id
3. 前端 URL 构建逻辑错误 (缺少 project_id)

**影响**: 所有生成结果无法展示

**修复方案**:
1. Runner 完成后将 artifacts 持久化到 AssetService
2. 在 SSE 结果中返回 `asset_id` 和 `project_id`
3. 前端使用正确的 URL: `/api/v1/assets/{project_id}/{asset_id}/download`

---

### DEFECT-002: 批处理结果内存存储

**位置**:
- `internal/application/orchestrator/batch.go` (需要确认)

**描述**:
之前的报告已提及: 批处理结果仅在内存中，服务器重启后丢失。

**影响**: 批量生成的所有结果丢失

**修复方案**: 将批处理结果持久化到 SQLite

---

### DEFECT-003: Plot 执行安全漏洞

**位置**:
- `internal/application/agents/visualizer/plot_executor.go`

**描述**:
直接执行 `exec.Command("python3", "-c", code)` 允许任意代码执行。

**影响**: 严重安全风险

**修复方案**:
1. 使用沙箱容器 (Docker/gVisor)
2. 或使用受限 Python 环境 (RestrictedPython)

---

## P1 - High (重要缺陷)

### DEFECT-004: 缺少优雅降级

**位置**:
- `internal/application/orchestrator/runner.go`

**描述**:
当 LLM 调用失败时，整个 session 失败。repo-cn 会继续尝试其他路径。

**对比 repo-cn**:
```python
# repo-cn 有 fallback 逻辑
if result and result[0] and result[0] != "Error":
    # 成功
else:
    # 继续处理，不中断
```

**修复方案**: 在 runner 中实现 fallback 逻辑

---

### DEFECT-005: 缺少 Lite Retrieval Mode

**位置**:
- `internal/application/agents/retriever/`

**描述**:
repo-cn 有 `lite` 模式，可节省 96% token。paperbanana-clean 没有这个功能。

**影响**: 高 token 消耗

**修复方案**: 移植 lite retrieval 模式

---

### DEFECT-006: SSE 接口不完整

**位置**:
- `web/src/lib/sse.ts:34-42`
- `internal/api/handlers/generate.go`

**描述**:
前端 `ResultEvent` 接口与后端发送的数据不匹配:
- 缺少 `asset_id`
- 缺少 `project_id`
- `summary` 字段从未被填充

**修复方案**: 统一前后端接口定义

---

## P2 - Medium (中等缺陷)

### DEFECT-007: 健康检查不完整

**位置**:
- `internal/api/router.go:56-61`

**描述**:
`/health` 和 `/ready` 始终返回 OK，没有检查数据库、LLM 连接等依赖。

**修复方案**: 实现真正的健康检查

---

### DEFECT-008: Session 状态膨胀

**位置**:
- `internal/infrastructure/persistence/sqlite/session_repository.go`

**描述**:
每次 critic 迭代都保存完整的 artifacts，导致数据库膨胀。

**修复方案**: 实现状态清理/压缩策略

---

## 对比 repo-cn 的优势特性 (缺失)

| 特性 | repo-cn | paperbanana-clean | 状态 |
|------|---------|-------------------|------|
| Base64 内联传输 | ✅ | ❌ | 缺失 |
| Lite Retrieval Mode | ✅ 96% 节省 | ❌ | 缺失 |
| Critic Early Stop | ✅ | ⚠️ 部分 | 需完善 |
| Fallback on Error | ✅ | ❌ | 缺失 |
| 零配置启动脚本 | ✅ | ❌ | 缺失 |

---

## 待后台 Agents 完成的分析

1. **深度缺陷挖掘** (a9d3961332ba3b783) - 运行中
2. **图片展示流程审查** (afb1dc1877c9e1be5) - 运行中
3. **安全漏洞扫描** (a5f59303a25f15f77) - 运行中
4. **批处理持久化审查** (a5e2218f9768ddd4d) - 运行中
5. **Supervisor** (a7d66fe1df3e0e8e4) - 运行中

---

*此报告将在所有 agents 完成后更新*
