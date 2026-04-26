# PaperBanana 性能优化实施指南

本指南提供可直接实施的性能优化代码和配置。

---

## 第一部分: Vite 配置优化

### 1.1 更新 vite.config.ts

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { visualizer } from 'rollup-plugin-visualizer';

// https://vite.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [
    react(),
    tailwindcss(),
    // Bundle 分析器 (仅在分析构建时启用)
    mode === 'analyze' && visualizer({
      open: true,
      gzipSize: true,
      brotliSize: true,
      filename: 'dist/stats.html',
    }),
  ],
  server: {
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'esnext',
    cssCodeSplit: true,
    sourcemap: mode !== 'production',
    // 资源内联阈值: 4KB 以下的资源内联
    assetsInlineLimit: 4096,
    // Chunk 大小警告阈值
    chunkSizeWarningLimit: 500,
    rollupOptions: {
      output: {
        // 手动分块策略
        manualChunks: {
          // 框架核心依赖 - 长期缓存
          'vendor': ['react', 'react-dom'],
          // i18n 相关 - 独立缓存
          'i18n': ['i18next', 'react-i18next'],
        },
        // 入口文件命名
        entryFileNames: 'js/[name]-[hash].js',
        // Chunk 文件命名
        chunkFileNames: 'js/[name]-[hash].js',
        // 资源文件命名
        assetFileNames: (assetInfo) => {
          const info = assetInfo.name.split('.');
          const ext = info[info.length - 1];
          if (/\.css$/i.test(assetInfo.name)) {
            return 'css/[name]-[hash][extname]';
          }
          return 'assets/[name]-[hash][extname]';
        },
      },
    },
    // 压缩配置
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      '**/e2e/**',
      'test-ui.spec.ts',
      'playwright-report/**',
      'test-results/**',
    ],
  },
}));
```

### 1.2 添加 package.json 脚本

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "build:analyze": "tsc -b && vite build --mode analyze",
    "lint": "eslint .",
    "preview": "vite preview",
    "test": "vitest",
    "test:run": "vitest --run"
  }
}
```

---

## 第二部分: 动态主题加载实现

### 2.1 创建主题加载 Hook

**文件: `src/hooks/useDynamicTheme.ts`**

