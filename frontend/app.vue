<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <DynamicDialog />
  <Toast position="bottom-center" />
  <div
    class="flex flex-col gap-2 justify-center items-center bg-black bg-opacity-20 backdrop-blur-sm fixed z-50 left-0 right-0 top-0 bottom-0 opacity-0 pointer-events-none transition-opacity duration-300"
    :class="{ '!opacity-100 !pointer-events-auto': globalSpinner }"
  >
    <ProgressSpinner />
    <p
      v-if="globalText"
      class="text-lg"
    >
      {{ globalText }}
    </p>
  </div>
</template>

<script setup lang="ts">
const { globalSpinner, globalText } = useGlobalSpinner()
const { isAuthenticated, getUser } = useAuth()

if (isAuthenticated.value) {
  // We fetch the user ones after page load, before render
  await getUser()
}

useHead({
  titleTemplate: (titleChunk) => {
    return titleChunk ? `Sloth - ${titleChunk}` : 'Sloth'
  },
  meta: [
    {
      name: 'msapplication-TileColor',
      content: '#d8d3ca',
    },
    {
      name: 'theme-color',
      content: '#d8d3ca',
    },
  ],
  link: [
    {
      rel: 'apple-touch-icon',
      sizes: '180x180',
      href: '/_/apple-touch-icon.png',
    },
    {
      rel: 'icon',
      type: 'image/png',
      sizes: '32x32',
      href: '/_/favicon-32x32.png',
    },
    {
      rel: 'icon',
      type: 'image/png',
      sizes: '16x16',
      href: '/_/favicon-16x16.png',
    },
    {
      rel: 'mask-icon',
      color: '#000000',
      href: '/_/safari-pinned-tab.svg',
    },
    {
      rel: 'manifest',
      href: '/_/site.webmanifest',
    },
  ],
})
</script>
