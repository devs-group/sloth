<template>
  <form
    class="flex flex-col gap-4 w-full h-full"
    @submit.prevent="onCreateService"
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
      <Button
        label="Add"
        type="submit"
      />
      <Button
        label="Cancel"
        severity="secondary"
        @click="onCancel"
      />
    </div>
  </form>
</template>

<script lang="ts" setup>
import type { IDialogInjectRef } from '~/config/interfaces'
import type { ServiceSchema } from '~/schema/schema'
import { EmptyServiceTemplate } from '~/service-templates/empty-service-template'
import { PostgreServiceTemplate } from '~/service-templates/postgre-service-template'
import { SeaweedfsServiceTemplate } from '~/service-templates/seaweedfs-service-template'

const dialogRef = inject<IDialogInjectRef<unknown, ServiceSchema>>('dialogRef')

const serviceTemplates: Record<string, ServiceSchema> = {
  'Custom Service': EmptyServiceTemplate,
  'Postgres': PostgreServiceTemplate,
  'SeaweedFS S3': SeaweedfsServiceTemplate,
}

const serviceForm = ref<string>('Custom Service')

const onCreateService = () => {
  dialogRef?.value.close(serviceTemplates[serviceForm.value])
}

const onCancel = () => {
  dialogRef?.value.close()
}
</script>
