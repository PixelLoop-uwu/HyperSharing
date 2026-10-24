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
      // tus-загрузки живут на отдельном порту backend (TUS_HTTP_ADDR), чтобы
      // долгие PATCH не делили таймауты с JSON-эндпоинтами.
      '/api/v1/uploads': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      // Cookie сессия живёт на origin фронта — проксируем, чтобы не настраивать CORS.
      '/api': {
        target: 'http://localhost:8800',
        changeOrigin: true,
      },
    },
  },
})
