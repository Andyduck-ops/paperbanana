# 深度缺陷报告 - 综合分析

**生成时间**: 2026-03-25 23:30
**分析方法**: 代码审查 + 架构分析
**状态**: 完成

---

## 🔴🔴 CRITICAL: useHistory Hook 与 App.tsx 不匹配

### 问题

**App.tsx:121** 期望：
```typescript
const { count: historyCount, restoreSession: restoreHistorySession } = useHistory();
```

**useHistory.ts:53-56** 实际返回：
```typescript
return {
  ...state,      // sessions, isLoading, error
  refresh: fetchHistory,
};
// 没有 count！没有 restoreSession！
```

### 影响

- **编译错误** - TypeScript 应该报错
- **运行时错误** - `historyCount` 和 `restoreHistorySession` 都是 `undefined`
- **历史恢复功能完全失效**

### 修复

需要在 `useHistory.ts` 中添加：

```typescript
export function useHistory(projectId?: string) {
  // ... existing code ...

  const restoreSession = useCallback(async (sessionId: string) => {
    try {
      const session = await apiClient.getSession(sessionId);
      // 处理并返回恢复数据
      return session;
    } catch (err) {
      return null;
    }
  }, []);

  return {
    ...state,
    count: state.sessions.length,
    refresh: fetchHistory,
    restoreSession,
  };
}
```

---

## 🔴 关键发现总结

### 1. Docker 完全不需要

| 组件 | 状态 | 说明 |
|------|------|------|
| Redis | 可选 | 仅 LLM 响应缓存，禁用不影响核心功能 |
| Docker | 不需要 | 仅部署用，本地开发 `go run` + `npm run dev` 即可 |
| SQLite | 必需 | 已经是本地文件，无需额外服务 |

### 2. PaperBananaBench 数据缺失

```
paperbanana-clean: ref.json = "[]" (空)
repo-cn:           ref.json = 4.5MB (241+ images)
```

**影响**: Planner 无法加载参考图像，生成质量严重下降

### 3. 图片展示三层不一致

| 层级 | 字段名 |
|------|--------|
| 后端 Artifact | `Bytes` → JSON: `bytes` |
| 前端 SSE 接口 | 期望 `data` |
| ExportModal | 需要 `imageData` |

**结果**: 图片无法显示

### 4. 历史恢复功能断裂

- Hook 不返回 `restoreSession`
- App.tsx 无法调用恢复逻辑
- 整个历史功能不可用

---

## 完整缺陷清单

### P0 Critical (新增后: 15)

| ID | 描述 | 新增 |
|----|------|------|
| CRITICAL-HOOK-001 | useHistory 与 App.tsx 不匹配 | ✅ |
| CRITICAL-DATA-001 | PaperBananaBench 数据缺失 | ✅ |
| CRITICAL-IMAGE-001 | JSON 字段名不匹配 | 之前 |
| CRITICAL-IMAGE-002 | Asset ID 永远为空 | 之前 |
| CRITICAL-IMAGE-003 | Asset URL 缺少 project_id | 之前 |
| CRITICAL-BATCH-001 | 批处理结果内存存储 | 之前 |
| CRITICAL-SECURITY-001 | Plot RCE 漏洞 | 之前 |
| CRITICAL-PIPELINE-001 | Stylist 未集成 | 之前 |
| CRITICAL-PIPELINE-002 | 无 Lite Retrieval | 之前 |
| CRITICAL-DEGRADE-001 | 无优雅降级 | 之前 |
| CRITICAL-STARTUP-001 | 无一键启动脚本 | ✅ |
| CRITICAL-DATA-002 | 数据获取脚本缺失 | ✅ |
| CRITICAL-EXPORT-001 | Export 字段名不一致 | ✅ |
| CRITICAL-HISTORY-001 | 历史恢复功能断裂 | ✅ |
| CRITICAL-THUMBNAIL-001 | 缩略图未实现 | ✅ |

### P1 High (15)

(省略详细列表，与之前报告类似)

---

## 修复优先级

### 立即修复 (0-2天)

| 任务 | 影响 | 工作量 |
|------|------|--------|
| 修复 useHistory hook | 历史功能恢复 | 2h |
| 统一 JSON 字段名 | 图片可显示 | 2h |
| 创建启动脚本 | 开发体验 | 1h |
| 复制基准数据 | 生成质量 | 30min |

### 短期修复 (3-7天)

| 任务 | 工作量 |
|------|--------|
| Asset 持久化 | 2天 |
| 优雅降级 | 2天 |
| Lite Retrieval | 1天 |

### 中期完善 (7-14天)

| 任务 | 工作量 |
|------|--------|
| Plot 安全化 | 2天 |
| 批处理持久化 | 1天 |
| 可观测性 | 2天 |

---

## 总工作量估算

| 阶段 | 估算 |
|------|------|
| 立即止血 | 1 天 |
| 核心修复 | 5 天 |
| 完善加固 | 5 天 |
| **总计** | **11-15 天** |

---

## 关于 Docker 的建议

**结论: 完全移除 Docker 依赖**

理由：
1. Redis 是可选的
2. SQLite 已经是本地文件
3. 本地开发只需 `go run` + `npm run dev`
4. Docker 增加了不必要的复杂性

**替代方案**:
- 创建 `start.bat` / `start.sh` 一键启动脚本
- 在 README 中说明本地开发流程
- Docker 仅用于生产部署（可选）

---

## 下一步行动

1. **立即**: 修复 useHistory hook
2. **今天**: 统一 JSON 字段名
3. **本周**: 复制基准数据，创建启动脚本
4. **下周**: 实现核心修复

---

*报告生成: 深度团队分析*
