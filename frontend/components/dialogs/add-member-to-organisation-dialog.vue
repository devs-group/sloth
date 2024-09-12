<template>
    <form @submit.prevent="onCreate" class="flex flex-col gap-4 w-full h-full">
      <div class="flex flex-col gap-2">
        <InputText autofocus v-model.trim="p.email" placeholder="Member email*" :invalid="!!formErrors?.fieldErrors.email" aria-describedby="member-email-help"/>
        <small v-if="formErrors?.fieldErrors.email" id="member-email-help" class="text-red-400">{{formErrors?.fieldErrors.email?.join()}}</small>
      </div>
      <div class="flex justify-end gap-2">
        <Button @click="onCreate" :loading="isSubmitting" label="Add" type="submit"/>
        <Button @click="onCancel" :loading="isSubmitting" label="Cancel" severity="secondary"/>
      </div>
    </form> 
</template>

<script setup lang="ts">
import {putMemberToOrganisation} from "~/schema/schema";
import {Routes} from "~/config/routes";
import { Constants } from "~/config/const";
import type { typeToFlattenedError} from "zod";
import type { IDialogInjectRef, IPutMemberToOrganisation, IPutMemberToOrganisationResponse } from "~/config/interfaces";

const dialogRef = inject<IDialogInjectRef<any>>('dialogRef');

const config = useRuntimeConfig()
const toast = useToast()

const isSubmitting = ref(false)
const formErrors = ref<typeToFlattenedError<any>>()
const organisation_id: number = dialogRef?.value.data.organisation_id ?? 0

const p = ref<IPutMemberToOrganisation>({
    email: "",
    organisation_id: organisation_id,
});

const onCreate = async () => {
  const parsed = putMemberToOrganisation.safeParse(p.value);
  if (!parsed.success) {
    formErrors.value = parsed.error.formErrors
    return
  }
  isSubmitting.value = true;
  $fetch<IPutMemberToOrganisationResponse>(`${config.public.backendHost}/v1/organisation/member`, {
    method: "PUT",
    body: parsed.data,
    credentials: "include",
  })
      .then(async (data) => {
        dialogRef?.value.close()
        toast.add({
          severity: "success",
          summary: "Success",
          detail: `Member has been successfully added`,
          life: Constants.ToasterDefaultLifeTime,
        });
        await navigateTo({name: Routes.ORGANISATION, params: {id: data.id}})
      })
      .catch(() => {
        isSubmitting.value = false;
        toast.add({
          severity: "error",
          summary: "Error",
          detail: "Something went wrong",
          life: Constants.ToasterDefaultLifeTime,
        });
      })
}
const onCancel = () => {
  dialogRef?.value.close()
}
</script>
