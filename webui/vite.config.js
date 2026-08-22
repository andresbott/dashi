import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// The Vue app is the admin/editor UI. In dev it talks only to the backend's
// admin port (the editor server on :8088). The public, server-rendered
// dashboard view lives on the separate public port (:8087) and is not served
// through this dev server at all.
const adminTarget = 'http://localhost:8088'

const proxyToAdmin = {
  target: adminTarget,
  changeOrigin: true,
  secure: false,
  cookieDomainRewrite: { '*': '' }
}

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
  base: "./",
  server: {
    proxy: {
      '/auth': proxyToAdmin,
      '/api': proxyToAdmin
    }
  }
})
