<template>
  <div class="flex-1 overflow-y-auto pt-2">
    <form
      class="grid grid-cols-1 items-start lg:grid-cols-2 gap-2"
      @submit.prevent
    >
      <FloatLabel
        variant="on"
      >
        <Select
          v-model="selectedTemplate"
          :options="serviceTemplates"
          auto-filter-focus
          option-label="name"
          option-value="template"
          class="w-full"
        />
        <label for="name">Select a template</label>
      </FloatLabel>

      <div class="col-span-full" />

      <FloatLabel
        variant="on"
      >
        <InputText
          id="name"
          v-model="formData.name"
          :invalid="!!errors.fieldErrors.name"
          class="w-full"
        />
        <label for="name">Name*</label>
      </FloatLabel>

      <FloatLabel
        variant="on"
      >
        <AutoComplete
          id="ports"
          v-model="formData.ports"
          :invalid="!!errors.fieldErrors.ports"
          multiple
          fluid
          :typeahead="false"
        />
        <label for="ports">Type port, then press Enter (Multiple)</label>
      </FloatLabel>

      <FloatLabel
        variant="on"
      >
        <InputText
          id="image"
          v-model="formData.image"
          class="w-full"
          :invalid="!!errors.fieldErrors.image"
        />
        <label for="image">Image* (eg. mysql)</label>
      </FloatLabel>

      <FloatLabel
        variant="on"
      >
        <InputText
          id="tag"
          v-model="formData.image_tag"
          :invalid="!!errors.fieldErrors.image_tag"
          class="w-full"
        />
        <label for="tag">Image Tag* (eg. latest)</label>
      </FloatLabel>

      <FloatLabel
        variant="on"
      >
        <InputText
          id="command"
          v-model="formData.command"
          :invalid="!!errors.fieldErrors.command"
          class="w-full"
        />
        <label for="command">Command</label>
      </FloatLabel>

      <div class="col-span-full h-4" />

      <div class="flex flex-col gap-2">
        <p class="text-sm text-prime-secondary-text">
          Enable Public URLs
        </p>
        <p
          v-if="!canExposePublicly"
          class="text-xs text-prime-warning"
        >
          You need at least one port to enable public URLs
        </p>
        <p
          v-else
          class="text-xs text-prime-primary"
        >
          If you want to use a custom domain, you need to add an "A-Record" on your host,
          pointing to this IP: {{ config.public.serverIp }}
        </p>
        <div
          v-for="p of formData.public"
          :key="p.port"
          class="flex flex-col gap-2"
        >
          <InputGroup class="font-mono text-sm">
            <InputGroupAddon>Port: {{ p.port }}</InputGroupAddon>
            <InputGroupAddon>
              <label
                class="flex items-center gap-2"
                :for="`enabled-${p.port}`"
              >
                <ToggleSwitch
                  v-model="p.enabled"
                  :input-id="`enabled-${p.port}`"
                />
                Enable
              </label>
            </InputGroupAddon>
            <InputText
              v-model="p.host"
              :disabled="!p.enabled"
              :pt="{ root: '!text-sm' }"
              placeholder="eg. your.domain.com (Empty = autogenerate)"
            />
            <InputGroupAddon>
              <label
                class="flex items-center gap-2"
                :for="`ssl-${p.port}`"
              >
                <ToggleSwitch
                  v-model="p.ssl"
                  :disabled="!p.enabled"
                  :input-id="`ssl-${p.port}`"
                />
                SSL
              </label>
            </InputGroupAddon>
            <InputGroupAddon>
              <label
                class="flex items-center gap-2"
                :for="`compress-${p.port}`"
              >
                <ToggleSwitch
                  v-model="p.compress"
                  :disabled="!p.enabled"
                  :input-id="`compress-${p.port}`"
                />
                Gzip
              </label>
            </InputGroupAddon>
          </InputGroup>
        </div>
      </div>

      <div class="flex flex-col gap-2">
        <label
          for="publicly-exposed"
          class="flex gap-2 cursor-pointer self-start"
        >
          <ToggleSwitch
            input-id="publicly-exposed"
          />
          Publicly Exposed?
        </label>
      </div>
    </form>
  </div>

  <div class="flex justify-end pt-4 gap-2">
    <Button
      label="Save"
      @click="onSave"
    />
    <Button
      label="Cancel"
      severity="secondary"
      outlined
      @click="onCancel"
    />
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import type { typeToFlattenedError } from 'zod'
import { cloneDeep } from 'lodash-es'
import type { IBaseDialog, IPersistentVolumeDialogData } from '~/interfaces/dialog-interfaces'
import { EmptyServiceTemplate } from '~/service-templates/empty-service-template'
import { PostgreServiceTemplate } from '~/service-templates/postgre-service-template'
import { MinioServiceTemplate } from '~/service-templates/minio-service-template'
import { type Service, serviceSchema, type ServiceSchema } from '~/schema/schema'

const dialogRef = inject<IBaseDialog<IPersistentVolumeDialogData>>('dialogRef')!
const confirm = useConfirm()
const config = useRuntimeConfig()

const selectedTemplate = ref<ServiceSchema>(EmptyServiceTemplate)
const formData = ref<Service>(cloneDeep(EmptyServiceTemplate))
const errors = ref<typeToFlattenedError<ServiceSchema>>({ formErrors: [], fieldErrors: {} })
const hasBeenSaved = ref(false)

const serviceTemplates = [
  {
    name: 'Empty Service',
    template: EmptyServiceTemplate,
  },
  {
    name: 'PostgreSQL',
    template: PostgreServiceTemplate,
  },
  {
    name: 'MinIO',
    template: MinioServiceTemplate,
  },
]

const project = ref(dialogRef.value.data.project)

watch(selectedTemplate, (value) => {
  if (!value) {
    return
  }
  formData.value = cloneDeep(value)
})

watch(formData, (value) => {
  for (const port of value.ports) {
    if (!formData.value.public.find(p => p.port === port)) {
      formData.value.public.push({
        enabled: false,
        port,
        host: '',
        ssl: true,
        compress: false,
      })
    }
  }

  if (!hasBeenSaved.value) {
    // Only validate on change when an attempt to save the form has been done
    return
  }

  const parsed = serviceSchema.safeParse(value)
  if (parsed.error) {
    errors.value = parsed.error.formErrors
    return
  }
  errors.value = { formErrors: [], fieldErrors: {} }
}, { deep: true })

const canExposePublicly = computed(() => {
  return formData.value.ports.length > 0
})

const onSave = async () => {
  hasBeenSaved.value = true
  const parsed = serviceSchema.safeParse(formData.value)
  if (parsed.error) {
    errors.value = parsed.error.formErrors
    return
  }
  errors.value = { formErrors: [], fieldErrors: {} }
}

const onCancel = () => {
  // TODO: Add check if something has changed and ensure user really wants to cancel
  confirm.require({
    header: 'Add New Service',
    message: 'Are you sure you want to cancel? All unsaved changes will be lost.',
    acceptLabel: 'Yes',
    rejectLabel: 'No',
    acceptClass: 'p-button-danger p-button-outlined',
    defaultFocus: 'reject',
    accept: async () => {
      dialogRef.value.close()
    },
  })
}
</script>
