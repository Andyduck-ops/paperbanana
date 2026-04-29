# PaperBanana 性能优化规范

## 1. 现状分析

### 1.1 前端现状

| 指标 | 当前值 | 目标值 |
|------|--------|--------|
| 主 JS Bundle | ~424 KB (gz: 124 KB) | < 300 KB |
| CSS Bundle | ~78 KB (gz: 13 KB) | < 50 KB |
| i18n Chunk | ~59 KB (gz: 20 KB) | 已分离，可接受 |
| 总 JS | ~517 KB | < 400 KB |
| 首次加载请求数 | ~6 个 | < 5 个 |
| LCP | 未测量 | < 2.5s |
| INP | 未测量 | < 200ms |

**已实现的优化：**
- Manual chunks: vendor (react) + i18n 已分离
- Source maps 生产环境已关闭
- Route-level lazy loading: ProjectsPage, ProviderEditPage

**缺失的优化：**
- CSS code splitting 未启用（所有 CSS 在一个文件）
- Terser minification 未启用（无 drop_console）
- Component-level lazy loading 缺失（SettingsDrawer/HistoryPanel/ExportModal  eagerly loaded）
- Dynamic theme loading 缺失（3 个主题 CSS 全部静态导入）
- Asset inline threshold 未配置
- Performance monitoring 未初始化

### 1.2 后端现状

| 指标 | 当前状态 | 目标 |
|------|----------|------|
| HTTP 框架 | Gin | 已足够高效 |
| 数据库 | SQLite (pure-Go) | 启用 WAL + 连接池优化 |
| 缓存 | Redis (可选) | 默认启用内存缓存 |
| 并发 | errgroup (batch) | 优化并发数 |
| API 压缩 | 未启用 | 启用 Gzip |
| 静态资源 | Go embed | 启用压缩 + 缓存头 |

**已实现的优化：**
- Circuit breaker + retry transport
- Pipeline stage timeouts
- Context-based cancellation
- Prometheus metrics

**缺失的优化：**
- HTTP response compression (Gzip)
- Static asset caching headers
- SQLite connection pool tuning
- Request/response JSON pooling
- Image optimization pipeline

## 2. 前端优化方案

### 2.1 Bundle 优化

#### Phase 1: Vite 构建配置升级
```typescript
// vite.config.ts 关键变更
build: {
  target: 'esnext',
  cssCodeSplit: true,                    // 启用 CSS 分割
  assetsInlineLimit: 4096,               // 4KB 以下资源内联
  chunkSizeWarningLimit: 500,
  sourcemap: false,
  minify: 'terser',
  terserOptions: {
    compress: {
      drop_console: true,
      drop_debugger: true,
      dead_code: true,
      merge_vars: true,
    },
  },
  rollupOptions: {
    output: {
      manualChunks: {
        'vendor': ['react', 'react-dom'],
        'i18n': ['i18next', 'react-i18next'],
        'query': ['@tanstack/react-query'],  // 新增分离
        'icons': ['lucide-react'],           // 如有图标库
      },
    },
  },
}
```

#### Phase 2: 组件懒加载
```typescript
// App.tsx 中当前 eagerly loaded 的组件改为 lazy
const SettingsDrawer = lazy(() => import('./components/settings/SettingsDrawer'));
const HistoryPanel = lazy(() => import('./components/history/HistoryPanel'));
const ExportModal = lazy(() => import('./components/ExportModal'));
const WelcomeWizard = lazy(() => import('./components/WelcomeWizard'));
```

#### Phase 3: 动态主题加载
```typescript
// 当前: import './themes/claude-light.css';
// 改为: 根据 colorScheme 动态 import()
const loadTheme = (scheme: 'light' | 'dark') => {
  if (scheme === 'dark') return import('./themes/linear-dark.css');
  return import('./themes/claude-light.css');
};
```

### 2.2 渲染优化

#### React.memo + useMemo 策略
- Workspace 组件（高频重渲染）包裹 memo
- Generation stages 列表使用 useMemo
- HistoryItem 列表已使用 memo，检查其他列表组件

#### 虚拟列表
- HistoryPanel 会话列表 > 50 条时启用虚拟滚动
- Batch candidate grid > 20 个时启用虚拟滚动

### 2.3 资源优化

| 资源类型 | 优化措施 |
|----------|----------|
| 图片 | WebP/AVIF 格式，响应式 srcset |
| 字体 | font-display: swap，预加载关键字体 |
| SVG | 内联小 SVG，大 SVG 懒加载 |
| CSS | PurgeCSS / Tailwind 已 tree-shake |

### 2.4 性能监控接入
```typescript
// main.tsx 初始化
import { initPerformanceMonitoring } from './lib/performance';

if (import.meta.env.PROD) {
  initPerformanceMonitoring({
    enableWebVitals: true,
    enableResourceMonitoring: true,
    logToConsole: false, // 生产环境发送到分析端点
  });
}
```

## 3. 后端优化方案

### 3.1 HTTP 层优化

