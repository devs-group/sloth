<template>
    <form
        @submit.prevent="onCreateService"
        class="flex flex-col gap-4 w-full h-full"
    >
        <div class="flex flex-col gap-2">
            <Dropdown
                v-model="serviceForm"
                :options="Object.keys(serviceTemplates)"
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
import type { ServiceSchema } from "~/schema/schema";
import { EmptyServiceTemplate } from "~/service-templates/empty-service-template";
import { MinioServiceTemplate } from "~/service-templates/minio-service-template";
import { PostgreServiceTemplate } from "~/service-templates/postgre-service-template";

const dialogRef = inject<IDialogInjectRef<any>>("dialogRef");

const serviceTemplates: Record<string, ServiceSchema> = {
    "Empty Service": EmptyServiceTemplate,
    Postgres: PostgreServiceTemplate,
    "Minio S3": MinioServiceTemplate,
};

const serviceForm = ref<string>("Empty Service");

const onCreateService = () => {
    dialogRef?.value.close(serviceTemplates[serviceForm.value]);
};

const onCancel = () => {
    dialogRef?.value.close();
};
</script>