```typescript
import { useState, useEffect, useCallback } from 'react';

export type Theme = 'qi-baishi' | 'pop-anime' | 'rococo' | 'japanese-bw';

const THEME_STORAGE_KEY = 'paperbanana-theme';
const THEME_LOAD_TIMEOUT = 5000;

// 迁移映射
const THEME_MIGRATION_MAP: Record<string, Theme> = {
  'pop-art': 'pop-anime',
  'classical-chinese': 'qi-baishi',
  'minimalist-bw': 'japanese-bw',
};

function migrateTheme(oldTheme: string): Theme {
  return THEME_MIGRATION_MAP[oldTheme] || (isValidTheme(oldTheme) ? oldTheme : 'qi-baishi');
}

function isValidTheme(theme: string): theme is Theme {
  return ['qi-baishi', 'pop-anime', 'rococo', 'japanese-bw'].includes(theme);
}

function getInitialTheme(): Theme {
  if (typeof window === 'undefined') return 'qi-baishi';
  
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  if (stored) {
    const migrated = migrateTheme(stored);
    if (migrated !== stored) {
      localStorage.setItem(THEME_STORAGE_KEY, migrated);
    }
    return migrated;
  }
  return 'qi-baishi';
}

interface UseDynamicThemeReturn {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  isLoading: boolean;
  isReady: boolean;
  error: string | null;
  themes: readonly { id: Theme; name: string }[];
}

export function useDynamicTheme(): UseDynamicThemeReturn {
  const [theme, setThemeState] = useState<Theme>(getInitialTheme);
  const [isLoading, setIsLoading] = useState(false);
  const [isReady, setIsReady] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadedThemes, setLoadedThemes] = useState<Set<Theme>>(new Set());

  const loadTheme = useCallback(async (themeToLoad: Theme): Promise<boolean> => {
    // 已加载的主题直接返回
    if (loadedThemes.has(themeToLoad)) {
      return true;
    }

    setIsLoading(true);
    setError(null);

    try {
      // 动态导入主题 CSS
      await import(/* @vite-ignore */ `./../themes/${themeToLoad}.css`);
      
      setLoadedThemes(prev => new Set(prev).add(themeToLoad));
      setIsLoading(false);
      return true;
    } catch (err) {
      console.error(`Failed to load theme: ${themeToLoad}`, err);
      setError(`Failed to load theme: ${themeToLoad}`);
      setIsLoading(false);
      return false;
    }
  }, [loadedThemes]);

  const setTheme = useCallback(async (newTheme: Theme) => {
    const success = await loadTheme(newTheme);
    if (success) {
      setThemeState(newTheme);
      localStorage.setItem(THEME_STORAGE_KEY, newTheme);
      document.documentElement.setAttribute('data-theme', newTheme);
    } else if (newTheme !== 'qi-baishi') {
      // 回退到默认主题
      const fallbackSuccess = await loadTheme('qi-baishi');
      if (fallbackSuccess) {
        setThemeState('qi-baishi');
        localStorage.setItem(THEME_STORAGE_KEY, 'qi-baishi');
        document.documentElement.setAttribute('data-theme', 'qi-baishi');
      }
    }
  }, [loadTheme]);

  // 初始加载
  useEffect(() => {
    let isMounted = true;
    
    const initTheme = async () => {
      const initialTheme = getInitialTheme();
      const success = await loadTheme(initialTheme);
      if (isMounted) {
        setIsReady(success);
        document.documentElement.setAttribute('data-theme', initialTheme);
      }
    };

    initTheme();

    return () => {
      isMounted = false;
    };
  }, [loadTheme]);

  return {
    theme,
    setTheme,
    isLoading,
    isReady,
    error,
    themes: [
      { id: 'qi-baishi' as const, name: 'Qi Baishi' },
      { id: 'pop-anime' as const, name: 'Pop Anime' },
      { id: 'rococo' as const, name: 'Rococo' },
      { id: 'japanese-bw' as const, name: 'Night Mono' },
    ] as const,
  };
}
```

### 2.2 更新主题选择器组件

**文件: `src/components/ThemeSelector.tsx`**

```tsx
import { useDynamicTheme } from '../hooks/useDynamicTheme';

export function ThemeSelector() {
  const { theme, setTheme, isLoading, themes } = useDynamicTheme();

  return (
    <div className="theme-selector">
      <label className="text-sm font-medium text-muted-foreground">
        Theme
      </label>
      <select
        value={theme}
        onChange={(e) => setTheme(e.target.value as typeof theme)}
        disabled={isLoading}
        className="w-full mt-1 px-3 py-2 rounded-md border border-border bg-card"
      >
        {themes.map((t) => (
          <option key={t.id} value={t.id}>
            {t.name} {theme === t.id && isLoading && '(Loading...)'}
          </option>
        ))}
      </select>
      {isLoading && (
        <div className="mt-2 text-xs text-muted-foreground flex items-center gap-2">
          <div className="w-3 h-3 border border-current border-t-transparent rounded-full animate-spin" />
          Loading theme...
        </div>
      )}
    </div>
  );
}
```

---

## 第三部分: 路由级代码分割

### 3.1 创建懒加载包装组件

**文件: `src/components/LazyRoute.tsx`**

```tsx
import { Suspense, lazy, ComponentType } from 'react';

interface LazyRouteProps {
  component: () => Promise<{ default: ComponentType }>;
  fallback?: React.ReactNode;
}

const DefaultFallback = () => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="flex items-center gap-2 text-muted-foreground">
      <div className="w-5 h-5 border-2 border-current border-t-transparent rounded-full animate-spin" />
      Loading...
    </div>
  </div>
);

export function LazyRoute({ component, fallback }: LazyRouteProps) {
  const LazyComponent = lazy(component);
  
  return (
    <Suspense fallback={fallback || <DefaultFallback />}>
      <LazyComponent />
    </Suspense>
  );
}
```

### 3.2 更新 App.tsx 使用懒加载

