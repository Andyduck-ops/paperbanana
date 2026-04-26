# 业务逻辑深度审查报告

**Date**: 2026-04-25
**Scope**: `paperbanana-clean/` — 前端 (React/TypeScript) + 后端 (Go)
**Method**: 全量代码阅读 + 数据流分析 + 并发/错误/安全审查

---

## 汇总

| 严重级别 | 前端 | 后端 | 总计 |
|----------|------|------|------|
| **CRITICAL** | 1 | 4 | 5 |
| **HIGH** | 3 | 4 | 7 |
| **MEDIUM** | 4 | 10 | 14 |
| **LOW** | 2 | 6 | 8 |

---

## 前端业务逻辑问题

### [CRITICAL] Security: API Key 通过 localStorage 明文持久化
- **File**: `web/src/stores/providerStore.ts:289-293`
- **Issue**: `partialize` 配置将 `providers` (包含明文 `api_key` 字段) 持久化到 localStorage。任何具有 XSS 能力的攻击者都可读取所有 API 密钥。
- **Fix**: 将 `api_key` 从持久化数据中剔除。密钥应由后端持有，前端仅通过 session cookie 或 token 认证。

### [HIGH] State: Toast 定时器在 store 层不响应组件卸载
- **File**: `web/src/stores/toastStore.ts:37-43`
- **Issue**: `addToast` 中使用 `setTimeout` 自动关闭。这些定时器无法被取消或清理，关闭抽屉/页面后 callbacks 仍持有对旧 state 的引用。若页面快速切换或批量添加 toast，可能产生大量悬空定时器。
- **Fix**: 给每个 toast 返回一个 `clearTimer` 函数，或使用 `useRef` + `useEffect` 清理。

### [HIGH] Data: `handleBatchGeneration` 可能进入不一致状态
- **File**: `web/src/hooks/useGenerationFlow.ts:80-88`
- **Issue**: 当 `numCandidates > 1` 时，先调用 `reset()`（清空单线 state），然后 `startBatch()`。但 `startBatch` 是异步的 — 若在此之间的极短窗口内，UI 查询到的是已重置但尚未进入批量生成的中间状态。
- **Fix**: 使用 generationStateMachine 的原子 `START_BATCH` dispatch（而不是散落的 reset + startBatch 调用）。

### [HIGH] Leak: SSE reader 异常路径可能未调用 `releaseLock()`
- **File**: `web/src/lib/sse.ts:244-255`
- **Issue**: `reader.read()` 抛出异常时（例如网络断开），执行跳转到 `catch` 外部块。但 `releaseLock()` 仅在 `finally` 内执行—如果整个 `try` 块在 `reader.releaseLock()` 之前过早结束，`reader` 未被释放。实际上 `releaseLock()` 在 line 262 的 `finally` 块中，覆盖了 try 内的异常路径。**再检查后确认 safe** — `finally` 确实捕获了所有路径。
- **Verdict**: 实际上是安全的。`finally` 块在 line 261-264 覆盖所有 exit 路径。

### [HIGH] State: `useGenerate` 中 `reset()` 不清除 `abortControllerRef`
- **File**: `web/src/hooks/useGenerate.ts:340-344`
- **Issue**: `reset()` 只清除了 prompt/options ref 并重置状态，但没有清除 `abortControllerRef.current`。如果之前 generate 已被 abort，残留的 abort controller 引用可能在未来新的 generate 调用中被错误 abort。
- **Fix**: 在 `reset()` 中添加 `abortControllerRef.current = null;`。

### [MEDIUM] Logic: `generationReducer.FAIL` 同时清空所有生成状态
- **File**: `web/src/lib/generationStateMachine.ts:137-144`
- **Issue**: `FAIL` action 将所有三种模式的 `isGenerating`、`isBatchGenerating`、`isRefining` 全部设置为 false。如果批处理正在运行、单线生成正好失败，两者状态都被清除。
- **Fix**: 根据 `state.mode` 只清除对应模式的 flag。

