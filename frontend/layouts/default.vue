<template>
  <div class="flex flex-col h-screen">
    <TopBar />
    <div class="flex flex-1">
      <Navigation />
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
</script>