```tsx
import { lazy, Suspense } from 'react';
import './themes/base.css'; // 只保留基础样式

// 动态导入页面组件
const ProjectsPage = lazy(() => import('./pages/ProjectsPage'));
const ProviderEditPage = lazy(() => import('./pages/ProviderEditPage'));
const SettingsDrawer = lazy(() => import('./components/settings/SettingsDrawer'));
const WelcomeWizard = lazy(() => import('./components/WelcomeWizard'));
const HistoryPanel = lazy(() => import('./components/history/HistoryPanel'));
const ExportModal = lazy(() => import('./components/ExportModal'));
const ShortcutsHelpPanel = lazy(() => import('./components/ShortcutsHelpPanel'));

// Loading 组件
const PageLoader = () => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="flex items-center gap-3 text-muted-foreground">
      <div className="w-6 h-6 border-2 border-current border-t-transparent rounded-full animate-spin" />
      <span className="text-sm">Loading...</span>
    </div>
  </div>
);

// 在路由渲染处使用 Suspense
if (currentPath === '/projects') {
  return (
    <ErrorBoundary>
      <Suspense fallback={<PageLoader />}>
        <ProjectsPage />
      </Suspense>
      <Toast toasts={toasts} onRemove={removeToast} />
    </ErrorBoundary>
  );
}

// Provider edit/new page
if (currentPage === "provider-new" || currentPage === "provider-edit") {
  return (
    <ErrorBoundary>
      <Suspense fallback={<PageLoader />}>
        <ProviderEditPage
          providerId={editingProvider}
          isNew={currentPage === "provider-new"}
          onBack={() => {
            setEditingProvider(undefined);
            setCurrentPage("main");
            setIsSettingsOpen(true);
          }}
          onSaveSuccess={() => addToast('Provider saved successfully', 'success')}
        />
      </Suspense>
      <Toast toasts={toasts} onRemove={removeToast} />
    </ErrorBoundary>
  );
}

// 在主页面中使用懒加载的组件
<Suspense fallback={null}>
  {showWelcomeWizard && (
    <WelcomeWizard
      onComplete={() => setShowWelcomeWizard(false)}
      onNavigateToSettings={() => {
        setShowWelcomeWizard(false);
        setIsSettingsOpen(true);
      }}
    />
  )}
</Suspense>

<Suspense fallback={null}>
  {isHistoryPanelOpen && (
    <HistoryPanel
      isOpen={isHistoryPanelOpen}
      onClose={() => setIsHistoryPanelOpen(false)}
      onSelectSession={handleSelectSession}
      selectedSessionId={selectedSessionId}
    />
  )}
</Suspense>

<Suspense fallback={null}>
  {isSettingsOpen && (
    <SettingsDrawer
      isOpen={isSettingsOpen}
      onClose={() => setIsSettingsOpen(false)}
    />
  )}
</Suspense>

<Suspense fallback={null}>
  {showExport && (
    <ExportModal
      isOpen={showExport}
      onClose={() => setShowExport(false)}
      imageData={exportArtifact?.data}
    />
  )}
</Suspense>
```

---

## 第四部分: 性能监控实现

### 4.1 Core Web Vitals 监控

**文件: `src/lib/performance.ts`**

