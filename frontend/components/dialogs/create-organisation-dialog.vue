<template>
  <form
    class="flex flex-col gap-4 w-full h-full"
    @submit.prevent="onCreate"
  >
    <div class="flex flex-col gap-2">
      <InputText
        v-model.trim="p.organisation_name"
        autofocus
        placeholder="Organisation name*"
        :invalid="!!formErrors?.fieldErrors.organisation_name"
        aria-describedby="username-help"
      />
      <small
        v-if="formErrors?.fieldErrors.organisation_name"
        id="username-help"
        class="text-red-400"
      >{{ formErrors?.fieldErrors.organisation_name?.join() }}</small>
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
import { createOrganisationSchema, type CreateOrganisationType, type Organisation } from '~/schema/schema'
import type { IDialogInjectRef } from '~/config/interfaces'
import { APIService } from '~/api'

const dialogRef = inject<IDialogInjectRef<unknown, Organisation | null>>('dialogRef')

const formErrors = ref<typeToFlattenedError<CreateOrganisationType>>()
const p = ref<CreateOrganisationType>({
  organisation_name: '',
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
  const parsed = createOrganisationSchema.safeParse(p.value)
  if (!parsed.success) {
    formErrors.value = parsed.error.formErrors
    return
  }
  await createOrganisation(p.value.organisation_name)
  dialogRef?.value.close(organisation.value)
}

const onCancel = () => {
  dialogRef?.value.close()
}
</script>
