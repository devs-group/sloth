import { primeVueConfig } from "./primvevue.config";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  app: {
    baseURL: "/_/",
  },
  devtools: { enabled: true },
  modules: ["@nuxtjs/tailwindcss", "@primevue/nuxt-module"],
  runtimeConfig: {
    backendHost: "http://localhost:9090",
    public: {
      backendHost: "http://localhost:9090",
    },
  },
  primevue: primeVueConfig,
  css: ["primeicons/primeicons.css", "~/assets/css/overrides.css"],
});
