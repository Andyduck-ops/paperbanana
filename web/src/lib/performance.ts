/**
 * 性能监控工具库
 * 
 * 提供 Core Web Vitals 监控、组件渲染监控和资源加载监控
 */

// Core Web Vitals 指标类型
type WebVitalsName = 'CLS' | 'FID' | 'FCP' | 'LCP' | 'TTFB' | 'INP';

interface WebVitalsMetric {
  name: WebVitalsName;
  value: number;
  rating: 'good' | 'needs-improvement' | 'poor';
  entries: PerformanceEntry[];
}

// Core Web Vitals 阈值定义 (基于 Google 标准)
const THRESHOLDS: Record<WebVitalsName, { good: number; poor: number }> = {
  CLS: { good: 0.1, poor: 0.25 },
  FID: { good: 100, poor: 300 },
  FCP: { good: 1800, poor: 3000 },
  LCP: { good: 2500, poor: 4000 },
  TTFB: { good: 800, poor: 1800 },
  INP: { good: 200, poor: 500 },
};

/**
 * 获取指标评级
 */
function getRating(name: WebVitalsName, value: number): WebVitalsMetric['rating'] {
  const threshold = THRESHOLDS[name];
  if (!threshold) return 'good';
  
  if (value <= threshold.good) return 'good';
  if (value <= threshold.poor) return 'needs-improvement';
  return 'poor';
}

/**
 * 观察 Core Web Vitals
 * 
 * @param onMetric - 指标回调函数
 * @param options - 配置选项
 */
export function observeWebVitals(
  onMetric?: (metric: WebVitalsMetric) => void,
  options: { logToConsole?: boolean; reportUrl?: string } = {}
): () => void {
  const { logToConsole = true } = options;
  
  if (typeof window === 'undefined' || !('PerformanceObserver' in window)) {
    return () => {};
  }

  const observers: PerformanceObserver[] = [];

  // CLS - Cumulative Layout Shift
  try {
    let clsValue = 0;
    const clsEntries: PerformanceEntry[] = [];
    
    const clsObserver = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        const layoutShift = entry as any;
        if (!layoutShift.hadRecentInput) {
          clsValue += layoutShift.value;
          clsEntries.push(entry);
        }
      }
      
      const metric: WebVitalsMetric = {
        name: 'CLS',
        value: clsValue,
        rating: getRating('CLS', clsValue),
        entries: clsEntries,
      };
      
      if (logToConsole) {
        console.log('[Web Vitals]', metric);
      }
      onMetric?.(metric);
    });
    
    clsObserver.observe({ type: 'layout-shift', buffered: true });
    observers.push(clsObserver);
  } catch (e) {
    // CLS 不支持
  }

  // LCP - Largest Contentful Paint
  try {
    const lcpObserver = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      const lastEntry = entries[entries.length - 1];
      
      const metric: WebVitalsMetric = {
        name: 'LCP',
        value: lastEntry.startTime,
        rating: getRating('LCP', lastEntry.startTime),
        entries: [lastEntry],
      };
      
      if (logToConsole) {
        console.log('[Web Vitals]', metric);
      }
      onMetric?.(metric);
    });
    
    lcpObserver.observe({ type: 'largest-contentful-paint', buffered: true });
    observers.push(lcpObserver);
  } catch (e) {
    // LCP 不支持
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
            entries: [entry],
          };
          
          if (logToConsole) {
            console.log('[Web Vitals]', metric);
          }
          onMetric?.(metric);
        }
      }
    });
    
    fcpObserver.observe({ type: 'paint', buffered: true });
    observers.push(fcpObserver);
  } catch (e) {
    // FCP 不支持
  }

  // INP - Interaction to Next Paint
  try {
    let inpValue = 0;
    const inpObserver = new PerformanceObserver((list) => {
      const entries = list.getEntries() as PerformanceEventTiming[];
      for (const entry of entries) {
        if ((entry as PerformanceEventTiming & { interactionId?: number }).interactionId! > 0) {
          const duration = entry.processingEnd - entry.startTime;
          if (duration > inpValue) {
            inpValue = duration;
            
            const metric: WebVitalsMetric = {
              name: 'INP',
              value: inpValue,
              rating: getRating('INP', inpValue),
              entries: [entry],
            };
            
            if (logToConsole) {
              console.log('[Web Vitals]', metric);
            }
            onMetric?.(metric);
          }
        }
      }
    });
    
    inpObserver.observe({ type: 'event', buffered: true, durationThreshold: 0 } as PerformanceObserverInit);
    observers.push(inpObserver);
  } catch (e) {
    // INP 不支持
  }

  // TTFB - Time to First Byte (使用 Navigation Timing)
  if (window.performance && window.performance.timing) {
    const reportTTFB = () => {
      const timing = window.performance.timing;
      const ttfb = timing.responseStart - timing.navigationStart;
      
      if (ttfb > 0) {
        const metric: WebVitalsMetric = {
          name: 'TTFB',
          value: ttfb,
          rating: getRating('TTFB', ttfb),
          entries: [],
        };
        
        if (logToConsole) {
          console.log('[Web Vitals]', metric);
        }
        onMetric?.(metric);
      }
    };
    
    if (document.readyState === 'complete') {
      reportTTFB();
    } else {
      window.addEventListener('load', reportTTFB);
    }
  }

  // 返回清理函数
  return () => {
    observers.forEach(observer => observer.disconnect());
  };
}

