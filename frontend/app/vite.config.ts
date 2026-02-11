import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'
import legacy from '@vitejs/plugin-legacy'
import path from 'path'

export default defineConfig({
  plugins: [
    uni(),
    legacy({
      targets: ['defaults', 'not IE 11', 'chrome >= 52', 'safari >= 10'],
      additionalLegacyPolyfills: ['regenerator-runtime/runtime'],
      modernPolyfills: true
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@share': path.resolve(__dirname, 'share'),
      '@shopower': path.resolve(__dirname, 'shopower'),
      '@operator': path.resolve(__dirname, 'operator'),
      '@platform': path.resolve(__dirname, 'platform')
    }
  },
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'http://localhost:19090',
        changeOrigin: true
      }
    }
  }
})
