<template>
  <div
    class="flex flex-col gap-2 items-center bg-black bg-opacity-20 backdrop-blur-sm absolute z-40 left-0 right-0 top-0 bottom-0 pt-24 opacity-0 pointer-events-none transition-opacity duration-300"
    :class="{ '!opacity-100 !pointer-events-auto': show }"
  >
    <ClientOnly>
      <Vue3Lottie
        :animation-data="animationData"
        :height="200"
        :width="200"
      />
    </ClientOnly>
    <p
      v-if="text.length"
      class="text-lg"
    >
      {{ text }}
    </p>
  </div>
</template>

<script setup lang="ts">
import animationData from 'assets/lottie/sloth_meditate.json'

defineProps({
  show: {
    type: Boolean,
    required: true,
  },
  text: {
    type: String,
    default: '',
  },
})
</script>
