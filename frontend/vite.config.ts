import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    extensions: ['.ts', '.js', '.json', '.vue']
  },
  // Tauri 在开发模式下会注入固定端口，关闭 host 校验
  clearScreen: false,
  server: {
    port: 5173,
    strictPort: false,
  },
  build: {
    target: 'es2021',
    cssTarget: 'chrome100'
  }
})