/**
 * 测量组件渲染时间
 * 
 * @param componentName - 组件名称
 * @returns 结束测量函数，调用后返回渲染耗时
 * 
 * 使用示例:
 * ```tsx
 * function MyComponent() {
 *   const endMeasure = startRenderMeasure('MyComponent');
 *   
 *   useEffect(() => {
 *     const duration = endMeasure();
 *     // duration 为渲染耗时 (ms)
 *   });
 *   
 *   return <div>...</div>;
 * }
 * ```
 */
export function startRenderMeasure(componentName: string): () => number {
  const startTime = performance.now();
  
  return () => {
    const duration = performance.now() - startTime;
    
    // 如果超过 16.67ms (60fps 帧时间)，输出警告
    if (duration > 16.67) {
      console.warn(
        `[Performance] ${componentName} render took ${duration.toFixed(2)}ms (>16.67ms)`
      );
    }
    
    return duration;
  };
}

/**
 * 观察资源加载性能
 * 
 * 检测加载缓慢的资源并输出警告
 * 
 * @param options - 配置选项
 */
export function observeResourceLoading(
  options: { 
    slowThreshold?: number; 
    logToConsole?: boolean;
    includeTypes?: string[];
  } = {}
): () => void {
  const { 
    slowThreshold = 1000, 
    logToConsole = true,
    includeTypes = ['script', 'stylesheet', 'image', 'font']
  } = options;
  
  if (!('PerformanceObserver' in window)) {
    return () => {};
  }

  let observer: PerformanceObserver | null = null;

  try {
    observer = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        const resource = entry as PerformanceResourceTiming;
        
        // 只监控指定类型的资源
        if (!includeTypes.includes(resource.initiatorType)) {
          continue;
        }
        
        // 检测加载缓慢的资源
        if (resource.duration > slowThreshold) {
          if (logToConsole) {
            console.warn(
              `[Resource] Slow ${resource.initiatorType}: ${resource.name} ` +
              `took ${resource.duration.toFixed(0)}ms ` +
              `(transferSize: ${(resource.transferSize / 1024).toFixed(1)}KB)`
            );
          }
        }
        
        // 检测缓存未命中的资源
        if (resource.transferSize > 0 && (resource as PerformanceResourceTiming & { deliveryType?: string }).deliveryType === '') {
          // 可以在这里收集缓存命中率数据
        }
      }
    });
    
    observer.observe({ type: 'resource', buffered: true });
  } catch (e) {
    console.warn('[Performance] Resource timing not supported');
  }

  return () => {
    observer?.disconnect();
  };
}

/**
 * 获取导航时间信息
 * 
 * 返回页面加载各阶段的时间数据
 */
export function getNavigationTiming(): Record<string, number> | null {
  if (typeof window === 'undefined' || !window.performance || !window.performance.timing) {
    return null;
  }

  const timing = window.performance.timing;
  
  return {
    // DNS 查询时间
    dns: timing.domainLookupEnd - timing.domainLookupStart,
    // TCP 连接时间
    tcp: timing.connectEnd - timing.connectStart,
    // 首字节时间 (TTFB)
    ttfb: timing.responseStart - timing.navigationStart,
    // 下载时间
    download: timing.responseEnd - timing.responseStart,
    // DOM 解析时间
    domParse: timing.domInteractive - timing.responseEnd,
    // DOM 就绪时间
    domReady: timing.domContentLoadedEventEnd - timing.navigationStart,
    // 完全加载时间
    loadComplete: timing.loadEventEnd - timing.navigationStart,
  };
}

/**
 * 性能标记和测量工具
 * 
 * 使用 Performance API 进行自定义测量
 */
export const performanceMarks = {
  /**
   * 添加性能标记
   */
  mark: (name: string) => {
    if (typeof performance !== 'undefined' && performance.mark) {
      performance.mark(name);
    }
  },
  
  /**
   * 测量两个标记之间的时间
   */
  measure: (name: string, startMark: string, endMark?: string) => {
    if (typeof performance !== 'undefined' && performance.measure) {
      try {
        performance.measure(name, startMark, endMark);
        const entries = performance.getEntriesByName(name, 'measure');
        return entries[entries.length - 1]?.duration;
      } catch (e) {
        console.warn(`[Performance] Failed to measure ${name}:`, e);
      }
    }
    return null;
  },
  
  /**
   * 清除所有标记和测量
   */
  clear: () => {
    if (typeof performance !== 'undefined' && performance.clearMarks) {
      performance.clearMarks();
      performance.clearMeasures();
    }
  },
};

/**
 * 初始化所有性能监控
 * 
 * 在应用入口调用此函数启用完整的性能监控
 */
export function initPerformanceMonitoring(options: {
  enableWebVitals?: boolean;
  enableResourceMonitoring?: boolean;
  logToConsole?: boolean;
} = {}): () => void {
  const { 
    enableWebVitals = true, 
    enableResourceMonitoring = true,
    logToConsole = true,
  } = options;

  const cleanupFns: (() => void)[] = [];

  if (enableWebVitals) {
    cleanupFns.push(observeWebVitals(undefined, { logToConsole }));
  }

  if (enableResourceMonitoring) {
    cleanupFns.push(observeResourceLoading({ logToConsole }));
  }

  // 输出导航时间信息
  if (logToConsole && typeof window !== 'undefined') {
    window.addEventListener('load', () => {
      setTimeout(() => {
        const timing = getNavigationTiming();
        if (timing) {
          console.log('[Performance] Navigation Timing:', timing);
        }
      }, 0);
    });
  }

  // 返回清理函数
  return () => {
    cleanupFns.forEach(fn => fn());
  };
}
