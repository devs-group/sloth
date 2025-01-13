<template>
  <WrappersListPage
    title="Projects"
    description="Your personal projects"
  >
    <template #actions>
      <IconButton
        icon="heroicons:plus"
        label="Add project"
        @click="openCreateProjectDialog"
      />
    </template>
    <template #content>
      <ProjectRow
        v-for="project of projects"
        :key="project.id"
        :project="project"
        @on-delete="onDeleteProject"
      />
      <OverlayProgressSpinner :show="isLoading" />
    </template>
  </WrappersListPage>
</template>

<script lang="ts" setup>
import CreateProjectDialog from '~/components/dialogs/create-project-dialog.vue'
import ProjectRow from '~/components/rows/project-row.vue'
import { DialogProps } from '~/config/const'
import { APIService } from '~/api'

const dialog = useDialog()

const {
  data: projects,
  isLoading,
  execute: getProjects,
} = useApi(() => APIService.GET_allProjects(), {
  errorMessage: 'Failed to load projects',
})

onMounted(async () => {
  await getProjects()
})

function openCreateProjectDialog() {
  dialog.open(CreateProjectDialog, {
    props: {
      header: 'Create New Project',
      ...DialogProps.BigDialog,
    },
  })
}

function onDeleteProject(id: number) {
  projects.value = projects.value?.filter(p => p.id !== id) || null
}
</script>
