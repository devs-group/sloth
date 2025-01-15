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
    backendHost: 'http://localhost',
    public: {
      backendHost: 'http://localhost',
      serverIp: '127.0.0.1',
    },
  },
  compatibilityDate: '2025-01-12',
  primevue: primeVueConfig,
})
