<template>
<div>
    <div v-if="service?.usn && serviceState">
    <div>
        <p class="pb-2">{{ service!.name }}</p>
        <p class="text-xs text-prime-secondary-text">State: {{ serviceState.state }}</p>
        <p class="text-xs text-prime-secondary-text">Status: {{ serviceState.status }}</p>
    </div>
    <Button label="Show logs" @click="fetchAndShowLogs"/>
    <Dialog v-model:visible="props.isLogsModalOpen" :header="service.name  + ' Logs'" modal>
        <div class="overflow-auto h-[80vh]">
        <code class="text-xs" v-for="line in logsLines" :key="line">
            <p>{{ line }}</p>
        </code>
        </div>
    </Dialog>
    <OverlayProgressSpinner :show="isFetchingLogs"/>
    </div>
  </div>
</template>
  
<script setup lang="ts">
import { ref, defineProps, defineEmits, type PropType } from 'vue';
import type { IServiceState } from '~/config/interfaces';
import type { Service } from '~/schema/schema';
  
const props = defineProps({
  service: Object as PropType<Service>,
  serviceState: Object as PropType<IServiceState>,
  logsLines: Array as PropType<string[]>,
  isLogsModalOpen: Object as PropType<boolean>,
});

const isFetchingLogs = ref(false);

const emit = defineEmits(['fetchAndShowLogs']);

function fetchAndShowLogs() {
  emit('fetchAndShowLogs', props.service!.usn);
}

</script>
  