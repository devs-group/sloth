<script setup>
const { state } =  useNotification()  

const callback = () => {
    state.value.show = false
}
</script>
<template>
    <UNotification v-if="state.show" :icon="state.icon" :color="state.color" :title="state.title" :description="state.message" :id="state.id" :callback="callback"/>
</template>