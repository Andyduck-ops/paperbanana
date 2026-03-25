# 修复优先级与实施路线图

**项目**: paperbanana-clean 漂移修复
**生成时间**: 2026-03-25
**目标**: 修复关键缺陷，对齐 repo-cn 运行逻辑

---

## 总体策略

**不建议替换 Go**，原因：
1. 架构优势明显（类型安全、多 Provider、状态恢复）
2. 修复工作量可控（9-13 天）
3. 已有投资不应浪费

**核心修复思路**:
1. 连接数据流断点（图片展示）
2. 移植 repo-cn 运行逻辑（降级、Lite Retrieval）
3. 完善安全性和可观测性

---

## 阶段 1: 止血 (3-5 天)

### 任务 1.1: 修复图片展示流程

**优先级**: P0
**工作量**: 2-3 天
**目标**: 用户能看到生成的图片

#### 步骤

**Step 1: 扩展 Artifact 结构**
```go
// internal/domain/agent/types.go
type Artifact struct {
    ID       string            `json:"id"`
    Kind     ArtifactKind      `json:"kind"`
    MIMEType string            `json:"mime_type"`
    URI      string            `json:"uri"`
    Content  string            `json:"content,omitempty"`
    Bytes    []byte            `json:"bytes,omitempty"`
    Metadata map[string]string `json:"metadata,omitempty"`
    // 新增
    AssetID  string            `json:"asset_id,omitempty"`
}
```

**Step 2: 在 Runner 中持久化 Artifacts**
```go
// internal/application/orchestrator/runner.go
type Runner struct {
    // 现有字段...
    assetStore AssetStore  // 新增
}

func (r *Runner) persistArtifacts(ctx context.Context, output *AgentOutput, projectID string) error {
    for i := range output.GeneratedArtifacts {
        artifact := &output.GeneratedArtifacts[i]
        if len(artifact.Bytes) > 0 {
            assetID, err := r.assetStore.Store(ctx, projectID, artifact.ID, artifact.Bytes, artifact.MIMEType)
            if err != nil {
                return err
            }
            artifact.AssetID = assetID
            artifact.URI = fmt.Sprintf("asset://%s/%s", projectID, assetID)
        }
    }
    return nil
}
```

**Step 3: 更新前端接口**
```typescript
// web/src/lib/sse.ts
export interface ResultEvent {
  session_id: string;
  project_id: string;  // 新增
  generated_artifacts: Array<{
    id: string;
    kind: string;
    mime_type: string;
    summary: string;
    data?: string;
    asset_id?: string;  // 新增
  }>;
}
```

**Step 4: 修复前端 URL 构建**
```typescript
// web/src/hooks/useGenerate.ts
artifacts: data.generated_artifacts.map((a) => ({
  kind: a.kind,
  mimeType: a.mime_type,
  summary: a.summary,
  data: a.data,
  assetId: a.asset_id,  // 直接使用，不再类型断言
})),

// web/src/components/ArtifactPreview.tsx
const imageUrl = artifact.data
  ? `data:${artifact.mimeType};base64,${artifact.data}`
  : artifact.assetId && projectId
  ? `/api/v1/assets/${projectId}/${artifact.assetId}/download`  // 修复 URL
  : null;
```

#### 验收标准
- [ ] 生成完成后，图片在 UI 中正确显示
- [ ] 刷新页面后，历史记录中的图片仍可显示
- [ ] 批量生成的图片可下载

---

### 任务 1.2: 批处理结果持久化

**优先级**: P0
**工作量**: 1 天

#### 步骤

**Step 1: 创建 BatchResult 表**
```sql
CREATE TABLE batch_results (
    batch_id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    candidate_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    artifacts JSON,  -- 存储 []Artifact
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id)
);
```

**Step 2: 在 BatchRunner 中持久化**
```go
// 完成 each candidate 后保存
func (r *BatchRunner) saveCandidateResult(ctx context.Context, result CandidateResult) error {
    return r.db.SaveBatchResult(ctx, result)
}
```

#### 验收标准
- [ ] 服务器重启后，批处理结果可恢复
- [ ] 批量下载功能正常工作

---

## 阶段 2: 稳定 (3-4 天)

### 任务 2.1: 实现优雅降级

**优先级**: P0
**工作量**: 2-3 天

#### 设计

