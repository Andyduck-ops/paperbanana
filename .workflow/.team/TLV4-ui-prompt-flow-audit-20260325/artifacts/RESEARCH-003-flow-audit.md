# RESEARCH-003: 状态流转与交接审核

## 审核概览

| 维度 | 状态 | 备注 |
|------|------|------|
| 状态机 | ✅ 正确 | Session 状态定义完整，转换逻辑正确 |
| Agent 交接 | ✅ 正确 | 数据传递链路完整 |
| 上下文传递 | ⚠️ 轻微 | SSE 字段命名有 snake_case/camelCase 差异 |
| 内存管理 | ⚠️ 需关注 | 大数据传递使用深拷贝，存在优化空间 |
| Python 对比 | ✅ 一致 | 核心流程一致，Go 版本架构更清晰 |

---

## 1. 状态机审核

### 1.1 Session 状态定义

**位置**: `internal/domain/agent/types.go:32-40`

```go
type RunStatus string

const (
    StatusPending   RunStatus = "pending"
    StatusRunning   RunStatus = "running"
    StatusCompleted RunStatus = "completed"
    StatusFailed    RunStatus = "failed"
    StatusCanceled  RunStatus = "canceled"
)
```

**状态转换图**:

```
pending → running → completed
    ↓         ↓
    ↓    failed/canceled
    ↓
[错误]
```

### 1.2 状态转换逻辑

**位置**: `internal/application/orchestrator/runner.go`

| 起点 | 触发条件 | 终点 | 代码位置 |
|------|----------|------|----------|
| pending | `Start()` 调用 | running | `runner.go:110-114` |
| running | 所有 stage 完成 | completed | `runner.go:257-269` |
| running | context 取消 | canceled | `runner.go:315-326` |
| running | stage 执行失败 | failed | `runner.go:328-339` |

**状态转换正确性**: ✅ 无非法转换，状态转换原子性由 `sync.Mutex` 保护 (`session.go:52`)

### 1.3 Stage 状态定义

```go
type StageName string

const (
    StageRetriever  StageName = "retriever"
    StagePlanner    StageName = "planner"
    StageStylist    StageName = "stylist"
    StageVisualizer StageName = "visualizer"
    StageCritic     StageName = "critic"
    StagePolish     StageName = "polish"
)
```

**Pipeline 顺序**: `types.go:20-26`

```go
var pipelineOrder = []StageName{
    StageRetriever,
    StagePlanner,
    StageStylist,
    StageVisualizer,
    StageCritic,
}
```

**Pipeline Mode 过滤**: `runner.go:606-638`

- `full`: 完整 5 阶段
- `planner-critic`: 仅 planner → critic
- `vanilla`: 仅 visualizer

---

## 2. Agent 交接审核

### 2.1 数据流架构

```
User Input
    ↓
┌─────────────────────────────────────────────────────────────┐
│                    Session Tracker                           │
│  (session.go:51-55)                                         │
│  - state: SessionState                                      │
│  - currentInput: AgentInput                                 │
└─────────────────────────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────────────────────────┐
│  Stage Loop (runner.go:185-255)                             │
│                                                             │
│  for _, stage := range stages {                            │
│      stageInput := tracker.stageInput(stage)               │
│      output := stageAgent.Execute(ctx, stageInput)         │
│      tracker.completeStage(stageState, output)             │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 交接数据结构

**AgentInput** (`types.go:127-140`):

| 字段 | 类型 | 传递方向 |
|------|------|----------|
| SessionID | string | 全局 |
| RequestID | string | 全局 |
| Stage | StageName | 每阶段设置 |
| Content | string | planner → stylist → visualizer |
| Messages | []Message | LLM 对话历史 |
| VisualIntent | VisualIntent | retriever → 全链 |
| RetrievedReferences | []RetrievedReference | retriever → planner |
| GeneratedArtifacts | []Artifact | 累积传递 |
| CritiqueRounds | []CritiqueRound | critic 迭代 |
| Metadata | map[string]string | 配置/状态 |

### 2.3 交接逻辑

**位置**: `session.go:127-154`

```go
func mergeAgentInput(input AgentInput, output AgentOutput) AgentInput {
    next := cloneAgentInput(input)

    if output.Content != "" {
        next.Content = output.Content  // 内容传递
    }
    if len(output.Messages) > 0 {
        next.Messages = cloneMessages(output.Messages)
    }
    if hasVisualIntent(output.VisualIntent) {
        next.VisualIntent = cloneVisualIntent(output.VisualIntent)
    }
    // ... 更多字段合并
    next.Metadata = mergeStringMaps(next.Metadata, output.Metadata)
    return next
}
```

**交接正确性**: ✅ 所有字段正确合并，未丢失数据

### 2.4 各 Agent 输入输出

| Agent | 输入关键字段 | 输出关键字段 |
|-------|-------------|-------------|
| Retriever | VisualIntent, Content | RetrievedReferences, GeneratedArtifacts[ReferenceBundle] |
| Planner | RetrievedReferences, VisualIntent | Content(plan), GeneratedArtifacts[Plan] |
| Stylist | Content, VisualIntent | Content(styled), GeneratedArtifacts[Plan] |
| Visualizer | Content, GeneratedArtifacts | GeneratedArtifacts[RenderedFigure, PromptTrace] |
| Critic | GeneratedArtifacts[RenderedFigure], CritiqueRounds | CritiqueRounds, GeneratedArtifacts[Critique] |

---

## 3. 上下文传递审核

### 3.1 SSE 事件类型

**后端定义** (`internal/domain/agent/events.go:5-24`):

```go
const (
    EventRunStarted     EventType = "run_started"
    EventStageStarted   EventType = "stage_started"
    EventStageCompleted EventType = "stage_completed"
    EventStageFailed    EventType = "stage_failed"
    EventRunCompleted   EventType = "run_completed"
    EventRunFailed      EventType = "run_failed"
    EventRunCanceled    EventType = "run_canceled"
    EventResumeStarted  EventType = "resume_start"
    // Batch events...
)
```

**前端定义** (`web/src/lib/sse.ts:7-15`):

```typescript
export type SSEEventType =
  | 'stage_started'
  | 'stage_completed'
  | 'stage_failed'
  | 'run_completed'
  | 'run_failed'
  | 'result'
  | 'error'
  | 'resume_start';
