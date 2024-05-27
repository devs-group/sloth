<template>
    <form @submit.prevent="onInvite" class="flex flex-col gap-4 w-full h-full">
        <div class="flex flex-col gap-2">
            <InputText autofocus v-model="invitationForm.eMail" placeholder="User E-Mail*" :invalid="!!formErrors?.fieldErrors.eMail" aria-describedby="email-help"></InputText>
            <small v-if="formErrors?.fieldErrors.eMail" id="email-help" class="text-red-400">{{ formErrors?.fieldErrors.eMail?.join() }}</small>
        </div>
        <div class="flex justify-end gap-2">
            <Button @click="onInvite" :loading="isSubmitting" label="Invite" type="submit" />
            <Button @click="onCancel" :disabled="isSubmitting" label="Cancel" severity="secondary" />
        </div>
    </form>
</template>
<script setup lang="ts">
import type { IInviteToOrganisation, IDialogInjectRef, IInviteToOrganisationResponse } from '~/config/interfaces';
import type { typeToFlattenedError} from "zod";
import { inviteToOrganisationSchema } from '~/schema/schema';
import { Constants } from '~/config/const';
import { Routes } from '~/config/routes';

const dialogRef = inject<IDialogInjectRef<any>>('dialogRef');

const config = useRuntimeConfig()
const toast = useToast()

const isSubmitting = ref(false)
const formErrors = ref<typeToFlattenedError<any>>()

const organisation_id: number = dialogRef?.value.data.organisation_id ?? 0

const invitationForm = ref<IInviteToOrganisation>({
    eMail: "",
    organisation_id: organisation_id
})

// TODO: Invite endpoint and maybe move fetch to a service
const onInvite = async () => {
    isSubmitting.value = true;
    const parsed = inviteToOrganisationSchema.safeParse(invitationForm.value);
    if (!parsed.success) {
        isSubmitting.value = false;
        formErrors.value = parsed.error.formErrors;
        return
    }
    $fetch<IInviteToOrganisationResponse>(`${config.public.backendHost}/v1/organisation/invite`, {
        method: 'POST',
        body: parsed.data,
        credentials: "include",
    }).then(async (response) => {
        dialogRef?.value.close();
        toast.add({
            summary: 'Success',
            severity: 'success',
            detail: `${invitationForm.value.eMail} has been invited to join the organisation`,
            life: Constants.ToasterDefaultLifeTime
        })
        await navigateTo({ name: Routes.ORGANISATION, params: { id: response.id } })
    }).catch((e) => {
        isSubmitting.value = false,
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Something went wrong',
            life: Constants.ToasterDefaultLifeTime
        })
    })
}

const onCancel = () => {
    dialogRef?.value.close()
}
</script>