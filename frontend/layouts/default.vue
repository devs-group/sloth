<script setup lang="ts">
const isLoading = useState("global.isLoading")
</script>

<template>
    <div v-if="!isLoading">
        <Toast position="bottom-right"/>
        <Confirmation></Confirmation>
        <TopBar></TopBar>
        <div class="flex flex-row h-full">
            <Navigation></Navigation>
            <slot></slot>
        </div>
    </div>
</template>