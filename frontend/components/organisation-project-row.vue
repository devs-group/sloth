<template>
  <div class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700">
    <div class="flex flex-col gap-1">
      <p class="break-all">{{ project?.name }}</p>
      <p class="text-xs text-prime-secondary-text">UPN: {{ project?.upn }}</p>
    </div>
    <div class="flex items-center gap-2">
        <IconButton     
            :loading="isDeleting"
            text
            severity="danger"
            icon="heroicons:trash"
            @click="promptProjectRemoval(project)"
        />
        <NuxtLink :to="{name: Routes.PROJECT, params: {id: project?.id}}">
            <IconButton icon="heroicons:arrow-right-on-rectangle" />
        </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, ref } from 'vue';
import { Routes } from "~/config/routes";
import { type OrganisationProject } from "~/schema/schema";

defineProps({
  project: {
    type: Object as PropType<OrganisationProject>,
    required: true,
  }
});

const emit = defineEmits(['removeProjectFromOrganisation']);

const isDeleting = ref(false);
const confirm = useConfirm();

function promptProjectRemoval(organisationProject: OrganisationProject) {
  confirm.require({
    header: 'Remove Project?',
    message: `Are you sure you want to remove the project '${organisationProject.name}'? This action cannot be undone.`,
    accept: () => {
      isDeleting.value = true;
      emit('removeProjectFromOrganisation', organisationProject.upn);
      isDeleting.value = false;
    },
    acceptLabel: 'Remove',
    rejectLabel: 'Cancel',
  });
}
</script>
