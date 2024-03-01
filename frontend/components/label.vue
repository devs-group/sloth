<script setup lang="ts">
    interface Props {
        required?: boolean
        label: string
    }

    const props = defineProps<Props>()

</script>
<template>
    <label>{{ props.label }}<span class="text-prime-danger" v-if="props.required">*</span></label>
</template>