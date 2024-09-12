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
                :loading="isSubmitting"
                label="Create"
                type="submit"
            />
            <Button
                @click="onCancel"
                :loading="isSubmitting"
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
import { Constants } from "~/config/const";
import type { typeToFlattenedError } from "zod";
import type {
    ICreateOrganisationResponse,
    IDialogInjectRef,
} from "~/config/interfaces";

const dialogRef = inject<IDialogInjectRef<any>>("dialogRef");

const config = useRuntimeConfig();
const toast = useToast();

const isSubmitting = ref(false);
const formErrors = ref<typeToFlattenedError<any>>();
const p = ref<CreateOrganisation>({
    organisation_name: "",
});

const onCreate = async () => {
    const parsed = createOrganisationSchema.safeParse(p.value);
    if (!parsed.success) {
        formErrors.value = parsed.error.formErrors;
        return;
    }
    isSubmitting.value = true;
    $fetch<ICreateOrganisationResponse>(
        `${config.public.backendHost}/v1/organisation`,
        {
            method: "POST",
            body: parsed.data,
            credentials: "include",
        },
    )
        .then(async (data) => {
            dialogRef?.value.close(data);
            toast.add({
                severity: "success",
                summary: "Success",
                detail: `Organisation "${parsed.data.organisation_name}" has been created successfully`,
                life: Constants.ToasterDefaultLifeTime,
            });
        })
        .catch(() => {
            isSubmitting.value = false;
            toast.add({
                severity: "error",
                summary: "Error",
                detail: "Something went wrong",
                life: Constants.ToasterDefaultLifeTime,
            });
        });
};
const onCancel = () => {
    dialogRef?.value.close();
};
</script>
