<script setup lang="ts">
const props = defineProps({
    isOpen: {
        type: Boolean,
        required: true,
    }
})
</script>

<template>
  <div>
    <UModal v-model="props.isOpen">
        <template #header>
          <p>Create a new project</p>
        </template>
        form
    </UModal>
  </div>
</template>