```typescript
// Core Web Vitals 监控
interface WebVitalsMetric {
  name: 'CLS' | 'FID' | 'FCP' | 'LCP' | 'TTFB';
  value: number;
  rating: 'good' | 'needs-improvement' | 'poor';
}

const THRESHOLDS: Record<string, { good: number; poor: number }> = {
  CLS: { good: 0.1, poor: 0.25 },
  FID: { good: 100, poor: 300 },
  FCP: { good: 1800, poor: 3000 },
  LCP: { good: 2500, poor: 4000 },
  TTFB: { good: 600, poor: 1800 },
};

function getRating(name: string, value: number): WebVitalsMetric['rating'] {
  const threshold = THRESHOLDS[name];
  if (!threshold) return 'good';
  
  if (value <= threshold.good) return 'good';
  if (value <= threshold.poor) return 'needs-improvement';
  return 'poor';
}

export function observeWebVitals(onMetric?: (metric: WebVitalsMetric) => void) {
  if (typeof window === 'undefined') return;

  // CLS - Cumulative Layout Shift
  if ('PerformanceObserver' in window) {
    try {
      const clsObserver = new PerformanceObserver((list) => {
        let clsValue = 0;
        for (const entry of list.getEntries()) {
          if (!(entry as any).hadRecentInput) {
            clsValue += (entry as any).value;
          }
        }
        const metric: WebVitalsMetric = {
          name: 'CLS',
          value: clsValue,
          rating: getRating('CLS', clsValue),
        };
        console.log('[Web Vitals]', metric);
        onMetric?.(metric);
      });
      clsObserver.observe({ entryTypes: ['layout-shift'] });
    } catch (e) {
      // CLS not supported
    }

    // LCP - Largest Contentful Paint
    try {
      const lcpObserver = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        const lastEntry = entries[entries.length - 1] as PerformanceEntry;
        const metric: WebVitalsMetric = {
          name: 'LCP',
          value: lastEntry.startTime,
          rating: getRating('LCP', lastEntry.startTime),
        };
        console.log('[Web Vitals]', metric);
        onMetric?.(metric);
      });
      lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] });
    } catch (e) {
      // LCP not supported
    }

    // FCP - First Contentful Paint
    try {
      const fcpObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if ((entry as any).name === 'first-contentful-paint') {
            const metric: WebVitalsMetric = {
              name: 'FCP',
              value: entry.startTime,
              rating: getRating('FCP', entry.startTime),
            };
            console.log('[Web Vitals]', metric);
            onMetric?.(metric);
          }
        }
      });
      fcpObserver.observe({ entryTypes: ['paint'] });
    } catch (e) {
      // FCP not supported
    }
  }

  // TTFB - Time to First Byte
  if (window.performance && window.performance.timing) {
    window.addEventListener('load', () => {
      setTimeout(() => {
        const timing = window.performance.timing;
        const ttfb = timing.responseStart - timing.navigationStart;
        const metric: WebVitalsMetric = {
          name: 'TTFB',
          value: ttfb,
          rating: getRating('TTFB', ttfb),
        };
        console.log('[Web Vitals]', metric);
        onMetric?.(metric);
      }, 0);
    });
  }
}

// 组件渲染性能监控
export function measureRender(componentName: string, startTime: number) {
  const duration = performance.now() - startTime;
  
  if (duration > 16.67) { // 超过 60fps 帧时间
    console.warn(
      `[Performance] ${componentName} took ${duration.toFixed(2)}ms to render (>16.67ms)`
    );
  }
  
  return duration;
}

// 资源加载监控
export function observeResourceLoading() {
  if (!('PerformanceObserver' in window)) return;

  try {
    const observer = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        if (entry.duration > 1000) {
          console.warn(
            `[Resource] Slow resource: ${(entry as PerformanceResourceTiming).name} took ${entry.duration.toFixed(0)}ms`
          );
        }
      }
    });
    observer.observe({ entryTypes: ['resource'] });
  } catch (e) {
    // Resource timing not supported
  }
}
```

### 4.2 在应用入口启用监控

**更新 `src/main.tsx`:**

```tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './i18n';
import { App } from './App';
import { observeWebVitals, observeResourceLoading } from './lib/performance';

// 启用性能监控
if (import.meta.env.PROD) {
  observeWebVitals((metric) => {
    // 可以发送到分析服务
    // analytics.track('web_vitals', metric);
  });
  observeResourceLoading();
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
```

---

## 第五部分: React 渲染优化

### 5.1 高性能列表组件

**文件: `src/components/VirtualList.tsx`**

