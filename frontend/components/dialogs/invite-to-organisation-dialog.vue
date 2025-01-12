<template>
  <form
    class="flex flex-col gap-4 w-full h-full"
    @submit.prevent="onInvite"
  >
    <div class="flex flex-col gap-2">
      <InputText
        v-model="invitationForm.email"
        autofocus
        placeholder="User E-Mail*"
        :invalid="!!formErrors?.fieldErrors.email"
        aria-describedby="email-help"
      />
      <small
        v-if="formErrors?.fieldErrors.email"
        id="email-help"
        class="text-red-400"
      >{{ formErrors.fieldErrors.email.join() }}</small>
    </div>
    <div class="flex justify-end gap-2">
      <Button
        :loading="isSubmitting"
        label="Invite"
        type="submit"
        @click="onInvite"
      />
      <Button
        :disabled="isSubmitting"
        label="Cancel"
        severity="secondary"
        @click="onCancel"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import type { typeToFlattenedError } from 'zod'
import type {
  IDialogInjectRef,
  IInviteToOrganisation,
  IInviteToOrganisationDialog,
  IInviteToOrganisationResponse,
} from '~/config/interfaces'
import { inviteToOrganisationSchema, type InviteToOrganisationType } from '~/schema/schema'
import { Constants } from '~/config/const'
import { Routes } from '~/config/routes'

const dialogRef = inject<IDialogInjectRef<IInviteToOrganisationDialog, unknown>>('dialogRef')

const config = useRuntimeConfig()
const toast = useToast()

const isSubmitting = ref(false)
const formErrors = ref<typeToFlattenedError<InviteToOrganisationType>>()

const organisation_id: number = dialogRef?.value.data.organisation_id ?? 0

const invitationForm = ref<IInviteToOrganisation>({
  email: '',
  organisation_id: organisation_id,
})

// TODO: Invite endpoint and maybe move fetch to a service
const onInvite = async () => {
  isSubmitting.value = true
  const parsed = inviteToOrganisationSchema.safeParse(invitationForm.value)
  if (!parsed.success) {
    isSubmitting.value = false
    formErrors.value = parsed.error.formErrors
    return
  }

  $fetch<IInviteToOrganisationResponse>(
    `${config.public.backendHost}/v1/organisation/member`,
    {
      method: 'PUT',
      body: parsed.data,
      credentials: 'include',
    },
  )
    .then(async (response) => {
      dialogRef?.value.close()
      toast.add({
        severity: 'success',
        summary: 'Invitation Sent',
        detail: 'Invitation has been sent successfully',
        life: Constants.ToasterDefaultLifeTime,
      })
      await navigateTo({
        name: Routes.ORGANISATION,
        params: { id: response.id },
      })
    })
    .catch((e) => {
      isSubmitting.value = false
      console.error('unable to invite', e)
      toast.add({
        severity: 'error',
        summary: 'Invitation Failed',
        detail: 'Unable to send invitation',
        life: Constants.ToasterDefaultLifeTime,
      })
    })
}

const onCancel = () => {
  dialogRef?.value.close()
}
</script>