```go
// internal/application/orchestrator/runner.go

type FallbackStrategy int

const (
    FallbackContinue FallbackStrategy = iota  // 继续执行
    FallbackRetry                              // 重试
    FallbackAbort                              // 终止
)

func (r *Runner) executeStageWithFallback(ctx context.Context, stage StageName, input AgentInput) (AgentOutput, error) {
    output, err := r.executeStage(ctx, stage, input)
    if err == nil {
        return output, nil
    }

    // 分类错误
    category := ClassifyError(err)
    strategy := r.fallbackStrategy(stage, category)

    switch strategy {
    case FallbackRetry:
        // 指数退避重试
        return r.retryStage(ctx, stage, input, 3)
    case FallbackContinue:
        // 降级：使用上一个有效输出
        return r.degradedOutput(input), nil
    default:
        return AgentOutput{}, err
    }
}

func (r *Runner) degradedOutput(input AgentInput) AgentOutput {
    // 如果 visualizer 失败，返回文本描述
    return AgentOutput{
        Stage:   input.Stage,
        Content: input.Content,
        GeneratedArtifacts: []Artifact{{
            Kind:     ArtifactKindRenderedFigure,
            Content:  "[Image generation failed - text only mode]",
            MIMEType: "text/plain",
        }},
    }
}
```

#### 验收标准
- [ ] LLM 超时后，pipeline 继续执行
- [ ] Visualizer 失败时，返回文本描述
- [ ] 错误信息清晰记录在 session 中

---

### 任务 2.2: Plot 执行安全化

**优先级**: P0
**工作量**: 2 天

#### 方案选择

| 方案 | 安全性 | 复杂度 | 推荐度 |
|------|--------|--------|--------|
| Docker 容器 | ⭐⭐⭐⭐⭐ | 高 | ⭐⭐⭐ |
| gVisor 沙箱 | ⭐⭐⭐⭐⭐ | 高 | ⭐⭐⭐ |
| RestrictedPython | ⭐⭐⭐ | 中 | ⭐⭐ |
| 禁用 Plot 模式 | ⭐⭐⭐⭐⭐ | 低 | ⭐⭐⭐⭐ |

**推荐方案**: 先禁用 Plot 模式，后续迭代中添加 Docker 支持

#### 步骤

**Step 1: 添加配置开关**
```go
// internal/config/config.go
type Config struct {
    EnablePlotMode bool `yaml:"enable_plot_mode" default:"false"`
}
```

**Step 2: 返回明确错误**
```go
// internal/application/agents/visualizer/agent.go
func (a *Agent) executePlot(ctx context.Context, input AgentInput, prompt PromptMetadata) (AgentOutput, error) {
    if !a.cfg.EnablePlotMode {
        return AgentOutput{}, errors.New("plot mode is disabled for security reasons; use diagram mode instead")
    }
    // 现有逻辑...
}
```

#### 验收标准
- [ ] Plot 模式默认禁用
- [ ] 尝试使用 Plot 模式时返回清晰错误
- [ ] 文档说明安全风险

---

### 任务 2.3: 移植 Lite Retrieval Mode

**优先级**: P1
**工作量**: 1 天

#### 设计

```go
// internal/application/agents/retriever/agent.go

type RetrievalMode string

const (
    RetrievalModeAuto     RetrievalMode = "auto"      // 智能选择
    RetrievalModeAutoFull RetrievalMode = "auto-full" // 完整检索
    RetrievalModeLite     RetrievalMode = "lite"      // 精简检索
)

func (a *Agent) executeLiteRetrieval(ctx context.Context, input AgentInput) (AgentOutput, error) {
    // 仅检索最相关的 3 个样本
    // 精简 prompt，减少 token 消耗
    // 类似 repo-cn 的实现
}
```

#### 验收标准
- [ ] Lite 模式减少 90%+ token 消耗
- [ ] 生成质量无明显下降

---

## 阶段 3: 增强 (3-4 天)

### 任务 3.1: 统一 SSE 接口

**工作量**: 1 天

### 任务 3.2: 实现真实健康检查

**工作量**: 0.5 天

### 任务 3.3: Session 状态压缩

**工作量**: 1 天

### 任务 3.4: 连接可观测性端点

**工作量**: 2 天

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Artifact 持久化失败 | 数据丢失 | 事务保护 + 重试 |
| 降级逻辑引入 bug | 质量下降 | 充分测试 + 金牌数据验证 |
| Docker 环境差异 | 部署问题 | 提供配置选项 |

---

## 总结

| 阶段 | 任务数 | 工作量 | 产出 |
|------|--------|--------|------|
| 阶段 1 止血 | 2 | 3-5 天 | 图片可展示，批处理可恢复 |
| 阶段 2 稳定 | 3 | 3-4 天 | 降级机制，安全加固，Lite 模式 |
| 阶段 3 增强 | 4 | 3-4 天 | 接口统一，可观测性 |
| **总计** | **9** | **9-13 天** | **完全修复** |

---

*路线图生成: Team Lifecycle v4*