```tsx
import { useRef, useState, useEffect, useCallback, memo } from 'react';

interface VirtualListProps<T> {
  items: T[];
  renderItem: (item: T, index: number) => React.ReactNode;
  itemHeight: number;
  containerHeight: number;
  overscan?: number;
}

export function VirtualList<T>({
  items,
  renderItem,
  itemHeight,
  containerHeight,
  overscan = 5,
}: VirtualListProps<T>) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [scrollTop, setScrollTop] = useState(0);

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop);
  }, []);

  const totalHeight = items.length * itemHeight;
  const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan);
  const endIndex = Math.min(
    items.length,
    Math.ceil((scrollTop + containerHeight) / itemHeight) + overscan
  );

  const visibleItems = items.slice(startIndex, endIndex);
  const offsetY = startIndex * itemHeight;

  return (
    <div
      ref={containerRef}
      onScroll={handleScroll}
      style={{ height: containerHeight, overflow: 'auto' }}
    >
      <div style={{ height: totalHeight, position: 'relative' }}>
        <div style={{ transform: `translateY(${offsetY}px)` }}>
          {visibleItems.map((item, index) => (
            <div key={startIndex + index} style={{ height: itemHeight }}>
              {renderItem(item, startIndex + index)}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
```

### 5.2 Memoized 组件包装

**文件: `src/hooks/useMemoizedComponent.ts`**

```typescript
import { memo, useMemo, ComponentType } from 'react';

export function useMemoizedComponent<T extends object>(
  Component: ComponentType<T>,
  propsEquality?: (prev: T, next: T) => boolean
): ComponentType<T> {
  return useMemo(
    () => memo(Component, propsEquality),
    [Component, propsEquality]
  );
}
```

### 5.3 防抖和节流 Hooks

**文件: `src/hooks/useDebouncedCallback.ts`**

```typescript
import { useCallback, useRef } from 'react';

export function useDebouncedCallback<T extends (...args: any[]) => void>(
  callback: T,
  delay: number
): T {
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>();

  return useCallback(
    (...args: Parameters<T>) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      timeoutRef.current = setTimeout(() => {
        callback(...args);
      }, delay);
    },
    [callback, delay]
  ) as T;
}

export function useThrottledCallback<T extends (...args: any[]) => void>(
  callback: T,
  limit: number
): T {
  const inThrottle = useRef(false);

  return useCallback(
    (...args: Parameters<T>) => {
      if (!inThrottle.current) {
        callback(...args);
        inThrottle.current = true;
        setTimeout(() => {
          inThrottle.current = false;
        }, limit);
      }
    },
    [callback, limit]
  ) as T;
}
```

---

## 第六部分: 实施检查清单

### Phase 1: Vite 配置 (优先级: 🔴 高)

- [ ] 安装 `rollup-plugin-visualizer`
- [ ] 更新 `vite.config.ts` 添加优化配置
- [ ] 添加 `build:analyze` 脚本
- [ ] 测试构建并分析 Bundle

### Phase 2: 主题优化 (优先级: 🔴 高)

- [ ] 创建 `useDynamicTheme.ts`
- [ ] 更新 `ThemeSelector.tsx`
- [ ] 从 `App.tsx` 移除静态主题导入
- [ ] 测试主题切换功能

### Phase 3: 代码分割 (优先级: 🟡 中)

- [ ] 创建 `LazyRoute.tsx`
- [ ] 更新 `App.tsx` 使用 `React.lazy`
- [ ] 测试懒加载行为
- [ ] 验证 Error Boundary

### Phase 4: 性能监控 (优先级: 🟢 低)

- [ ] 创建 `performance.ts`
- [ ] 更新 `main.tsx` 启用监控
- [ ] 验证 Web Vitals 数据收集

---

## 第七部分: 验证步骤

### 构建验证

```bash
# 1. 构建项目
npm run build

# 2. 分析 Bundle
npm run build:analyze

# 3. 检查输出
dir dist/assets

# 4. 预期结果:
# - JS 文件应该有多个 chunk (vendor, i18n, main)
# - CSS 文件应该与 JS 分离
# - 总大小应该比优化前小 20-30%
```

### 运行时验证

```bash
# 1. 启动预览服务器
npm run preview

# 2. 打开 DevTools > Network
# 3. 验证:
# - 首屏只加载必要资源
# - 主题 CSS 按需加载
# - 路由切换时加载对应 chunk

# 4. 检查 Console 中的性能日志
# - Web Vitals 指标
# - 慢渲染警告
```

---

*实施指南版本: 1.0*  
*最后更新: 2026-04-06*
