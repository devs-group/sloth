<template>
  <form @submit.prevent="onCreate" class="flex flex-col gap-4 w-full h-full">
    <div class="flex flex-col gap-2">
      <InputText
        autofocus
        v-model.trim="p.name"
        placeholder="Project name*"
        :invalid="!!formErrors?.fieldErrors.name"
        aria-describedby="username-help"
      />
      <small
        v-if="formErrors?.fieldErrors.name"
        id="username-help"
        class="text-red-400"
        >{{ formErrors?.fieldErrors.name?.join() }}</small
      >
    </div>
    <div class="flex justify-end gap-2">
      <Button
        @click="onCreate"
        :loading="isLoading"
        label="Create"
        type="submit"
      />
      <Button
        @click="onCancel"
        :disabled="isLoading"
        label="Cancel"
        severity="secondary"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import { type CreateProject } from "~/schema/schema";
import { Routes } from "~/config/routes";
import type { typeToFlattenedError } from "zod";
import type { IDialogInjectRef } from "~/config/interfaces";
import { APIService } from "~/api";

const dialogRef = inject<IDialogInjectRef<any>>("dialogRef");

const formErrors = ref<typeToFlattenedError<any>>();
const p = ref<CreateProject>({
  name: "",
});

const {
  data: project,
  execute: createProject,
  isLoading,
} = useApi(() => APIService.POST_project(p.value.name), {
  showSuccessToast: true,
  successMessage: `Project "${p.value.name}" has been created successfully`,
});

async function onCreate() {
  await createProject();
  dialogRef?.value.close();
  await navigateTo({ name: Routes.PROJECT, params: { id: project.value?.id } });
}

async function onCancel() {
  dialogRef?.value.close();
}
</script>