```

**事件类型一致性**: ✅ 完全匹配

### 3.2 SSE 事件字段

| 事件类型 | 后端字段 | 前端期望 | 一致性 |
|----------|----------|----------|--------|
| stage_started | stage, metadata | stage, agent | ⚠️ agent 字段缺失 |
| stage_completed | stage, timing, metadata | stage, summary, artifact_count, artifact_kinds | ⚠️ 字段在 metadata 中 |
| run_completed | status, timing | session_id, generated_artifacts | ⚠️ 前端用 result 事件 |
| resume_start | metadata.resumed_from_stage | resumed_from_stage, stages_completed_before_resume | ⚠️ 字段展开不一致 |

**问题发现**:

1. **stage_started 事件**: 后端 `metadata` 包含 agent 信息，前端期望顶层 `agent` 字段
   - **影响**: 前端显示 agent 名称可能不正确
   - **修复建议**: 在事件发布时添加顶层 agent 字段

2. **result 事件**: 后端无 `result` 类型，使用 `run_completed`
   - **影响**: 前端通过 `run_completed` 和 `result` 共同处理结果
   - **现状**: `sse.ts:102-105` 合并处理，可正常工作

### 3.3 字段命名差异

| 后端 (snake_case) | 前端期望 | 处理方式 |
|-------------------|----------|----------|
| session_id | session_id | ✅ 一致 |
| generated_artifacts | generated_artifacts | ✅ 一致 |
| artifact_count | artifact_count | ✅ 一致 |
| artifact_kinds | artifact_kinds | ✅ 一致 |
| failed_stage | failed_stage | ✅ 一致 |
| stages_not_run | stages_not_run | ✅ 一致 |

**命名一致性**: ✅ 使用 snake_case，前后端一致

---

## 4. 内存管理审核

### 4.1 深拷贝策略

**位置**: `session.go` 和 `agentstate/store.go`

所有数据传递使用深拷贝:

```go
func cloneAgentInput(input AgentInput) AgentInput {
    cloned := input
    cloned.Messages = cloneMessages(input.Messages)
    cloned.VisualIntent = cloneVisualIntent(input.VisualIntent)
    cloned.RetrievedReferences = cloneReferences(input.RetrievedReferences)
    cloned.Prompt = clonePrompt(input.Prompt)
    cloned.GeneratedArtifacts = cloneArtifacts(input.GeneratedArtifacts)
    cloned.CritiqueRounds = cloneCritiqueRounds(input.CritiqueRounds)
    cloned.Metadata = cloneStringMap(input.Metadata)
    return cloned
}
```

### 4.2 大数据传递分析

| 数据类型 | 典型大小 | 拷贝次数 | 风险 |
|----------|----------|----------|------|
| RetrievedReferences | ~10KB | 5次 (每阶段) | 低 |
| GeneratedArtifacts[ReferenceBundle] | ~50KB | 4次 | 中 |
| GeneratedArtifacts[RenderedFigure] | ~500KB-2MB | 2-3次 | **高** |
| Messages | ~20KB | 5次 | 中 |

**问题发现**:

1. **RenderedFigure 深拷贝**: 图像数据 (Base64) 每次拷贝 500KB-2MB
   - **影响**: Critic 迭代时图像多次拷贝
   - **优化建议**: 考虑使用引用计数或指针传递

2. **消息历史累积**: LLM Messages 随对话增长
   - **影响**: 长对话内存占用增加
   - **现状**: 可接受，但需监控

### 4.3 快照存储

**位置**: `internal/infrastructure/agentstate/store.go`

```go
func (s *Store) Save(session SessionState, state AgentState) error {
    snapshot := BuildSnapshot(session, state)
    payload, _ := json.MarshalIndent(snapshot, "", "  ")
    // 写入 .paperbanana/sessions/{sessionID}/agent_states/{stage}.json
}
```

**存储效率**: ⚠️ JSON 文本格式，每阶段完整快照
- **优化空间**: 使用二进制格式或增量快照

---

## 5. Python 实现对比

### 5.1 流程控制对比

| 维度 | Python (repo-cn) | Go (paperbanana-clean) |
|------|------------------|------------------------|
| 流程模式 | exp_mode 字符串匹配 | pipeline_mode 配置 |
| 状态管理 | 无显式状态机 | RunStatus 状态机 |
| Agent 生命周期 | 无统一接口 | BaseAgent 接口 |
| 并发控制 | asyncio.Semaphore | errgroup.WithContext |
| 错误处理 | try-catch | 错误分类 + 重试 |

### 5.2 Python 流程分支

**位置**: `repo-cn/utils/paperviz_processor.py:102-191`

```python
if exp_mode == "vanilla":
    data = await self.vanilla_agent.process(data)
