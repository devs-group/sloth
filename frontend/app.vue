<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <DynamicDialog />
  <Toast position="bottom-center" />
  <OverlayProgressSpinner
    :show="globalSpinner"
    :text="globalText"
    :is-fixed="true"
  />
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
      rel: 'shortcut icon',
      href: '/_/favicon.ico',
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
