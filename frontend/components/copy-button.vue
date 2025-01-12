<template>
  <div class="flex items-center">
    <IconButton
      v-if="isSupported"
      :key="getCopyIcon"
      :icon="getCopyIcon"
      text
      @click="copyToClipboard(props.string)"
    />
    <p
      v-if="copied"
      class="text-xs"
    >
      Copied ...
    </p>
  </div>
</template>

<script lang="ts" setup>
import { useClipboard } from '@vueuse/core'

const props = defineProps<{
  string: string
}>()

const { copy, copied, isSupported } = useClipboard()

const copyIcon = 'heroicons:document-duplicate'
const copiedIcon = 'heroicons:check-circle'

async function copyToClipboard(s: string) {
  await copy(s)
}

const getCopyIcon = computed(() => copied.value ? copiedIcon : copyIcon)
</script>