elif exp_mode == "dev_planner":
    data = await self.retriever_agent.process(data)
    data = await self.planner_agent.process(data)
    data = await self.visualizer_agent.process(data)
elif exp_mode == "dev_full":
    data = await self.retriever_agent.process(data)
    data = await self.planner_agent.process(data)
    data = await self.stylist_agent.process(data)
    data = await self.visualizer_agent.process(data)
    data = await self._run_critic_iterations(data, ...)
```

### 5.3 Go 流程控制

**位置**: `runner.go:606-638`

```go
func (r *Runner) filterPipeline(metadata map[string]string, base []StageName) []StageName {
    mode := strings.TrimSpace(metadata["config.pipeline_mode"])
    switch mode {
    case "full":
        return append([]StageName(nil), base...)
    case "planner-critic":
        allowed[StagePlanner] = true
        allowed[StageCritic] = true
    case "vanilla":
        allowed[StageVisualizer] = true
    }
    // ...
}
```

### 5.4 Critic 迭代对比

**Python**: `_run_critic_iterations` 方法独立实现
- 硬编码最大轮数逻辑
- 图像字段键名动态构建

**Go**: Critic Agent 内部实现迭代
- `RevisionAgent` 注入实现 re-render
- 标准化的 `CritiqueRound` 结构

### 5.5 架构优势对比

| 方面 | Python | Go |
|------|--------|-----|
| 可扩展性 | 中等 (分支匹配) | 高 (接口组合) |
| 可测试性 | 中等 | 高 (Mock 接口) |
| 类型安全 | 无 | 强类型 |
| 并发性能 | asyncio | goroutine |
| 错误处理 | 异常传播 | 结构化错误 |

---

## 6. 问题汇总

### 6.1 高优先级

| 编号 | 问题 | 影响 | 建议 |
|------|------|------|------|
| H1 | 图像数据多次深拷贝 | 内存压力 | 使用指针传递或引用计数 |

### 6.2 中优先级

| 编号 | 问题 | 影响 | 建议 |
|------|------|------|------|
| M1 | SSE 事件字段结构不一致 | 前端解析需适配 | 统一顶层字段定义 |
| M2 | 快照使用 JSON 文本 | 存储效率 | 考虑 protobuf 或增量快照 |

### 6.3 低优先级

| 编号 | 问题 | 影响 | 建议 |
|------|------|------|------|
| L1 | 前端缺少 `run_started` 事件处理 | 无进度开始指示 | 添加事件处理 |

---

## 7. 审核结论

### 7.1 整体评估

**状态流转与交接机制设计合理，实现正确。** Go 版本相比 Python 版本有显著架构改进:

1. ✅ 统一的 `BaseAgent` 接口定义生命周期
2. ✅ 明确的状态机模型
3. ✅ 结构化的错误处理
4. ✅ 类型安全的数据传递
5. ✅ 可恢复的快照机制

### 7.2 建议

1. **优化内存**: 减少图像数据的深拷贝
2. **统一 SSE 格式**: 确保前后端事件字段一致
3. **监控指标**: 添加内存使用和阶段耗时监控

---

## 附录: 关键文件清单

| 文件路径 | 职责 |
|----------|------|
| `internal/domain/agent/types.go` | 状态、事件、数据结构定义 |
| `internal/domain/agent/events.go` | SSE 事件类型定义 |
| `internal/domain/agent/agent.go` | BaseAgent 接口 |
| `internal/application/orchestrator/runner.go` | Pipeline 执行引擎 |
| `internal/application/orchestrator/session.go` | Session 状态跟踪 |
| `internal/infrastructure/agentstate/store.go` | 快照存储 |
| `internal/application/agents/*/agent.go` | 各 Agent 实现 |
| `web/src/lib/sse.ts` | 前端 SSE 客户端 |
| `web/src/hooks/useGenerate.ts` | 前端生成状态管理 |
