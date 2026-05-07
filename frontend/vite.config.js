import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  base: "/",
  plugins: [react()],
  preview: {
    port: "5137",
    strictPort: true,
  },
  server: {
    port: 5137,
    strictPort: true,
    host: true,
    origin: "http://0.0.0.0:5137",
  },
})
