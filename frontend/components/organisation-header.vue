<template>
  <div class="flex flex-row w-full justify-between items-center pl-5 pr-5">
    <div class="flex flex-col gap-2">
      <p class="text-lg">
        {{ props.organisation_name }}
      </p>
    </div>
    <div class="flex flex-col">
      <IconButton
        :label="props.button.label"
        :icon="props.button.icon"
        aria-label="add"
        @click="props.button.onClick"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps({
  props: {
    required: true,
    type: Object as PropType<{ organisation_name: string, button: { label: string, icon: string, onClick: () => void } }>,
  },
})
</script>
