# 最终综合缺陷报告 - paperbanana-clean 深度分析

**生成时间**: 2026-03-25 22:45
**分析方法**: 5 个并行 Agent + 人工审核
**状态**: 完成

---

## 执行摘要

| 指标 | 数值 |
|------|------|
| **总缺陷数** | 29 个 |
| **P0 Critical** | 8 个 |
| **P1 High** | 12 个 |
| **P2 Medium** | 6 个 |
| **P3 Low** | 3 个 |
| **估算修复时间** | 15-20 天 |

### 核心发现

**最关键问题**: 图片展示流程完全断裂 - 用户无法看到任何生成结果

**根本原因**:
1. 后端 `Artifact.Bytes` 字段被 Gin 序列化为 `bytes` (base64)
2. 前端期望字段名为 `data` (base64)
3. 没有持久化到 Asset Store，`asset_id` 永远为空
4. URL 格式错误，缺少 `project_id`

---

## P0 Critical 缺陷 (8 个)

### 🔴 DEFECT-IMAGE-001: JSON 字段名不匹配 [新发现]

**严重程度**: Critical
**影响**: 100% 图片无法显示

**问题描述**:
- 后端发送: `{"bytes": "base64..."}`
- 前端期望: `{"data": "base64..."}`

**代码位置**:
- `internal/domain/agent/types.go:94` - `Bytes` 字段 JSON tag
- `web/src/lib/sse.ts:40` - 前端期望 `data` 字段

**修复**: 统一字段名为 `data` 或在前端适配

---

### 🔴 DEFECT-IMAGE-002: Asset ID 永不为空

**严重程度**: Critical
**影响**: 无法通过 Asset API 获取图片

**问题描述**:
- Visualizer 创建 `memory://` URI 的 artifacts
- Runner 没有调用 AssetService 存储
- `asset_id` 字段从未被赋值

**代码位置**:
- `internal/application/agents/visualizer/agent.go:420-430`
- `internal/application/orchestrator/runner.go` - 无持久化逻辑

---

### 🔴 DEFECT-IMAGE-003: Asset API URL 格式错误

**严重程度**: Critical
**影响**: 404 错误

**问题描述**:
- API 端点: `/api/v1/assets/:project_id/:asset_id`
- 前端调用: `/api/v1/assets/${assetId}` (缺少 project_id)

**代码位置**:
- `web/src/components/ArtifactPreview.tsx:24`

---

### 🔴 DEFECT-BATCH-001: 批处理结果内存存储

**严重程度**: Critical
**影响**: 服务器重启丢失所有批量生成结果

**代码位置**:
- `internal/application/orchestrator/batch_runner.go`
- `BatchRunner.results map[string]*domainagent.BatchResult`

---

### 🔴 DEFECT-SECURITY-001: 任意代码执行漏洞

**严重程度**: Critical (安全)
**影响**: 远程代码执行 (RCE)

**问题描述**:
```go
cmd := exec.CommandContext(ctx, e.command, "-c", plotExecutorScript)
cmd.Stdin = strings.NewReader(cleaned)  // 用户代码直接执行
```

Python 脚本中: `exec(code, namespace)` - 无任何沙箱

**代码位置**: `internal/application/agents/visualizer/plot_executor.go:37-82`

**修复**: Docker 容器隔离或禁用 Plot 模式

---

### 🔴 DEFECT-PIPELINE-001: Stylist Agent 未集成

**严重程度**: Critical
**影响**: 生成图片缺少 NeurIPS 风格

**问题描述**:
- Stylist agent 存在但在 canonical pipeline 中是可选的
- Python 的 `dev_full` 模式总是包含 stylist

**代码位置**: `internal/application/orchestrator/runner.go:52-62`

---

### 🔴 DEFECT-PIPELINE-002: 缺少 Lite Retrieval Mode

**严重程度**: Critical
**影响**: Token 消耗高 25 倍

**对比**:
- Python lite mode: ~3K tokens
- Go (无 lite): ~80K tokens

---

### 🔴 DEFECT-DEGRADE-001: 无优雅降级机制

**严重程度**: Critical
**影响**: LLM 失败导致整个 session 失败

**对比 repo-cn**:
```python
if result == "Error":
    return text_fallback  # 继续执行
```

---

## P1 High 缺陷 (12 个)

| ID | 描述 | 位置 |
|----|------|------|
| DEFECT-SECURITY-002 | 开发环境加密密钥持久化风险 | `crypto/aesgcm/service.go:117-143` |
| DEFECT-SECURITY-003 | CORS 配置过于宽松 (`*`) | `middleware/cors.go` |
| DEFECT-IMAGE-004 | BatchArtifact 类型不完整 | `internal/api/dto/batch.go` |
| DEFECT-PIPELINE-003 | Planner 图片限制 2 张 (Python 无限制) | `planner/prompt.go:17` |
| DEFECT-PIPELINE-004 | Context 取消时 Python 进程未清理 | `plot_executor.go` |
| DEFECT-RESUME-001 | Session 恢复验证不完整 | `runner.go` |
| DEFECT-STYLIST-001 | Stylist 上下文标签缺失 | `stylist/agent.go` |
| DEFECT-HEALTH-001 | 健康检查不检查依赖 | `router.go:56-61` |
| DEFECT-STATE-001 | Session 状态膨胀 | `session_repository.go` |
| DEFECT-VALIDATE-001 | 输入验证缺失 | `generate.go` |
| DEFECT-LOG-001 | 敏感数据可能泄露到日志 | 多处 |
| DEFECT-RETRY-001 | 重试逻辑不一致 | `visualizer/agent.go` |

