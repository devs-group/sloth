import {primeVueConfig} from "./primvevue.config"

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  app: {
    baseURL: "/_/"
  },
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss', 'nuxt-primevue'],
  runtimeConfig: {
    backendHost: "http://localhost:8080",
    public: {
      backendHost: "http://localhost:8080",
    }
  },
  primevue: primeVueConfig,
  css: [
    'primevue/resources/themes/aura-dark-green/theme.css',
    'primeicons/primeicons.css',
],
})
