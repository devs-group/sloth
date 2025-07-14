<template>
  <div
    class="flex flex-col gap-2 justify-center items-center bg-black bg-opacity-5 backdrop-blur-sm left-0 right-0 top-0 bottom-0 pt-24 opacity-0 pointer-events-none transition-opacity duration-300"
    :class="{ '!opacity-100 !pointer-events-auto': showDelayed, 'fixed z-50': isFixed, 'absolute z-40': !isFixed }"
  >
    <ClientOnly>
      <Vue3Lottie
        :animation-data="animationData"
        height="16rem"
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
import { Vue3Lottie } from 'vue3-lottie'

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
  text: {
    type: String,
    default: '',
  },
  isFixed: {
    type: Boolean,
    required: true,
  },
})

const showDelayed = ref(false)
watch(() => props.show, (value) => {
  setTimeout(() => {
    showDelayed.value = value == true
  }, 150)
}, { immediate: true })
</script>
