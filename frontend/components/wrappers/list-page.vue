<template>
  <div class="w-full">
    <div class="flex justify-between items-center p-6">
      <div>
        <h1 class="text-2xl">
          {{ props.title }}
        </h1>
        <div>
          <p class="text-sm text-gray-400 mt-2">
            {{ props.description }}
          </p>
        </div>
      </div>
      <div>
        <slot name="actions" />
      </div>
    </div>
    <div class="relative">
      <slot name="content" />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    required: true,
  },
})
</script>
