# 快速发现报告 - Docker 与本地开发

**生成时间**: 2026-03-25 23:15
**状态**: 初步分析 (agents 运行中)

---

## 🔴 关键发现 1: Redis 是可选的，Docker 不必需！

### 代码证据

**main.go:58-65**:
```go
if cfg.Cache.Redis.Enabled {  // <-- 条件判断，不是必需！
    redisClient := goredis.NewClient(&goredis.Options{...})
    options.Cache = rediscache.NewCache(rediscache.NewStore(redisClient))
}
```

**Redis 的用途**: 仅用于 **LLM 响应缓存**（避免重复调用 API）

### 结论

| 组件 | 必需性 | 用途 |
|------|--------|------|
| Redis | ❌ 可选 | LLM 响应缓存 (可禁用) |
| Docker | ❌ 不必需 | 仅用于部署，本地开发完全不需要 |
| SQLite | ✅ 必需 | 数据持久化 (本地文件) |

### 建议移除 Docker 依赖

本地开发只需：
```bash
# 后端
go run ./cmd/server --config ./configs/config.yaml

# 前端
cd web && npm run dev
```

---

## 🔴 关键发现 2: PaperBananaBench 数据缺失确认

### 对比

| 项目 | ref.json 大小 | images 目录 |
|------|---------------|-------------|
| paperbanana-clean | 3 bytes (`[]`) | 无 |
| repo-cn | 4.5 MB | 241+ images |

### 代码中如何使用

**planner/agent.go:21**:
```go
const defaultExamplesRoot = "data/PaperBananaBench"
```

**planner/agent.go:199-204**:
```go
func (a *Agent) loadExampleImage(mode domainagent.VisualMode, path string) ([]byte, string, error) {
    resolved, err := a.resolveExamplePath(mode, path)
    if err != nil {
        return nil, "", err  // 错误会传播，可能导致静默失败
    }
    // ...
}
```

### 影响

1. **无参考图像** - Planner 无法加载 few-shot 示例
2. **生成质量下降** - 没有 reference，LLM 输出不稳定
3. **静默失败** - 错误可能被忽略

---

## 🟠 发现 3: 缺少一键启动脚本

### repo-cn 启动体验

```
win-start.bat:
1. 自动检测 Python
2. 自动安装 Python (winget 或下载)
3. 自动创建虚拟环境
4. 自动安装依赖
5. 双击启动完成！
```

### paperbanana-clean 启动体验

```
README.md 说明:
1. 手动安装 Go
2. 手动配置环境变量
3. 手动运行 go run ./cmd/server
4. 手动 cd web && npm install && npm run dev
5. 无启动脚本！
```

### 建议添加

```batch
@echo off
REM start.bat - Windows 一键启动

REM 检查 Go
where go >nul 2>&1 || (
    echo 请先安装 Go: https://go.dev/dl/
    pause
    exit /b 1
)

REM 启动后端
start "PaperBanana Server" cmd /c "go run ./cmd/server --config ./configs/config.yaml"

REM 启动前端
cd web
if not exist node_modules (
    call npm install
)
start "PaperBanana Web" cmd /c "npm run dev"

echo 服务已启动！
echo 后端: http://localhost:8080
echo 前端: http://localhost:5173
pause
```

---

## 🟠 发现 4: 历史恢复功能不完整

### 当前实现

**useHistory.ts**:
```typescript
// 只能列出历史
const response = await apiClient.listHistory(projectId);

// 缺少：恢复特定 session 的功能
// 缺少：重新加载 session 的 artifacts
```

### 缺失功能

1. **点击历史恢复** - HistorySidebar 有 UI 但可能缺少恢复逻辑
2. **缩略图显示** - HistorySession 有 thumbnailUrl 但可能未实现
3. **离线恢复** - 无法从本地存储恢复中断的任务

---

## 🟠 发现 5: ExportModal 需要 base64 imageData

### ExportModal.tsx:35-39

```typescript
if (imageData) {
    const img = new window.Image();
    img.src = `data:image/png;base64,${imageData}`;  // 需要 base64!
    // ...
}
```

### 问题

结合之前的发现：
- 后端发送 `bytes` 字段
- 前端期望 `data` 字段
- ExportModal 需要 `imageData`

**三层不一致！**

---

## 总结

| 问题 | 严重度 | 修复难度 |
|------|--------|----------|
| Docker 不必需但被推广 | Medium | 低 (更新文档) |
| Redis 可选但默认启用 | Low | 低 (配置) |
| PaperBananaBench 数据缺失 | **Critical** | 中 (需下载脚本) |
| 无一键启动脚本 | High | 低 (写脚本) |
| 历史恢复不完整 | Medium | 中 |
| Export 字段名不一致 | **Critical** | 低 |

---

*报告状态: 初步分析*
*Agents 仍在运行中，完整报告稍后更新*
