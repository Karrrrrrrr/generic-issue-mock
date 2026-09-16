import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(() => {
  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 3000,
      proxy: {
        '^/slash/ui': 'http://127.0.0.1:8000',
        '^/photonpay/ui': 'http://127.0.0.1:8000',
        '^/paynda/ui': 'http://127.0.0.1:8000',
      },
    },
  }
})
