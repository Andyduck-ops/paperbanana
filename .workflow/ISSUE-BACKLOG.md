# PaperBanana Issue Backlog - 执行计划

**创建时间**: 2026-03-26
**总问题数**: 54
**预估总工时**: ~57 小时 (7-8 天)

---

## 执行批次总览

| 批次 | 名称 | 问题数 | 优先级 | 预估工时 | 依赖 | 状态 |
|------|------|--------|--------|----------|------|------|
| batch-1 | 停止机制与提示词 | 3 | Critical | 9h | - | ⏳ pending |
| batch-2 | 图片展示核心修复 | 3 | Critical | 7h | - | ⏳ pending |
| batch-3 | Hook与历史功能 | 2 | Critical | 5h | - | ⏳ pending |
| batch-4 | 数据与启动 | 2 | Critical | 1.5h | - | ⏳ pending |
| batch-5 | Pipeline流程修复 | 3 | Critical | 13h | batch-1 | ⏳ pending |
| batch-6 | 安全与持久化 | 3 | Critical | 10h | - | ⏳ pending |
| batch-7 | 内存优化 | 1 | High | 4h | batch-2 | ⏳ pending |
| batch-8 | 国际化修复 | 7 | High | 11.5h | - | ⏳ pending |
| batch-9 | 停止机制完善 | 7 | High | 11h | batch-1 | ⏳ pending |
| batch-10 | 安全加固 | 3 | High | 6h | - | ⏳ pending |
| batch-11 | Pipeline细节修复 | 5 | High | 8h | batch-5 | ⏳ pending |
| batch-12 | 配置与健康检查 | 3 | High | 6h | - | ⏳ pending |
| batch-13 | 数据与状态优化 | 3 | High+Medium | 9h | batch-6 | ⏳ pending |
| batch-14 | SSE与事件 | 2 | Medium | 3h | - | ⏳ pending |
| batch-15 | 架构优化 | 6 | Medium+Low | 12h | - | ⏳ pending |

---

## 依赖关系图

```
batch-1 (停止机制) ──┬──> batch-5 (Pipeline) ──> batch-11 (Pipeline细节)
                     └──> batch-9 (停止完善)

batch-2 (图片核心) ──────> batch-7 (内存优化)

batch-6 (安全持久化) ─────> batch-13 (数据状态)

独立批次: batch-3, batch-4, batch-8, batch-10, batch-12, batch-14, batch-15
```

---

## 可并行执行的批次

### 第一波 (无依赖)
```
batch-1, batch-2, batch-3, batch-4, batch-6, batch-8, batch-10, batch-12, batch-14, batch-15
```
**预估总工时**: 67h (可并行，实际耗时取决于并行度)

### 第二波 (依赖 batch-1)
```
batch-5, batch-9
```

### 第三波 (依赖 batch-2 或 batch-6)
```
batch-7, batch-13
```

### 第四波 (依赖 batch-5)
```
batch-11
```

---

## 详细问题清单

### Batch 1: 停止机制与提示词 (Critical)

| ID | 问题 | 文件 | 工时 | 状态 |
|----|------|------|------|------|
| C1 | 无 Cancel API 端点 | `internal/api/router.go` | 4h | ⏳ |
| C2 | useGenerate 缺少 AbortController | `web/src/hooks/useGenerate.ts` | 2h | ⏳ |
| C3 | Style Guide 严重压缩 | `internal/application/agents/stylist/prompts.go` | 3h | ⏳ |

### Batch 2: 图片展示核心修复 (Critical)

| ID | 问题 | 文件 | 工时 | 状态 |
|----|------|------|------|------|
| CRITICAL-IMAGE-001 | JSON 字段名不匹配 | `internal/domain/agent/types.go` | 2h | ⏳ |
| CRITICAL-IMAGE-002 | Asset ID 永远为空 | `internal/application/agents/visualizer/agent.go` | 4h | ⏳ |
| CRITICAL-IMAGE-003 | Asset URL 缺少 project_id | `web/src/components/ArtifactPreview.tsx` | 1h | ⏳ |

### Batch 3: Hook与历史功能 (Critical)

| ID | 问题 | 文件 | 工时 | 状态 |
|----|------|------|------|------|
| CRITICAL-HOOK-001 | useHistory 与 App.tsx 不匹配 | `web/src/hooks/useHistory.ts` | 2h | ⏳ |
| CRITICAL-HISTORY-001 | 历史恢复功能断裂 | `web/src/hooks/useHistory.ts` | 3h | ⏳ |

### Batch 4: 数据与启动 (Critical)

| ID | 问题 | 文件 | 工时 | 状态 |
|----|------|------|------|------|
| CRITICAL-DATA-001 | PaperBananaBench 数据缺失 | `data/PaperBananaBench/` | 0.5h | ⏳ |
| CRITICAL-STARTUP-001 | 无一键启动脚本 | `start.bat` (new) | 1h | ⏳ |

### Batch 5: Pipeline流程修复 (Critical)

| ID | 问题 | 文件 | 工时 | 状态 |
|----|------|------|------|------|
| CRITICAL-PIPELINE-001 | Stylist 未集成 | `internal/application/orchestrator/runner.go` | 3h | ⏳ |
| CRITICAL-PIPELINE-002 | 无 Lite Retrieval Mode | `internal/application/agents/retriever/` | 4h | ⏳ |
| CRITICAL-DEGRADE-001 | 无优雅降级机制 | `internal/application/orchestrator/runner.go` | 6h | ⏳ |

### Batch 6: 安全与持久化 (Critical)

| ID | 问题 | 文件 | 工时 | 状态 |
|----|------|------|------|------|
| CRITICAL-SECURITY-001 | Plot RCE 漏洞 | `internal/application/agents/visualizer/plot_executor.go` | 4h | ⏳ |
| CRITICAL-BATCH-001 | 批处理结果内存存储 | `internal/application/orchestrator/batch_runner.go` | 4h | ⏳ |
| CRITICAL-EXPORT-001 | Export 字段名不一致 | `web/src/components/ExportModal.tsx` | 2h | ⏳ |

---

## 启动命令

### 启动单个批次
```
/team-coordinate 修复 batch-1 问题 (C1, C2, C3)

详细任务:
- C1: 添加 POST /api/v1/sessions/:session_id/cancel 端点
- C2: useGenerate 添加 AbortController 和 cancel 方法
- C3: 扩展 Style Guide，使用完整 NeurIPS 2025 指南

项目路径: D:\git有趣\学习项目合集\绘图\paperbanana-clean
参考实现: D:\git有趣\学习项目合集\绘图\repo-cn
```

### 启动并行批次
```
/team-coordinate 并行修复以下批次:
- batch-1: 停止机制与提示词
- batch-2: 图片展示核心修复
- batch-3: Hook与历史功能
- batch-4: 数据与启动

项目路径: D:\git有趣\学习项目合集\绘图\paperbanana-clean
```

---

## 进度追踪

| 日期 | 完成批次 | 累计工时 | 累计问题 | 备注 |
|------|----------|----------|----------|------|
| - | - | 0h | 0/54 | - |

---

## 状态说明

- ⏳ pending - 待处理
- 🔄 in_progress - 进行中
- ✅ completed - 已完成
- ❌ blocked - 被阻塞
- ⏭️ skipped - 已跳过
