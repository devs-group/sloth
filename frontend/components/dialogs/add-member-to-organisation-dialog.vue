<template>
  <form
    class="flex flex-col gap-4 w-full h-full"
    @submit.prevent="onCreate"
  >
    <div class="flex flex-col gap-2">
      <InputText
        v-model.trim="p.email"
        autofocus
        placeholder="Member email*"
        :invalid="!!formErrors?.fieldErrors.email"
        aria-describedby="member-email-help"
      />
      <small
        v-if="formErrors?.fieldErrors.email"
        id="member-email-help"
        class="text-red-400"
      >{{ formErrors.fieldErrors.email.join() }}</small>
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
import { putMemberToOrganisation, type UpdateOrganisationType } from '~/schema/schema'
import { Routes } from '~/config/routes'
import { Constants } from '~/config/const'
import type {
  IAddMemberToOrganisationDialog,
  IDialogInjectRef,
  IPutMemberToOrganisation,
  IPutMemberToOrganisationResponse,
} from '~/config/interfaces'

const dialogRef = inject<IDialogInjectRef<IAddMemberToOrganisationDialog, unknown>>('dialogRef')

const config = useRuntimeConfig()
const toast = useToast()

const isSubmitting = ref(false)
const formErrors = ref<typeToFlattenedError<UpdateOrganisationType>>()
const organisation_id: number = dialogRef?.value.data.organisation_id ?? 0

const p = ref<IPutMemberToOrganisation>({
  email: '',
  organisation_id: organisation_id,
})

const onCreate = async () => {
  const parsed = putMemberToOrganisation.safeParse(p.value)
  if (!parsed.success) {
    formErrors.value = parsed.error.formErrors
    return
  }
  isSubmitting.value = true
  $fetch<IPutMemberToOrganisationResponse>(`${config.public.backendHost}/v1/organisation/member`, {
    method: 'PUT',
    body: parsed.data,
    credentials: 'include',
  })
    .then(async (data) => {
      dialogRef?.value.close()
      toast.add({
        severity: 'success',
        summary: 'Success',
        detail: `Member has been successfully added`,
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
