<template>
  <form @submit.prevent="onCreate" class="flex flex-col gap-4 w-full h-full">
    <div class="flex flex-col gap-2">
      <InputText autofocus v-model.trim="p.name" placeholder="Project name*" :invalid="!!formErrors?.fieldErrors.name" aria-describedby="username-help"/>
      <small v-if="formErrors?.fieldErrors.name" id="username-help" class="text-red-400">{{formErrors?.fieldErrors.name?.join()}}</small>
    </div>
    <div class="flex justify-end gap-2">
      <Button @click="onCreate" :loading="isSubmitting" label="Create" type="submit"/>
      <Button @click="onCancel" :loading="isSubmitting" label="Cancel" severity="secondary"/>
    </div>
  </form>
</template>

<script setup lang="ts">
import {type CreateProjectSchema, createProjectSchema} from "~/schema/schema";
import {Routes} from "~/config/routes";
import {type typeToFlattenedError, ZodError} from "zod";
import {Constants} from "~/config/const";

const dialogRef = inject<IDialogInjectRef<any>>('dialogRef');

const config = useRuntimeConfig()
const toast = useToast()

const isSubmitting = ref(false)
const formErrors = ref<typeToFlattenedError<any>>()
const p = ref<CreateProjectSchema>({
  name: "",
});

const onCreate = async () => {
  const parsed = createProjectSchema.safeParse(p.value);
  if (!parsed.success) {
    formErrors.value = parsed.error.formErrors
    return
  }
  isSubmitting.value = true;
  $fetch<ICreateProjectResponse>(`${config.public.backendHost}/v1/project`, {
    method: "POST",
    body: parsed.data,
    credentials: "include",
  })
      .then(async (data) => {
        dialogRef?.value.close()
        toast.add({
          severity: "success",
          summary: "Success",
          detail: `Project "${parsed.data.name}" has been created successfully`,
          life: Constants.ToasterDefaultLifeTime,
        });
        await navigateTo({name: Routes.PROJECT, params: {id: data.id}})
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
