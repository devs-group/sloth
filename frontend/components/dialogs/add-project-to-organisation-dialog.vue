<template>
  <form
    class="flex flex-col gap-4 w-full h-full"
    @submit.prevent="onCreate"
  >
    <div class="flex flex-col gap-2">
      <Dropdown
        v-model="p.upn"
        :options="projects"
        option-label="upn"
        option-value="upn"
        placeholder="Select Project*"
        :invalid="!!formErrors?.fieldErrors.upn"
        aria-describedby="select-project-help"
        :loading="isLoading"
      >
        <template #value="slotProps">
          <div
            v-if="slotProps.value"
            class="flex align-items-center"
          >
            <div>{{ projects?.find((project) => project.upn === slotProps.value)?.name }} ({{ slotProps.value }})</div>
          </div>
          <span v-else>
            {{ slotProps.placeholder }}
          </span>
        </template>
        <template #option="slotProps">
          <div class="flex align-items-center">
            <div>{{ slotProps.option.name }} ({{ slotProps.option.upn }})</div>
          </div>
        </template>
      </Dropdown>
      <small
        v-if="formErrors?.fieldErrors.upn"
        id="select-project-help"
        class="text-red-400"
      >{{ formErrors?.fieldErrors.upn.join() }}</small>
    </div>
    <div class="flex justify-end gap-2">
      <Button
        :loading="isSubmitting"
        label="Add"
        type="submit"
        @click="onCreate"
      />
      <Button
        :loading="isSubmitting"
        label="Cancel"
        severity="secondary"
        @click="onCancel"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import type { typeToFlattenedError } from 'zod'
import type { AddProjectToOrganisationType, OrganisationProject, Project } from '~/schema/schema'
import { addProjectToOrganisation } from '~/schema/schema'
import { Routes } from '~/config/routes'
import { Constants } from '~/config/const'
import type {
  IAddProjectToOrganisation,
  IAddProjectToOrganisationDialog,
  IAddProjectToOrganisationResponse,
  IDialogInjectRef,
} from '~/config/interfaces'

const dialogRef = inject<IDialogInjectRef<IAddProjectToOrganisationDialog, unknown>>('dialogRef')

const config = useRuntimeConfig()
const toast = useToast()

const isSubmitting = ref(false)
const projects = ref<Project[]>()
const formErrors = ref<typeToFlattenedError<AddProjectToOrganisationType>>()
const organisation_id: number = dialogRef?.value.data.organisation_id ?? 0
const organisationProjects: OrganisationProject[] = dialogRef?.value.data.organisationProjects ?? []

const { isLoading, loadProjects } = useProjects()

const p = ref<IAddProjectToOrganisation>({
  upn: '',
  organisation_id: organisation_id,
})

onMounted(() => {
  loadProjects().then(async (fetchedProjects) => {
    projects.value = fetchedProjects?.filter(fetchedProject => !organisationProjects.find(organisationProject => organisationProject.upn === fetchedProject.upn)) ?? []
  }).catch(error =>
    console.error('Failed to fetch projects', error))
})

const onCreate = async () => {
  const parsed = addProjectToOrganisation.safeParse(p.value)
  if (!parsed.success) {
    formErrors.value = parsed.error.formErrors
    return
  }
  isSubmitting.value = true
  $fetch<IAddProjectToOrganisationResponse>(`${config.public.backendHost}/v1/organisation/project`, {
    method: 'PUT',
    body: parsed.data,
    credentials: 'include',
  })
    .then(async (data) => {
      dialogRef?.value.close()
      toast.add({
        severity: 'success',
        summary: 'Success',
        detail: `Project has been successfully added`,
        life: Constants.ToasterDefaultLifeTime,
      })
      await navigateTo({ name: Routes.ORGANISATION, params: { id: data.id } })
    })
    .catch(() => {
      isSubmitting.value = false
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: 'Something went wrong',
        life: Constants.ToasterDefaultLifeTime,
      })
    })
}
const onCancel = () => {
  dialogRef?.value.close()
}
</script>
