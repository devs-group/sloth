<template>
  <div class="relative flex flex-col gap-4 w-full p-6">
    <NuxtLink
      class="lg:hidden"
      :to="{ name: Routes.PROJECTS }"
    >
      <IconButton icon="heroicons:arrow-uturn-left" />
    </NuxtLink>

    <template v-if="project">
      <ProjectInfo
        :project="project"
        :is-updating-loading="isUpdatingProject"
        :is-updating-and-restarting-loading="isUpdatingProject"
        @update-and-restart-project="() => updateAndRestartProject(project!)"
      />

      <form @submit.prevent>
        <Menubar
          :model="tabItems"
          @change="onChangeTab"
        />
        <component
          :is="activeTabComponent"
          :credentials="project.docker_credentials"
          :project="project"
          :submitted="submitted"
          @add-credential="addCredential"
          @remove-credential="removeCredential"
          @add-service="addService"
          @add-env="addEnv"
          @remove-env="removeEnv"
          @add-volume="addVolume"
          @remove-volume="removeVolume"
          @remove-service="removeService"
          @add-port="addPort"
          @remove-port="removePort"
          @add-host="addHost"
          @remove-host="removeHost"
          @remove-post-deploy-action="removePostDeployAction"
          @add-post-deploy-action="addPostDeployAction"
        />
      </form>
    </template>
    <Message
      v-else-if="pageErrorMessage"
      severity="error"
      :closable="false"
    >
      {{ pageErrorMessage }},
      <NuxtLink
        class="underline"
        :to="{ name: Routes.PROJECTS }"
      >go back</NuxtLink>
    </Message>
    <OverlayProgressSpinner :show="isLoadingProject" />
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import type { TabItem } from '~/config/interfaces'
import { type Project, projectSchema } from '~/schema/schema'
import { Routes } from '~/config/routes'
import ServicesForm from '~/components/services-form.vue'
import DockerCredentialsForm from '~/components/docker-credentials-form.vue'
import ProjectInfo from '~/components/project-info.vue'
import { APIService } from '~/api'

const route = useRoute()
const projectID = parseInt(route.params.id.toString())

const pageErrorMessage = ref('')
const submitted = ref(false)
const toast = useToast()

const tabItems = computed(
  () =>
    [
      {
        label: 'Services',
        component: ServicesForm,
        command: () => onChangeTab(0),
      },
      {
        label: 'Docker Credentials',
        component: DockerCredentialsForm,
        command: () => onChangeTab(1),
      },
      { label: 'Monitoring', disabled: true },
    ] as TabItem[],
)

const { activeTabComponent, onChangeTab } = useTabs(tabItems)

const {
  data: project,
  isLoading: isLoadingProject,
  execute: getProject,
} = useApi((id: number) => APIService.GET_projectByID(id))

const {
  data: updatedProject,
  isLoading: isUpdatingProject,
  execute: updateProject,
  error: updateProjectError,
} = useApi((p: Project) => APIService.PUT_updateProject(p), {
  showSuccessToast: true,
  successMessage: 'Project has been updated succesfully',
})

const {
  addCredential,
  removeCredential,
  addEnv,
  removeEnv,
  addHost,
  removeHost,
  addPort,
  removePort,
  addService,
  removeService,
  addVolume,
  removeVolume,
  removePostDeployAction,
  addPostDeployAction,
} = useService(project)

onMounted(async () => {
  await getProject(projectID)
})

const updateAndRestartProject = async (p: Project) => {
  const parsed = projectSchema.safeParse(p)
  if (!parsed.success) {
    submitted.value = true

    console.error('error:', parsed)
    let errMsg = 'Some errors appeard in the following forms:\n'

    Object.keys(parsed.error.formErrors.fieldErrors).forEach((key) => {
      errMsg = errMsg.concat(`${key}\n`)
    })

    toast.add({
      severity: 'error',
      summary: 'Unable to save the project',
      detail: errMsg,
    })
    return
  }
  await updateProject(p)
  if (!updateProjectError.value) {
    project.value = updatedProject.value
  }
}
</script>
