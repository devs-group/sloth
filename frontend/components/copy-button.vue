<script lang="ts" setup>
import { useClipboard } from "@vueuse/core/index";

const props = defineProps({
  string: {
    required: true
  }
})

const { copy, copied, isSupported } = useClipboard()

const defaultCopyIcon = "i-heroicons-document-duplicate"

const copyIcon = ref(defaultCopyIcon)

async function copyToClipboard(s: string) {
  await copy(s)
  if (copied) {
    copyIcon.value = "i-heroicons-check"
    setTimeout(() => {
      copyIcon.value = defaultCopyIcon
    }, 1000)
  }
}
</script>

<template>
  <UButton
      v-if="isSupported"
      :icon="copyIcon"
      :ui="{ rounded: 'rounded-full' }"
      variant="ghost"
      @click="copyToClipboard(props.string)"
  ></UButton>
</template>