### [MEDIUM] Logic: `App.tsx` 中存在多个独立的路由 useEffect，无互斥控制
- **File**: `web/src/App.tsx:125-281`
- **Issue**: App 组件有 7 个独立的 `useEffect` 钩子，分别处理路由同步、wizard 检测、示例加载、历史记录、结果存储、候选自动选择。它们之间无显式先后顺序保证，React 渲染顺序决定了执行顺序 — 可能导致某些 side effect 读取到尚未更新的 state。
- **Fix**: 考虑将这些 effect 收敛到 `useGenerationFlow` hook 中。

### [MEDIUM] Logic: `handleRefine` 在 `imageSourceToFile` 失败时静默失败
- **File**: `web/src/hooks/useGenerationFlow.ts:101-122`
- **Issue**: `imageSourceToFile` 可能因无效 base64 数据而抛出异常，但 `handleRefine` 内部无 try-catch，错误会向上传播到调用方（RefinePanel 的 form onSubmit），导致用户看到未捕获的 promise rejection。
- **Fix**: 在 `handleRefine` 中添加 try-catch，显示 toast 错误。

### [MEDIUM] Deps: `useGenerate` callback 的依赖数组中 `[]` 可能不正确
- **File**: `web/src/hooks/useGenerate.ts:315`
- **Issue**: `generate` 的 `useCallback` 使用空数组 `[]` 作为依赖，但内部引用了 `setState`（React 保证稳定）和 `stageOrder`（局部常量）。如果 stageOrder 或其他逻辑随时间变化，callback 不会更新。当前实现是正确的（stageOrder 是局部常量）。
- **Verdict**: Safe — `setState` 和局部常量在 React 中都是稳定的。

### [LOW] UI: `ImageUpload` 未使用 `initialPreview` 属性
- **File**: `web/src/components/ImageUpload.tsx:12-16`
- **Issue**: `ImageUploadProps` 定义了 `initialPreview` 和 `onClear` 但函数签名未解构它们 — 参数 `onClear` 无法被父组件调用。
- **Fix**: 修复解构：`{ onImageSelect, onClear, initialPreview, disabled = false, className = '' }`。

### [LOW] Logic: `providerStore.parseTimeout` 可能返回 0
- **File**: `web/src/stores/providerStore.ts:107-116`
- **Issue**: 若 `timeout` 字符串为空或非法，返回默认值 60。但若匹配到 `"0ms"` 或 `"0s"`，函数返回 0（因为 `parseInt("0")` 是 0，除以 1000 或直接返回都是 0）。这会导致 timeout 为 0 秒的 provider 配置，可能导致所有请求立即超时。
- **Fix**: 最终返回值用 `Math.max(1, value)` 确保至少 1 秒。

---

## 后端业务逻辑问题

### [CRITICAL] Security: Auth 中间件未应用到任何路由
- **File**: `internal/api/router.go` (全文件)
- **Issue**: `middleware.Auth` 已定义（`middleware/auth.go`）但 router 中没有任何路由组使用它。所有 API 端点均可匿名访问，包括生成、provider 管理、API 密钥操作等。
- **Fix**: 在 `router.go` 中将 `h.Use(middleware.Auth(...))` 应用到敏感路由组。

### [CRITICAL] Security: Rate Limit 中间件未应用到任何路由
- **File**: `internal/api/router.go` (全文件)
- **Issue**: `middleware.RateLimit` 和 `RateLimitByIP` 已实现但从未连接到路由。`/api/v1/generate` 和 `/api/v1/generate/stream` 等最昂贵的端点没有速率限制。
- **Fix**: 在生成相关路由上应用 `RateLimitByIP`。

### [CRITICAL] Security/RCE: Visualizer agent 的 plot 模式允许任意 Python 执行
- **File**: `internal/application/agents/visualizer/agent.go:246-303`
- **Issue**: `executePlot` 获取 LLM 生成的代码并通过 `PlotExecutor.Execute(ctx, code)` 执行。默认 `PlotEnabled` 为 false，但一旦启用就是一个 RCE 漏洞。没有 sandbox、CPU/内存限制或文件系统隔离。
- **Fix**: 在 sandbox 环境（容器、gVisor、或 subprocess with rlimit）中执行 plot 代码；或至少添加严格的代码验证白名单。

