<template>
  <div class="w-full max-w-6xl">
    <div class="flex justify-between items-center p-6">
      <div>
        <h1 class="text-2xl">Projects</h1>
      </div>
      <IconButton @click="onCreateProject" icon="heroicons:plus" label="Add New Project"/>
    </div>

    <div class="relative">
      <ProjectRow v-for="project of projects" @on-delete="loadProjects" :project="project"/>
      <OverlayProgressSpinner :show="isLoading"/>
    </div>
  </div>
</template>

<script lang="ts" setup>
import type {Project} from "~/schema/schema";
import CreateProjectDialog from "~/components/dialogs/create-project-dialog.vue";
import ProjectRow from "~/components/rows/project-row.vue";
import {Constants, DialogProps} from "~/config/const";

const dialog = useDialog();
const { isLoading, loadProjects } = useProjects()
const projects = ref<Project[]>()

onMounted(() => {
  loadProjects().then(async (fetchedProjects) => {
    projects.value = fetchedProjects ?? []
  }).catch((error) => 
    console.error("Failed to fetch projects", error))
});

const onCreateProject = () => {
  dialog.open(CreateProjectDialog, {
    props: {
      header: 'Create New Project',
      ...DialogProps.BigDialog,
    },
  })
}
</script>
