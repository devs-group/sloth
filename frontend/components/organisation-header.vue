<template>
  <div class="flex flex-row w-full justify-between items-center pl-5 pr-5">
    <div class="flex flex-col gap-2">
      <p class="text-lg">
        {{ organisationName }}
      </p>
    </div>
    <div class="flex flex-col">
      <IconButton
        :label="buttonLabel"
        :icon="buttonIcon"
        aria-label="add"
        @click="emits('onClick')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps({
  organisationName: {
    type: String,
    required: true,
  },
  buttonLabel: {
    type: String,
    required: true,
  },
  buttonIcon: {
    type: String,
    required: true,
  },
})

const emits = defineEmits(['onClick'])
</script>
