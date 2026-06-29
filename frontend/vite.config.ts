import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import wails from '@wailsio/runtime/plugins/vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), wails('bindings')],
  resolve: {
    extensions: ['.ts', '.js', '.json', '.vue']
  },
  build: {
    target: 'es2021',
    cssTarget: 'chrome100'
  }
})
