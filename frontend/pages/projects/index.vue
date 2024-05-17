<template>
  <div class="w-full max-w-6xl">
    <div class="flex justify-between items-center p-6">
      <div>
        <h1 class="text-2xl">Projects</h1>
      </div>
      <IconButton @click="onCreateProject" icon="heroicons:plus" label="Add New Project"/>
    </div>

    <div class="relative">
      <ProjectRow v-for="project of projects" @on-delete="fetchProjects" :project="project"/>
      <OverlayProgressSpinner :show="isLoading"/>
    </div>
  </div>
</template>

<script lang="ts" setup>
import type {Project} from "~/schema/schema";
import CreateProjectDialog from "~/components/dialogs/create-project-dialog.vue";
import ProjectRow from "~/components/rows/project-row.vue";
import {Constants, DialogProps} from "~/config/const";

const config = useRuntimeConfig()
const projects = ref<Project[]>()
const toast = useToast()
const dialog = useDialog();

const isLoading = ref(true)

fetchProjects()

function fetchProjects() {
  isLoading.value = true
  $fetch<Project[]>(`${config.public.backendHost}/v1/projects`, {credentials: "include"})
      .then(payload => {
        projects.value = payload
      })
      .catch(() => {
        toast.add({
          severity: "error",
          summary: "Error",
          detail: "There was an error loading your projects",
          life: Constants.ToasterDefaultLifeTime,
        })
      })
      .finally(() => {
        isLoading.value = false
      })
}

const onCreateProject = () => {
  dialog.open(CreateProjectDialog, {
    props: {
      header: 'Create New Project',
      ...DialogProps.BigDialog,
    },
  })
}
</script>
