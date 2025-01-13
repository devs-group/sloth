<script setup lang="ts">
const { label, icon, onClick, ...buttonAttrs } = useAttrs()
</script>

<template>
  <Button
    class="flex gap-2"
    v-bind="buttonAttrs"
  >
    <Icon
      style="font-size: 20px;"
      :icon="icon"
    />
    <span
      v-if="label"
      class="p-button-label"
    >{{ label }}</span>
    <slot v-else />
  </Button>
</template>
