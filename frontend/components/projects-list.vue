<script lang="ts" setup>
const columns = [
    {
        key: 'id',
        label: 'ID'
    },
    {
        key: 'unique_project_name',
        label: 'Project name'
    },
    {
        key: 'docker_image',
        label: 'Docker image'
    },
    {
        key: 'docker_image_tag',
        label: 'Docker image tag'
    },
]
const people = [
    {
        "id": 1,
        "unique_project_name": "asdas",
        "docker_image": "asdas",
        "docker_image_tag": "asdas"
    }
]
</script>

<template>
    <div class="w-full">
        <div class="p-6">
            <h1 class="text-2xl">Projects</h1>
            <p>2 Projects in your organisation</p>
        </div>

        <div>
            <div class="p-6 flex flex-row flex-1 items-center border border-1 border-x-0 border-gray-200 dark:border-gray-700">
                <UAvatar alt="Project Name" size="sm" class="mr-3"/>
                <p>Project-name-1</p>
            </div>
        </div>
    </div>
</template>