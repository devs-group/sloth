<template>
  <div class="flex flex-col w-full relative">
    <div class="flex items-center p-4 gap-2">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold">
          {{ props.title }}
        </h1>
        <div v-if="props.description?.length">
          <p class="text-sm text-gray-400 mt-2">
            {{ props.description }}
          </p>
        </div>
      </div>
      <div>
        <slot name="actions" />
      </div>
    </div>
    <hr>
    <div class="relative flex-1 p-4 pt-4">
      <slot name="content" />
    </div>
    <slot name="loader" />
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
    required: false,
  },
})
</script>
