<template>
  <form
    class="flex flex-col gap-4 w-full h-full"
    @submit.prevent="onCreate"
  >
    <div class="flex flex-col gap-2">
      <InputText
        v-model.trim="p.name"
        autofocus
        placeholder="Project name*"
        :invalid="!!formErrors?.fieldErrors.name"
        aria-describedby="username-help"
      />
      <small
        v-if="formErrors?.fieldErrors.name"
        id="username-help"
        class="text-red-400"
      >{{ formErrors.fieldErrors.name?.join() }}</small>
    </div>
    <div class="flex justify-end gap-2">
      <Button
        :loading="isLoading"
        label="Create"
        type="submit"
        @click="onCreate"
      />
      <Button
        :disabled="isLoading"
        label="Cancel"
        severity="secondary"
        @click="onCancel"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import type { typeToFlattenedError } from 'zod'
import type { CreateProjectType } from '~/schema/schema'
import { Routes } from '~/config/routes'
import type { IDialogInjectRef } from '~/config/interfaces'
import { APIService } from '~/api'

const dialogRef = inject<IDialogInjectRef<unknown, unknown>>('dialogRef')

const formErrors = ref<typeToFlattenedError<CreateProjectType>>()
const p = ref<CreateProjectType>({
  name: '',
})

const {
  data: project,
  execute: createProject,
  isLoading,
} = useApi(() => APIService.POST_project(p.value.name), {
  showSuccessToast: true,
  successMessage: `Project has been created successfully`,
})

async function onCreate() {
  await createProject()
  dialogRef?.value.close()
  await navigateTo({ name: Routes.PROJECT, params: { id: project.value?.id } })
}

async function onCancel() {
  dialogRef?.value.close()
}
</script>
