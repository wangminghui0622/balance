import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig(({ command, mode }) => {
  // 开发环境 base 为空，生产环境使用 /balance/admin/
  const base = command === 'build' ? '/balance/admin/' : '/'
  
  return {
    base,
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@share': resolve(__dirname, 'share'),
      '@operator': resolve(__dirname, 'operator'),
      '@platform': resolve(__dirname, 'platform'),
      '@shopowner': resolve(__dirname, 'shopower')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:19090', // admin服务端口
        changeOrigin: true
        // 不重写路径，保持 /api/v1/auth/login 格式
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,  // 生产环境关闭 sourcemap
    minify: 'esbuild',  // 使用 esbuild（更快，无需额外依赖）
    // 如果需要使用 terser，需要安装: npm install -D terser
    // minify: 'terser',
    // terserOptions: {
    //   compress: {
    //     drop_console: true,
    //     drop_debugger: true
    //   }
    // },
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          'echarts': ['echarts', 'vue-echarts'],
          'vue-vendor': ['vue', 'vue-router', 'axios']
        }
      }
    }
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: 'modern-compiler'
      }
    }
  }
  }
})