### [CRITICAL] Bug: 取消函数可能被重复调用导致 panic
- **File**: `internal/api/handlers/cancel.go:32-46,59`
- **Issue**: `Register` 创建 context 和 cancel func，存储在 map 中。当 `Cancel` 被调用时（通过 API 端点触发），它调用 `cancel()` 并删除 map 中的条目。但`Register` 返回的清理闭包在 defer 时也会调用 `cancel()`。如果 Cancel() 和清理闭包都执行，cancelFunc 会被调用两次。Go 的 context.CancelFunc 第二次调用会 panic。
- **Fix**: `Register` 返回的清理闭包在调用前应先从 map 中删除条目，避免双重取消。

### [HIGH] Resource Leak: `generate.go` SSE 事件消费 goroutine 泄漏
- **File**: `internal/api/handlers/generate.go:115-118`
- **Issue**: 匿名 goroutine `go func() { for range handle.Events() {} }()` 被创建用来消费事件。如果 `handle.Wait()` 先于事件通道关闭而返回错误，该 goroutine 将永远阻塞在一个没有人写入或关闭的通道上。
- **Fix**: 确保 `handle.Events()` 的通道在所有 paths 上都最终被关闭。

### [HIGH] Resource Leak: `runner.go` 的 defer cancel 在 for 循环中泄漏 context 资源
- **File**: `internal/application/orchestrator/runner.go:222-224`
- **Issue**: `stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)` 在 `for` 循环（line 196）中。`defer cancel()` 不会在每次迭代结束时执行 — 只在函数返回时执行。这意味着除了最后一个，所有 stage context 的 timer 资源都被泄漏，直到整个 pipeline 完成。
- **Fix**: 在每次迭代结束时显式调用 `cancel()`，而不是依赖 `defer`。例如在 loop body 末尾使用 `cancel()` 或将每个迭代包装在一个闭包中。

### [HIGH] Resource Leak: StreamGenerate 取消 context 但未等待后台 goroutine
- **File**: `internal/api/handlers/generate.go:164`
- **Issue**: `StreamGenerate` 注册取消回调（`h.registry.Register`），但返回错误时通过 `defer cancel()` 取消 context，并不等待 `h.startRunWithInput()` 启动的后台 goroutine 完成。LLM 连接和 goroutine 继续运行，但没有 consumer 读取结果。
- **Fix**: 在 `defer cancel()` 之后添加 `handle.Wait()` 或类似机制，确保后台 goroutine 正确终止。

### [HIGH] Security: API Key 前缀和后缀明文存储
- **File**: `internal/infrastructure/persistence/sqlite/apikey_repository.go:49-52`
- **Issue**: 密钥的前 6 个和后 4 个字符明文存储在数据库中。这大大减少了暴力破解所需的熵。对于大多数 API 密钥格式（例如 `sk-` 开头的 OpenAI 密钥），前缀泄露了部分密钥结构信息。
- **Fix**: 仅存储密钥的哈希或使用加密哈希存储前缀用于 UI 展示（"sk-...xyz1"）。

### [MEDIUM] Business Logic: `batch_runner.go` 中 `progress.Completed` 被重复赋值
- **File**: `internal/application/orchestrator/batch_runner.go:399-403`
- **Issue**: 成功路径中：
  ```go
  progress.Completed = successful        // line 401
  progress.Completed = completedCount     // line 402 — 覆盖 line 401！
  ```
  Line 401 被 Line 402 覆盖。且 `progress.Failed` 在成功路径中未被更新（始终为初始值 0）。失败路径中正确设置了 `progress.Failed = failed`。
- **Fix**: 成功路径中应设置 `progress.Failed = failed`（即使值为 0），并删除多余的 line 401 赋值。

