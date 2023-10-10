<script lang="ts" setup>

</script>
<template>
    <div class="w-full h-12 border border-gray-200 dark:border-gray-700 flex flex-row items-center justify-end px-5">
        <UButton icon="i-heroicons-pencil-square" size="sm" color="gray" variant="solid" :trailing="false">
            <NuxtLink to="/project/new">New Project</NuxtLink>
        </UButton>
    </div>
</template>