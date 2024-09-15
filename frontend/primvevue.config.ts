import Aura from "@primevue/themes/aura";

export const primeVueConfig = {
  options: {
    theme: {
      preset: Aura,
    },
  },
  cssLayerOrder: "tailwind-base, primevue, tailwind-utilities",
  css: ["primeicons/primeicons.css"],
};
