<template>
  <form @submit.prevent="onCreate" class="flex flex-col gap-4 w-full h-full">
    <div class="flex flex-col gap-2">
      <InputText
        autofocus
        v-model.trim="p.organisation_name"
        placeholder="Organisation name*"
        :invalid="!!formErrors?.fieldErrors.organisation_name"
        aria-describedby="username-help"
      />
      <small
        v-if="formErrors?.fieldErrors.organisation_name"
        id="username-help"
        class="text-red-400"
        >{{ formErrors?.fieldErrors.organisation_name?.join() }}</small
      >
    </div>
    <div class="flex justify-end gap-2">
      <Button
        @click="onCreate"
        :loading="isSavingOrganisastion"
        label="Create"
        type="submit"
      />
      <Button
        @click="onCancel"
        :disabled="isSavingOrganisastion"
        label="Cancel"
        severity="secondary"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import {
  createOrganisationSchema,
  type CreateOrganisation,
} from "~/schema/schema";
import type { typeToFlattenedError } from "zod";
import type { IDialogInjectRef } from "~/config/interfaces";
import { APIService } from "~/api";

const dialogRef = inject<IDialogInjectRef<any>>("dialogRef");

const formErrors = ref<typeToFlattenedError<any>>();
const p = ref<CreateOrganisation>({
  organisation_name: "",
});

const {
  execute: createOrganisation,
  isLoading: isSavingOrganisastion,
  data: organisation,
} = useApi((name: string) => APIService.POST_organisation(name), {
  showSuccessToast: true,
  successMessage: "Succesfully created an organisation.",
});

const onCreate = async () => {
  const parsed = createOrganisationSchema.safeParse(p.value);
  if (!parsed.success) {
    formErrors.value = parsed.error.formErrors;
    return;
  }
  await createOrganisation(p.value.organisation_name);
  dialogRef?.value.close(organisation.value);
};

const onCancel = () => {
  dialogRef?.value.close();
};
</script>
