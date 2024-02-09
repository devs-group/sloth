<script setup lang="ts">
const isLoading = useState("global.isLoading")
</script>

<template>
    <div v-if="!isLoading">
        <Toast position="bottom-right"/>
        <ConfirmDialog />
        <TopBar></TopBar>
        <div class="flex flex-row h-full">
            <Navigation></Navigation>
            <slot></slot>
        </div>
    </div>
</template>