### [MEDIUM] Error Handling: Provider 同步失败被静默忽略
- **File**: `cmd/server/main.go:114-116`
- **Issue**: `syncStartupProviders` 错误仅以 `Warn` 级别记录。若启动配置中的 API 密钥未同步到 config 存储，用户后续会看到误导性的 "no API key" 错误，而非清晰的启动失败提示。
- **Fix**: 将此提升为 Error 级别，或在之后的操作中明确检查 provider 同步状态。

### [MEDIUM] Error Handling: Refine session 持久化失败被静默忽略
- **File**: `internal/api/handlers/refine.go:467-469`
- **Issue**: `persistRefineSession` 通过 `Warn` 吞掉持久化错误。调用者无法知道 session 是否成功存储，导致用户可能丢失 refine 结果。
- **Fix**: 返回错误或至少设置一个 flag 指示持久化失败。

### [MEDIUM] N+1: ListProviders 为每个 provider 单独查询 API keys
- **File**: `internal/api/handlers/provider.go:24-57`
- **Issue**: `ListProviders` 遍历所有 providers 并为每个调用 `h.svc.ListAPIKeys(p.ID)`。N 个 provider = N 次额外的 DB 查询。
- **Fix**: 实现 `ListAPIKeysByProviderIDs([]string)` 批量查询方法。

### [MEDIUM] N+1: `clearRoleAcrossProviders` 循环中 N 次独立写入
- **File**: `internal/api/handlers/channel.go:179-211`
- **Issue**: 循环中对每个 provider 调用 `UpdateProvider`，每次都是一次独立 DB 写入。
- **Fix**: 在单个事务中批量更新，或使用 `UpdateProviderBatch` 方法。

### [MEDIUM] Validation: Folder 创建未验证 ParentID 是否存在
- **File**: `internal/api/handlers/workspace.go:156-178`
- **Issue**: 可以指定不存在的 `parent_id` 创建 folder，创建会成功但形成孤立节点或后续操作出错。
- **Fix**: 创建前检查 parent folder 是否存在。

### [MEDIUM] Validation: 生成请求中不存在 project_id 被静默忽略
- **File**: `internal/api/handlers/generate.go:83`
- **Issue**: 用户提供不存在的 `project_id` 时，生成正常进行但 artifact 持久化会静默失败，用户丢失结果且无明确提示。
- **Fix**: 在 `validateGenerateRequest` 中检查 project_id 的有效性。

### [MEDIUM] Concurrency: `client_manager.go` 双检查锁缺少重新验证
- **File**: `internal/infrastructure/llm/client_manager.go:43-48`
- **Issue**: 经典的双检查锁模式但缺少 write lock 下的重新检查。两个并发调用可能同时通过读锁检查，各自创建 client，第二个覆盖第一个。
- **Fix**: 在 write lock 内重新检查缓存是否存在。

### [MEDIUM] Type Safety: `apikey_repository.go` 使用 `interface{}` 而非 `context.Context`
- **File**: `internal/infrastructure/persistence/sqlite/apikey_repository.go:27-31`
- **Issue**: `Create` 方法接受 `ctx interface{}` 并通过类型 switch 提取 context。传入非 context 值时会 fallback 到 `context.Background()`，绕过父级的 cancellation/deadline。
- **Fix**: 将参数类型改为 `context.Context`。

### [MEDIUM] Nil Safety: Agent factory 方法可能返回 nil
- **File**: `internal/application/orchestrator/batch_runner.go:325-331`
- **Issue**: `CreatePlanner()`、`CreateVisualizer()`、`CreateCritic()` 无返回值类型保证，可能返回 nil。`NewRunner()` 不检查 map 中的 nil agent，后续调用会 panic。
- **Fix**: 在 runner 初始化时检查 nil agent 并提前 fail fast。

