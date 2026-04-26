import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { visualizer } from 'rollup-plugin-visualizer';

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(), 
    tailwindcss(),
    // Bundle visualizer - only in development (generates stats.html for analysis)
    ...(process.env.NODE_ENV !== 'production' ? [
      visualizer({
        open: false,
        gzipSize: true,
        brotliSize: true,
        filename: 'stats.html',
      }),
    ] : []),
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
  build: {
    // Source maps disabled in production for security
    sourcemap: false,
    rollupOptions: {
      output: {
        // Manual chunk splitting for better caching
        manualChunks: {
          'vendor': ['react', 'react-dom'],
          'i18n': ['i18next', 'react-i18next'],
        },
      },
    },
  },
});
