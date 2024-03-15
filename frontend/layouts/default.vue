<script setup lang="ts">
const isLoading = useState("global.isLoading")
</script>

<template>
    <div v-if="!isLoading" class="h-[100vh] w-[100-vw] flex flex-col">
        <Toast position="bottom-right"/>
        <ConfirmDialog />
        <TopBar></TopBar>
        <div class="flex flex-row flex-1">
            <Navigation></Navigation>
            <slot></slot>
        </div>
    </div>
</template>