<template>
  <form
    @submit.prevent="onCreateService"
    class="flex flex-col gap-4 w-full h-full"
  >
    <div class="flex flex-col gap-2">
      <Dropdown
        v-model="serviceForm"
        :options="predefinedServices"
        placeholder="Select Service*"
        aria-describedby="select-service-help"
      />
    </div>
    <div class="flex justify-end gap-2">
      <Button label="Add" type="submit" />
      <Button @click="onCancel" label="Cancel" severity="secondary" />
    </div>
  </form>
</template>
<script lang="ts" setup>
import type { IDialogInjectRef } from "~/config/interfaces";

const predefinedServices = ["Empty Service", "Postgres"];
const dialogRef = inject<IDialogInjectRef<any>>("dialogRef");

const serviceForm = ref<string>(predefinedServices[0]);

const onCreateService = () => {
  dialogRef?.value.close(serviceForm.value);
};

const onCancel = () => {
  dialogRef?.value.close();
};
</script>
