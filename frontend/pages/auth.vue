<template>
  <NuxtPage />
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'auth',
})

useHead({
  title() {
    return 'Auth'
  },
})
</script>
