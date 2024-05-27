<template>
  <form @submit.prevent="onInvite" class="flex flex-col gap-4 w-full h-full">
    <div class="flex flex-col gap-2">
      <InputText
        autofocus
        v-model="invitationForm.eMail"
        placeholder="User E-Mail*"
        :invalid="!!formErrors?.fieldErrors.eMail"
        aria-describedby="email-help"
      ></InputText>
      <small
        v-if="formErrors?.fieldErrors.eMail"
        id="email-help"
        class="text-red-400"
        >{{ formErrors?.fieldErrors.eMail?.join() }}</small
      >
    </div>
    <div class="flex justify-end gap-2">
      <Button
        @click="onInvite"
        :loading="isSubmitting"
        label="Invite"
        type="submit"
      />
      <Button
        @click="onCancel"
        :disabled="isSubmitting"
        label="Cancel"
        severity="secondary"
      />
    </div>
  </form>
</template>
<script setup lang="ts">
import type {
  IInviteToOrganisation,
  IDialogInjectRef,
  IInviteToOrganisationResponse,
} from "~/config/interfaces";
import type { typeToFlattenedError } from "zod";
import { inviteToOrganisationSchema } from "~/schema/schema";
import { Constants } from "~/config/const";
import { Routes } from "~/config/routes";

const dialogRef = inject<IDialogInjectRef<any>>("dialogRef");

const config = useRuntimeConfig();
const toast = useToast();

const isSubmitting = ref(false);
const formErrors = ref<typeToFlattenedError<any>>();

const organisation_name: string = dialogRef?.value.data.organisation_name ?? "";

const invitationForm = ref<IInviteToOrganisation>({
  eMail: "",
  organisation_name: organisation_name,
});

// TODO: Invite endpoint and maybe move fetch to a service
const onInvite = async () => {
  isSubmitting.value = true;
  const parsed = inviteToOrganisationSchema.safeParse(invitationForm.value);
  if (!parsed.success) {
    isSubmitting.value = false;
    formErrors.value = parsed.error.formErrors;
    return;
  }

  $fetch<IInviteToOrganisationResponse>(
    `${config.public.backendHost}/v1/organisation/member`,
    {
      method: "PUT",
      body: parsed.data,
      credentials: "include",
    }
  )
    .then(async (response) => {
      dialogRef?.value.close();
      toast.add({
        severity: "success",
        summary: "Invitation Sent",
        detail: "Invitation has been sent successfully",
        life: Constants.ToasterDefaultLifeTime,
      });
      await navigateTo({
        name: Routes.ORGANISATION,
        params: { id: response.id },
      });
    })
    .catch((e) => {
      isSubmitting.value = false;
      console.error("unable to invite", e);
      toast.add({
        severity: "error",
        summary: "Invitation Failed",
        detail: "Unable to send invitation",
        life: Constants.ToasterDefaultLifeTime,
      });
    });
};

const onCancel = () => {
  dialogRef?.value.close();
};
</script>
