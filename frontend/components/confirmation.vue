<script lang="ts" setup>
const { state } =  useConfirmation()
</script>

<template>
  <UModal v-model="state.show">
    <div class="p-6">
      <div class="text-center">
        <p class="text-xl">{{ state.title }}</p>
        <p class="mt-4 text-sm text-gray-500">{{ state.description }}</p>
      </div>
      <div class="w-full flex-row items-center justify-center space-x-4 text-center mt-8">
        <UButton variant="solid" label="Confirm" @click="state.confirm"/>
        <UButton color="gray" variant="solid" label="Cancel" @click="state.show = false"/>
      </div>
    </div>
  </UModal>
</template>
