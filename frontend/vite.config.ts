import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/api/v1/uploads': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '^/f(?:/|$)': {
        target: 'http://localhost:8800',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/f/, '/api/v1/f'),
      },
      '^/avatars(?:/|$)': {
        target: 'http://localhost:8800',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/avatars/, '/api/v1/avatars'),
      },
      '/api': {
        target: 'http://localhost:8800',
        changeOrigin: true,
      },
    },
  },
})
