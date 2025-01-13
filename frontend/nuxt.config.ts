import { resolve } from 'path'
import fs from 'fs'
import { primeVueConfig } from './primvevue.config'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: ['@primevue/nuxt-module', '@nuxtjs/tailwindcss', '@nuxt/eslint'],
  ssr: false,
  devtools: { enabled: false, telemetry: false },
  app: {
    baseURL: '/_/',
  },
  css: ['primeicons/primeicons.css', '~/assets/css/overrides.css'],
  runtimeConfig: {
    backendHost: 'http://localhost:9090',
    public: {
      backendHost: 'http://localhost:9090',
      serverIp: '127.0.0.1',
    },
  },
  compatibilityDate: '2025-01-12',
  hooks: {
    'vite:compiled': () => {
      // HACK FOR DEVELOPMENT to make sure the backend can start
      if (process.env.NODE_ENV !== 'development') {
        return
      }
      const outputDir = resolve(__dirname, '.output/public')
      const placeholderFileName = 'index.html'
      const placeholderFilePath = resolve(outputDir, placeholderFileName)

      if (!fs.existsSync(outputDir)) {
        fs.mkdirSync(outputDir, { recursive: true })
        console.log(`Created directory: ${outputDir}`)
      }

      // Watch for deletions and recreate the placeholder file
      fs.watch(outputDir, (eventType, filename) => {
        if (eventType === 'rename' && filename === placeholderFileName && !fs.existsSync(placeholderFilePath)) {
          fs.writeFileSync(placeholderFilePath, '<p style="color: red">DEV: You should not see this page, the proxy does not seem to work</p>')
          console.log(`Recreated placeholder file: ${placeholderFileName}`)
        }
      })

      fs.writeFileSync(placeholderFilePath, '<p>DEV: You should not see this page, the proxy does not seem to work</p>')
      console.log(`Created placeholder file: ${placeholderFileName}`)
    },
  },
  primevue: primeVueConfig,

})
