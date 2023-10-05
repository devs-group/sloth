<script lang="ts" setup>
const isNewProjectModalOpen = ref(false)
</script>
<template>
    <div class="w-full h-12 border border-gray-200 dark:border-gray-700 flex flex-row items-center justify-end px-5">
        <UButton icon="i-heroicons-pencil-square" size="sm" color="gray" variant="solid" label="New project" :trailing="false" @click="isNewProjectModalOpen = true"/>
        <ModalCreateProject :isOpen="isNewProjectModalOpen"></ModalCreateProject>
    </div>
</template>