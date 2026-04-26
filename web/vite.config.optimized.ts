/**
 * 优化后的 Vite 配置
 * 
 * 使用方法:
 * 1. 备份原有的 vite.config.ts
 * 2. 将此文件重命名为 vite.config.ts
 * 3. 安装依赖: npm install -D rollup-plugin-visualizer
 * 4. 运行构建: npm run build:analyze
 */

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

// 可选: Bundle 分析器 (取消注释以启用)
// import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig(({ mode }) => ({
  plugins: [
    react(),
    tailwindcss(),
    // Bundle 分析器 (仅在 analyze 模式下启用)
    // mode === 'analyze' && visualizer({
    //   open: true,
    //   gzipSize: true,
    //   brotliSize: true,
    //   filename: 'dist/stats.html',
    // }),
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
    
    // 启用 CSS 代码分割
    cssCodeSplit: true,
    
    // 生成 sourcemap (生产环境可关闭)
    sourcemap: mode !== 'production',
    
    // 资源内联阈值: 4KB 以下的资源内联到 bundle 中
    assetsInlineLimit: 4096,
    
    // Chunk 大小警告阈值 (KB)
    chunkSizeWarningLimit: 500,

    rollupOptions: {
      output: {
        // 手动分块策略 - 关键优化点
        manualChunks: {
          // 框架核心依赖 - 稳定不变，长期缓存
          'vendor': ['react', 'react-dom'],
          // i18n 相关 - 独立缓存
          'i18n': ['i18next', 'react-i18next'],
        },
        
        // 入口文件命名
        entryFileNames: 'js/[name]-[hash].js',
        
        // Chunk 文件命名
        chunkFileNames: 'js/[name]-[hash].js',
        
        // 资源文件命名 - 按类型分组
        assetFileNames: (assetInfo) => {
          const info = assetInfo.name.split('.');
          const ext = info[info.length - 1];
          
          // CSS 文件放入 css/ 目录
          if (/\.css$/i.test(assetInfo.name)) {
            return 'css/[name]-[hash][extname]';
          }
          
          // 图片资源放入 images/ 目录
          if (/\.(png|jpe?g|gif|svg|webp|avif)$/i.test(assetInfo.name)) {
            return 'images/[name]-[hash][extname]';
          }
          
          // 字体文件放入 fonts/ 目录
          if (/\.(woff2?|ttf|otf|eot)$/i.test(assetInfo.name)) {
            return 'fonts/[name]-[hash][extname]';
          }
          
          // 其他资源
          return 'assets/[name]-[hash][extname]';
        },
      },
    },

    // 压缩配置 (生产环境)
    minify: mode === 'production' ? 'terser' : false,
    terserOptions: mode === 'production' ? {
      compress: {
        drop_console: true,
        drop_debugger: true,
        // 移除死代码
        dead_code: true,
        // 优化重复代码
        merge_vars: true,
        // 优化循环
        loops: true,
      },
      mangle: {
        // 压缩属性名
        properties: false,
      },
      format: {
        comments: false,
      },
    } : undefined,
  },

  // 依赖预构建优化
  optimizeDeps: {
    include: ['react', 'react-dom', 'i18next', 'react-i18next'],
    exclude: [],
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
