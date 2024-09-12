<template>
<div>
    <div v-if="service?.usn && serviceState">
    <div>
        <p class="pb-2">{{ service!.name }}</p>
        <p class="text-xs text-prime-secondary-text">State: {{ serviceState.state }}</p>
        <p class="text-xs text-prime-secondary-text">Status: {{ serviceState.status }}</p>
    </div>
    <div class="flex flex-col items-start gap-2">
      <Button label="Show logs" @click="fetchAndShowLogs"/>
      <Button label="Open shell" @click="emit('startShell')"/>
    </div>
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
  
const props = defineProps<{
  service: Service,
  serviceState: IServiceState,
  logsLines: string[],
  isLogsModalOpen: boolean,
  dialogHeaderName: string
}>();

const isLogsModalOpen = ref(false);
const isFetchingLogs = ref(false);

const emit = defineEmits(['fetchAndShowLogs', 'startShell', 'closeLogsModal']);

function fetchAndShowLogs() {
  emit('fetchAndShowLogs', props.service!.usn, props.service!.name);

}

function closeLogModals() {
  emit('closeLogsModal');
  
}
</script>
  