<template>
  <div class="flex flex-col h-screen">
    <TopBar></TopBar>
    <div class="flex flex-1">
      <Navigation></Navigation>
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>