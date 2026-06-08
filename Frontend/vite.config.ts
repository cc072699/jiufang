import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'
import { unlinkSync, existsSync } from 'fs'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    {
      name: 'exclude-msw-sw',
      closeBundle() {
        // Remove MSW service worker from production build
        const swPath = resolve(__dirname, 'dist/mockServiceWorker.js')
        if (existsSync(swPath)) {
          unlinkSync(swPath)
        }
      },
    },
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
})