#### Gzip 压缩
```go
import "github.com/gin-contrib/gzip"

// router.go
r.Use(gzip.Gzip(gzip.BestSpeed)) // 或 DefaultCompression
```

#### 静态资源缓存
```go
// 前端静态文件服务添加缓存头
r.StaticFS("/", embedFS)
// 对 CSS/JS 添加 Cache-Control: public, max-age=31536000, immutable
```

#### Request/Response Pooling
```go
// 复用 JSON 编解码缓冲区
var jsonBufferPool = sync.Pool{
  New: func() interface{} { return new(bytes.Buffer) },
}
```

### 3.2 数据库优化

#### SQLite 调优
```go
// 连接配置优化
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)        // 根据 CPU 核心调整
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)

// PRAGMA 优化
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA synchronous=NORMAL")
db.Exec("PRAGMA cache_size=-64000")  // 64MB cache
db.Exec("PRAGMA temp_store=MEMORY")
```

#### 查询优化
- 为高频查询添加索引（session.project_id, session.created_at）
- N+1 查询检查（GORM Preload 优化）
- 分页查询限制最大 page size

### 3.3 缓存策略

#### 多级缓存
```
L1: In-memory (sync.Map) - API keys, provider configs
L2: Redis (optional) - LLM response cache
L3: SQLite - Persistent data
```

#### LLM Response Cache
- 当前：Redis 缓存，SHA256 key，24h TTL
- 优化：增加内存缓存层，减少 Redis 往返

### 3.4 并发优化

#### Pipeline Runner
- 当前：Sequential stage execution
- 优化：Independent stages 并行执行（Retriever + Stylist 可并行）

#### Batch Runner
- 当前：maxConcurrent: 10
- 优化：根据 provider rate limit 动态调整

## 4. EXE 封装路径（Tauri）

### 4.1 架构迁移

```
当前: Go HTTP Server (8080) + React SPA (Vite dev server / static)
目标: Tauri (Rust) + Go Sidecar + React Frontend

通信变更:
- REST API → Tauri IPC Commands
- SSE Stream → Tauri Events
- Static files → Tauri Asset Protocol
```

### 4.2 Tauri 配置要点

```json
// tauri.conf.json
{
  "build": {
    "frontendDist": "../web/dist",
    "devUrl": "http://localhost:5173"
  },
  "bundle": {
    "active": true,
    "targets": ["msi", "nsis"],
    "windows": {
      "webviewInstallMode": {
        "type": "embedBootstrapper"
      }
    }
  },
  "app": {
    "withGlobalTauri": true,
    "security": {
      "csp": "default-src 'self'; connect-src 'self' ipc://localhost;"
    }
  }
}
```

### 4.3 Go Sidecar 集成

```rust
// src-tauri/src/main.rs
use tauri::Manager;

fn main() {
  tauri::Builder::default()
    .setup(|app| {
      // 启动 Go sidecar
      let sidecar = app.shell().sidecar("paperbanana-server")?;
      sidecar.spawn()?;
      Ok(())
    })
    .invoke_handler(tauri::generate_handler![api_commands::generate])
    .run(tauri::generate_context!())
    .expect("error while running tauri application");
}
```

### 4.4 性能优势

| 指标 | Electron | Tauri | 提升 |
|------|----------|-------|------|
| 安装包大小 | ~150MB | ~5MB | 30x |
| 内存占用 | ~300MB | ~50MB | 6x |
| 启动时间 | ~3s | ~0.5s | 6x |
| 冷启动 | 慢 | 快 | 显著 |

## 5. 实施优先级

### P0（立即实施）
1. [ ] 升级 Vite 构建配置（terser, cssCodeSplit, assetInlineLimit）
2. [ ] 启用 Gin Gzip 压缩
3. [ ] SQLite WAL + 连接池调优
4. [ ] 初始化性能监控

### P1（短期）
5. [ ] 组件懒加载（SettingsDrawer, HistoryPanel, ExportModal）
6. [ ] 动态主题加载
7. [ ] 静态资源缓存头
8. [ ] 添加数据库索引

### P2（中期）
9. [ ] 虚拟列表（HistoryPanel, Batch candidates）
10. [ ] 内存缓存层（API keys, configs）
11. [ ] Pipeline 并行化
12. [ ] Image optimization

### P3（长期）
13. [ ] Tauri 迁移
14. [ ] Go Sidecar 集成
15. [ ] IPC 通信替换 HTTP
16. [ ] EXE 打包

## 6. 验收标准

### 前端
- [ ] Lighthouse Performance Score >= 90
- [ ] LCP < 2.5s
- [ ] INP < 200ms
- [ ] Bundle size < 400 KB (gzipped total)
- [ ] No render-blocking resources

### 后端
- [ ] API P95 latency < 200ms (非 LLM 调用)
- [ ] SQLite query P95 < 50ms
- [ ] Memory usage < 200MB (idle)
- [ ] Concurrent generation: 10+ users

### EXE 封装
- [ ] 安装包 < 20MB
- [ ] 启动时间 < 2s
- [ ] 内存占用 < 150MB
- [ ] 无外部依赖（单文件安装）
