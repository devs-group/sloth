<script setup lang="ts">
import { ref, type Ref } from 'vue';
import { PreDefinedServices, type ProjectSchema } from '~/schema/schema';
import { useService } from '@/composables/useService';

const services = PreDefinedServices;
const toast = useToast();
const confirm = useConfirm();
const p: Ref<ProjectSchema | undefined> = ref({
  name: '',
  services: [],
  docker_credentials: [],
});

const { addService } = useService(p);

// Function to trigger confirmation dialog
function confirmAddService() {
  confirm.require({
    header: 'Add New Service?',
    message: 'Are you sure you want to add a new service to this project?',
    accept: () => {
      addService("");
      toast.add({
        severity: 'success',
        summary: 'Service Added',
        detail: 'The new service has been successfully added to the project.',
      });
    },
    acceptLabel: 'Accept',
    rejectLabel: 'Cancel',
  });
}
</script>

<template>
  <div class="p-12 flex flex-col flex-1 overflow-hidden">
    <!-- Header and button to add new service -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl">Project Services</h1>
      <IconButton
        label="Add Service"
        icon="heroicons:plus"
        @click="confirmAddService"
        aria-label="Add Service"
      />
    </div>

    <!-- List existing services -->
    <div class="flex flex-col">
      <div v-for="(service, index) in services" :key="index">
        <p>{{ service }}</p>
        <!-- Service details and actions here -->
      </div>
    </div>
  </div>
</template>