---

## P2 Medium 缺陷 (6 个)

| ID | 描述 |
|----|------|
| DEFECT-COUPLING-001 | main.go 375 行紧耦合 |
| DEFECT-SCHEMA-001 | 缺少 JSON Schema 验证 |
| DEFECT-COST-001 | UI 不显示 token 成本 |
| DEFECT-OBS-001 | Prometheus 端点未连接 |
| DEFECT-RATE-001 | 限流仅在内存中 |
| DEFECT-HEADER-001 | 缺少安全头 |

---

## P3 Low 缺陷 (3 个)

| ID | 描述 |
|----|------|
| DEFECT-VERSION-001 | Prompt 版本管理不一致 |
| DEFECT-MODE-001 | Pipeline mode 验证不完整 |
| DEFECT-STYLE-001 | Style guide 灵活性不足 |

---

## 修复路线图

### 阶段 1: 紧急止血 (3-5 天)

| 任务 | 缺陷 | 工作量 |
|------|------|--------|
| 修复 JSON 字段名 | IMAGE-001 | 0.5 天 |
| 添加 Asset 持久化 | IMAGE-002 | 2 天 |
| 修复 Asset URL | IMAGE-003 | 0.5 天 |
| 禁用 Plot 模式 | SECURITY-001 | 0.5 天 |

### 阶段 2: 核心修复 (5-7 天)

| 任务 | 缺陷 | 工作量 |
|------|------|--------|
| 批处理持久化 | BATCH-001 | 1 天 |
| 实现 Lite Retrieval | PIPELINE-002 | 1 天 |
| 集成 Stylist | PIPELINE-001 | 1 天 |
| 实现优雅降级 | DEGRADE-001 | 2-3 天 |

### 阶段 3: 加固完善 (5-6 天)

| 任务 | 缺陷 | 工作量 |
|------|------|--------|
| 安全加固 | SECURITY-002,003 | 1 天 |
| 健康检查 | HEALTH-001 | 0.5 天 |
| 状态压缩 | STATE-001 | 1 天 |
| 可观测性连接 | OBS-001 | 2 天 |

---

## 功能对比矩阵

| 功能 | repo-cn | paperbanana-clean | 状态 |
|------|---------|-------------------|------|
| **图片展示** | ✅ Base64 内联 | ❌ **字段名不匹配** | **需立即修复** |
| **优雅降级** | ✅ | ❌ | **需移植** |
| **Lite Retrieval** | ✅ 96% 节省 | ❌ | **需移植** |
| **Stylist 集成** | ✅ 总是包含 | ⚠️ 可选 | **需修复** |
| **Plot 执行** | ⚠️ 无沙箱 | ❌ 无沙箱 | **都有风险** |
| **状态恢复** | ❌ | ✅ | Go 优势 |
| **多 Provider** | 2 个 | 16+ | Go 优势 |
| **类型安全** | ❌ | ✅ | Go 优势 |
| **批处理持久化** | ❌ 内存 | ❌ 内存 | **都需修复** |

---

## 结论与建议

### 核心结论

paperbanana-clean 的架构设计优于 repo-cn，但存在**严重的数据流断裂**问题，导致用户完全无法看到生成结果。

### 建议

1. **不建议替换 Go** - 架构优势明显，修复工作量可控
2. **立即修复** IMAGE-001, IMAGE-002, IMAGE-003 - 这是阻塞性缺陷
3. **短期移植** repo-cn 的降级逻辑和 Lite Retrieval
4. **中期加固** 安全性和可观测性

### 风险评估

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| RCE 漏洞 | **极高** | 立即禁用 Plot 模式 |
| 数据丢失 | 高 | 批处理持久化 |
| 用户体验差 | 高 | 图片展示修复 |

---

## 生成的文档

| 文件 | 描述 |
|------|------|
| `FINAL-REPORT.md` | 本报告 |
| `consolidated-defect-report.md` | 综合缺陷清单 |
| `remediation-roadmap.md` | 修复路线图 |
| `image-flow-diagram.md` | 图片展示流程分析 |
| `RESEARCH-003-security-audit.md` | 安全审计报告 |
| `RESEARCH-defect-catalog.md` | 深度缺陷清单 |
| `RESEARCH-004-batch-architecture-comparison.md` | 批处理架构对比 |

---

*报告生成: Team Lifecycle v4 - 5 Agent 并行分析*
*分析时间: 约 9 分钟*
