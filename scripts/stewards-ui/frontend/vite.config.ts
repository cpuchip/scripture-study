import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    // Dev: proxy /api/* to the Go binary running on localhost:8080.
    // Build sets up `npm run dev` to feel like prod.
    proxy: {
      '/api': 'http://localhost:8080',
      '/healthz': 'http://localhost:8080',
    },
  },
  build: {
    // Output goes into ../dist so the Go embed.FS picks it up at
    // `frontend/dist/`. (Default would be `dist/` inside frontend/.)
    outDir: 'dist',
    emptyOutDir: true,
  },
})
