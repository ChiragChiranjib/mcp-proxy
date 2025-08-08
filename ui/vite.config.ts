import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
      '/live': 'http://localhost:8080',
      '/ready': 'http://localhost:8080',
      '/servers': 'http://localhost:8080',
    },
  },
})

