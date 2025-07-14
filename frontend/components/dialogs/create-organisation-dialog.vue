<template>
  <form
    class="flex flex-col gap-4 w-full h-full"
    @submit.prevent="onCreate"
  >
    <div class="flex flex-col gap-2">
      <InputText
        v-model.trim="p.organisationName"
        autofocus
        placeholder="Organisation name*"
        :invalid="!!formErrors?.fieldErrors.organisationName"
        aria-describedby="username-help"
      />
      <small
        v-if="formErrors?.fieldErrors.organisationName"
        id="username-help"
        class="text-red-400"
      >{{ formErrors?.fieldErrors.organisationName?.join() }}</small>
    </div>
    <div class="flex justify-end gap-2">
      <Button
        :loading="isSavingOrganisation"
        label="Create"
        type="submit"
        @click="onCreate"
      />
      <Button
        :disabled="isSavingOrganisation"
        label="Cancel"
        severity="secondary"
        @click="onCancel"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import type { typeToFlattenedError } from 'zod'
import { type CreateOrganisationType, type Organisation, organisationNameSchema } from '~/schema/schema'
import type { IDialogInjectRef } from '~/config/interfaces'
import { APIService } from '~/api'

const dialogRef = inject<IDialogInjectRef<{ organisation: Organisation } | undefined, Organisation | null>>('dialogRef')

const formErrors = ref<typeToFlattenedError<CreateOrganisationType>>()
const p = ref<CreateOrganisationType>({
  organisationName: dialogRef?.value.data?.organisation.organisationName ?? '',
})

const {
  execute: createOrganisation,
  isLoading: isSavingOrganisation,
  data: organisation,
} = useApi((name: string) => APIService.POST_organisation(name), {
  showSuccessToast: true,
  successMessage: 'Successfully created an organisation.',
})

const onCreate = async () => {
  const parsed = organisationNameSchema.safeParse(p.value)
  if (!parsed.success) {
    formErrors.value = parsed.error.formErrors
    return
  }
  await createOrganisation(p.value.organisationName)
  dialogRef?.value.close(organisation.value)
}

const onCancel = () => {
  dialogRef?.value.close()
}
</script>
