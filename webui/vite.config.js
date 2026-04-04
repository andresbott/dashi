import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  base: "/",
  server: {
    proxy: {
      '/auth': {
        target: 'http://localhost:8087',
        changeOrigin: true,
        secure: false,
        cookieDomainRewrite: { '*': '' }
      },
      '/api': {
        target: 'http://localhost:8087',
        changeOrigin: true,
        secure: false,
        cookieDomainRewrite: { '*': '' }
      }
    }
  }
})