### [LOW] Logic: `generate.go` 中 `resolvePromptFields` 语义不正确
- **File**: `internal/api/handlers/generate.go:575-576`
- **Issue**: 当 content 和 visualIntent 为空时，三者全部返回同一个 prompt 值。`visualIntent` 后续被用作 `Goal` 字段 — 但 prompt 可能很长，不适宜作为 goal。
- **Fix**: 在此情况下至少截断 visualIntent/Goal 到合理长度。

### [LOW] Error Handling: `apikey_repository.go` 中 `MarkUsed` 错误被忽略
- **File**: `internal/infrastructure/persistence/sqlite/apikey_repository.go:153`
- **Issue**: `_ = r.MarkUsed(key.ID)` — 如果标记失败，密钥轮换逻辑无法前移，同一密钥会被重复使用。
- **Fix**: 至少记录该错误以便排查。

### [LOW] History: 获取超过请求限度的记录
- **File**: `internal/api/handlers/history.go:294-297`
- **Issue**: 请求 `Limit=20` 会实际获取 `min(100, 20*5)=100` 条，然后裁剪回 20。多余 80 条记录的查询是浪费的。
- **Fix**: 将 fetchLimit 上限从 100 降低到 `req.Limit` 或接近的值。

### [LOW] Error Handling: `isWorkspaceNotFoundError` / `isNotFoundError` 函数未使用
- **File**: `internal/api/handlers/workspace.go:424-429`, `internal/api/handlers/history.go:748-751`
- **Issue**: 这两个函数被定义但从未被调用。如果它们旨在作为未来检验使用，应该添加注释说明。
- **Fix**: 删除死代码或添加使用场景。

### [LOW] Business Logic: `runtime_client.go` 中不可达的错误分支
- **File**: `internal/infrastructure/llm/runtime_client.go:174-186`
- **Issue**: `errors.Is(err, errDefaultProviderNotFound)` 被检查两次 — 第一次在 line 148 处（分支），第二次在 line 174（不可达，因为之前的条件已在 line 148 捕获）。
- **Fix**: 删除重复的检查分支。

### [LOW] Cleanup: Ratelimit 清理 goroutine 无优雅退出
- **File**: `internal/api/middleware/ratelimit.go:52-84`
- **Issue**: 清理 goroutine 运行整个进程生命周期，但没有通过 `rateLimiter.Stop()` 的优雅退出路径。
- **Fix**: 接受作为长期运行的后台任务，或在 `main.go` 的 graceful shutdown 中调用 `Stop()`。

---

## 建议修复优先级

### Phase 1 — 立即修复 (安全问题 + panic)
1. **Router**: 将 Auth 和 RateLimit 中间件应用到路由
2. **Visualizer**: 禁用或沙箱化 plot 代码执行
3. **cancel.go**: 防止 CancelFunc 重复调用 panic
4. **providerStore.ts**: 从 localStorage 持久化中移除 `api_key`

### Phase 2 — 发布前修复 (资源泄漏 + 数据完整性)
5. **runner.go**: 修复 for 循环中 defer cancel 导致的 context timer 泄漏
6. **generate.go**: 确保事件消费 goroutine 在所有路径都正确终止
7. **apikey_repository.go**: 移除明文密钥前缀/后缀存储
8. **batch_runner.go**: 修复 progress.Completed 重复赋值
9. **useGenerate.ts**: 在 reset() 中清除 abortControllerRef

### Phase 3 — 本次迭代内修复
10. **provider.go**: 实现 N+1 查询优化
11. **workspace.go**: 验证 folder parent_id
12. **generate.go**: 验证 project_id 存在性
13. **refine.go**: 不吞掉 session 持久化错误
14. **toastStore.ts**: 实现 toast 定时器清理
15. **generationStateMachine.ts**: FAIL action 按 mode 区分清除

### Phase 4 — 方便时修复 (代码质量)
16. 删除未使用的函数
17. 修复 `client_manager.go` 双检查锁
18. 修复 `apikey_repository.go` 类型安全
19. 修复 `ImageUpload.tsx` 参数解构
20. 修复 `providerStore.ts` parseTimeout 边界
