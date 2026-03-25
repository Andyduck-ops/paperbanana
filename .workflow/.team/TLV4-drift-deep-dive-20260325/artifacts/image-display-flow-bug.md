# 图片展示流程关键缺陷报告

**发现时间**: 2026-03-25
**严重程度**: P0 (Critical)
**影响范围**: 所有生成结果展示

---

## 问题描述

生成的图片无法正确展示，因为前端与后端之间的数据流存在断层。

---

## 根本原因分析

### 1. 后端 `Artifact` 结构体缺少 `asset_id`

**位置**: `internal/domain/agent/types.go:87-95`

```go
type Artifact struct {
	ID       string            `json:"id"`       // 有 ID
	Kind     ArtifactKind      `json:"kind"`
	MIMEType string            `json:"mime_type"`
	URI      string            `json:"uri"`      // 有 URI
	Content  string            `json:"content,omitempty"`
	Bytes    []byte            `json:"bytes,omitempty"`  // 有 Bytes (base64)
	Metadata map[string]string `json:"metadata,omitempty"`
}
```

**缺失**: 没有 `asset_id` 字段！

### 2. SSE 事件 `ResultEvent` 接口不匹配

**位置**: `web/src/lib/sse.ts:34-42`

```typescript
export interface ResultEvent {
  session_id: string;
  generated_artifacts: Array<{
    kind: string;
    mime_type: string;
    summary: string;
    data?: string;
  }>;
}
```

**缺失**: 没有 `asset_id` 字段！

### 3. 前端尝试使用不存在的字段

**位置**: `web/src/hooks/useGenerate.ts:170-176`

```typescript
artifacts: data.generated_artifacts.map((a) => ({
  kind: a.kind,
  mimeType: a.mime_type,
  summary: a.summary,
  data: a.data,
  assetId: (a as { asset_id?: string }).asset_id,  // 类型断言，实际为 undefined!
})),
```

### 4. 图片 URL 构建逻辑

**位置**: `web/src/components/ArtifactPreview.tsx:21-25`

```typescript
const imageUrl = artifact.data
  ? `data:${artifact.mimeType};base64,${artifact.data}`  // 需要 base64 data
  : artifact.assetId
  ? `/api/v1/assets/${artifact.assetId}`  // 需要 assetId
  : null;  // 都没有 → null → 无图片
```

---

## 数据流追踪

```
后端生成图片
    ↓
Artifact.Bytes (原始字节)
    ↓
SSE 事件发送 → generated_artifacts
    ↓
问题: Artifact 只有 ID, URI, Bytes，没有 asset_id
    ↓
前端尝试获取 asset_id → undefined
    ↓
如果没有 base64 data，imageUrl = null
    ↓
图片不显示
```

---

## 关键问题

1. **后端没有持久化图片到 Asset Store**
   - 生成后图片只在 `Artifact.Bytes` 中
   - 没有调用 `AssetService.Store()` 保存

2. **SSE 没有发送 base64 data**
   - 大图片通过 SSE 发送 base64 不现实
   - 应该存储后发送 `asset_id`

3. **API 路由需要 `project_id`**
   - `/api/v1/assets/:project_id/:asset_id`
   - 前端 URL 构建缺少 project_id

---

## 修复方案

### 方案 A: 完整持久化流程 (推荐)

1. **后端**: 生成完成后，将图片存储到 Asset Store
2. **后端**: 在 SSE 结果中返回 `asset_id` 和 `project_id`
3. **前端**: 使用正确的 URL 格式 `/api/v1/assets/{project_id}/{asset_id}/download`

### 方案 B: Base64 内联 (临时方案)

1. **后端**: 在 SSE 结果中包含 base64 编码的图片数据
2. **前端**: 直接使用 data URI 显示
3. **限制**: 仅适用于小图片，大图片会导致 SSE 超时

---

## 验证测试

```bash
# 测试 SSE 返回的数据结构
curl -N -X POST http://localhost:8080/api/v1/generate/stream \
  -H "Content-Type: application/json" \
  -d '{"prompt": "test diagram"}' | jq '.generated_artifacts'

# 检查是否包含 asset_id 或 base64 data
```

---

## 相关文件

| 文件 | 作用 |
|------|------|
| `internal/domain/agent/types.go` | Artifact 结构定义 |
| `internal/api/handlers/generate.go` | SSE 处理 |
| `internal/infrastructure/assets/localstore/store.go` | Asset 存储 |
| `web/src/lib/sse.ts` | SSE 客户端 |
| `web/src/hooks/useGenerate.ts` | 结果处理 |
| `web/src/components/ArtifactPreview.tsx` | 图片显示 |
