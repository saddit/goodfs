import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import VueRouter from 'unplugin-vue-router/vite'
import { VueRouterAutoImports } from 'unplugin-vue-router'
import Components from 'unplugin-vue-components/vite'
import vueI18n from '@intlify/vite-plugin-vue-i18n'
import { HeadlessUiResolver } from 'unplugin-vue-components/resolvers'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    vueI18n({
      include: resolve(__dirname, 'src/locale/**')
    }),
    VueRouter({ importMode: 'sync' }),
    vue(),
    Components({ resolvers: [HeadlessUiResolver()] }),
    AutoImport({
      imports: [
          'vue', '@vueuse/head', '@vueuse/core', VueRouterAutoImports,
        {
          '@/api/': [['default', 'api']]
        },
        {
          '@/pkg/': [['default','pkg']]
        },
        {
          'vue-toastification':[['useToast', 'useToast']]
        },
        {
          '@/store':[['useStore', 'useStore']]
        },
        {
          'vue-i18n':[['useI18n', 'useI18n']]
        }
      ],
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    open: true,
  },
  build: {
    outDir: "../resource/web",
    emptyOutDir: true
  }
})
