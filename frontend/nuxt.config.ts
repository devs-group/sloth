// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  app: {
    baseURL: "/_/"
  },
  devtools: { enabled: true },
  modules: ['@nuxt/ui'],
  runtimeConfig: {
    backendHost: "http://localhost:8080",
    public: {
      backendHost: "http://localhost:8080",
    }
  },
  ui: {
    global: true,
    icons: ['mdi']
  }
})
