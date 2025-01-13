<script setup lang="ts">
interface Props {
  required?: boolean
  label: string
}

const props = defineProps<Props>()
</script>

<template>
  <label>{{ props.label }}<span
    v-if="props.required"
    class="text-prime-danger"
  >*</span></label>
</template>
