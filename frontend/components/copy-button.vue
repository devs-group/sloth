<script lang="ts" setup>
import { useClipboard } from "@vueuse/core"

const props = defineProps<{
  string: string
}>()

const { copy, copied, isSupported } = useClipboard()

const defaultCopyIcon = "heroicons:document-duplicate"

const copyIcon = ref(defaultCopyIcon)

async function copyToClipboard(s: string) {
  await copy(s)
  if (copied) {
    copyIcon.value = "heroicons:check"
    setTimeout(() => {
      copyIcon.value = defaultCopyIcon
    }, 1000)
  }
}
</script>

<template>
  <IconButton
      v-if="isSupported"
      :icon="copyIcon"
      text
      @click="copyToClipboard(props.string)"
  ></IconButton>
</template>