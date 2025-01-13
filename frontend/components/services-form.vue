<template>
  <div class="flex flex-col flex-1">
    <div class="flex flex-row items-center gap-4 py-6">
      <p class="text-prime-secondary-text">
        Services
      </p>
      <IconButton
        icon="heroicons:plus"
        outlined
        @click="openAddServiceDialog()"
      />
    </div>
    <div class="flex gap-12 overflow-auto flex-1">
      <div
        v-for="(service, sIdx) in props.project.services"
        :key="service.id"
        class="flex flex-col gap-6 min-w-[24em] max-w-[24em]"
      >
        <ProgressSpinner
          v-if="isLoadingServiceStates"
        />
        <ServiceDetail
          v-else
          :project="props.project"
          :service="service"
          :service-state="serviceStates!"
        />
        <span
          v-if="service.usn"
          class="text-xs text-gray-400"
        >
          Unique service name: {{ service.usn }}</span>
        <div class="flex flex-col gap-1">
          <Label
            label="Name"
            required
          />
          <InputText
            v-model="service.name"
            @blur="validate(sIdx, 'name')"
          />
          <small class="text-prime-danger">
            {{ getError(sIdx, "name")?.message }}
          </small>
        </div>
        <div class="flex flex-col gap-1">
          <Label
            label="Ports"
            required
          />
          <div class="flex flex-col gap-2">
            <template
              v-for="(port, pIdx) in service.ports"
              :key="port"
            >
              <InputGroup>
                <InputText
                  v-model="service.ports[pIdx]"
                  @blur="validate(sIdx, 'ports', pIdx)"
                />
                <IconButton
                  v-if="pIdx === service.ports.length - 1"
                  :disabled="port === ''"
                  icon="heroicons:plus"
                  severity="secondary"
                  outlined
                  class="text-prime-primary"
                  @click="$emit('addPort', sIdx)"
                />
                <IconButton
                  v-else
                  icon="heroicons:minus"
                  severity="secondary"
                  outlined
                  class="text-prime-danger"
                  @click="$emit('removePort', pIdx, sIdx)"
                />
              </InputGroup>
              <small class="text-prime-danger">
                {{ getError(sIdx, "ports", pIdx)?.message }}
              </small>
            </template>
          </div>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Command" />
          <p class="text-xs text-prime-secondary-text">
            Command will be executed on container start
          </p>
          <InputText v-model="service.command" />
        </div>
        <div class="flex flex-col gap-1">
          <Label
            label="Image"
            required
          />
          <p class="text-xs text-prime-secondary-text">
            Valid docker image
          </p>
          <InputText
            v-model="service.image"
            @blur="validate(sIdx, 'image')"
          />
          <small class="text-prime-danger">
            {{ getError(sIdx, "image")?.message }}
          </small>
        </div>
        <div class="flex flex-col gap-1">
          <Label
            label="Image tag"
            required
          />
          <p class="text-xs text-prime-secondary-text">
            Valid docker image version tag
          </p>
          <InputText
            v-model="service.image_tag"
            @blur="validate(sIdx, 'image_tag')"
          />
          <small class="text-prime-danger">
            {{ getError(sIdx, "image_tag")?.message }}
          </small>
        </div>
        <div class="flex justify-between">
          <Label label="Publicly exposed" />
          <InputSwitch v-model="service.public.enabled" />
        </div>
        <template v-if="service.public.enabled">
          <div class="flex flex-col gap-1">
            <Label label="Hosts" />
            <p class="text-xs text-prime-secondary-text">
              For custom domains DNS A-Record to IP
              {{ config.public.serverIp }} is required
            </p>
            <p class="text-xs text-prime-secondary-text py-2">
              Leave empty to auto generate
            </p>
            <div class="flex flex-col gap-2">
              <InputGroup
                v-for="(host, hIdx) in service.public.hosts"
                :key="host"
              >
                <InputText
                  v-model="service.public.hosts[hIdx]"
                />
                <IconButton
                  v-if="
                    hIdx === service.public.hosts.length - 1
                  "
                  :disabled="host === ''"
                  icon="heroicons:plus"
                  severity="secondary"
                  outlined
                  class="text-prime-primary"
                  @click="$emit('addHost', sIdx)"
                />
                <IconButton
                  v-else
                  icon="heroicons:minus"
                  severity="secondary"
                  outlined
                  class="text-prime-danger"
                  @click="$emit('removeHost', hIdx, sIdx)"
                />
              </InputGroup>
            </div>
          </div>
          <div class="flex flex-col gap-1">
            <Label
              label="Port"
              required
            />
            <Dropdown
              v-model="service.public.port"
              :options="service.ports.filter((port) => port)"
            />
          </div>
          <div class="flex flex-col gap-4">
            <div class="flex justify-between">
              <p>SSL</p>
              <div
                v-tooltip.bottom="
                  'Currently only SSL endpoints are supported'
                "
              >
                <InputSwitch
                  v-model="service.public.ssl"
                  disabled
                />
              </div>
            </div>
            <div class="flex justify-between">
              <p>Compress</p>
              <InputSwitch v-model="service.public.compress" />
            </div>
          </div>
        </template>
        <div class="flex flex-col gap-1">
          <Label label="Volumes" />
          <div class="flex flex-col gap-2">
            <div
              v-for="(volume, vIdx) in service.volumes"
              :key="volume"
              class="flex flex-col"
            >
              <InputGroup>
                <InputText
                  v-model="service.volumes[vIdx]"
                  @blur="validate(sIdx, 'volumes', vIdx)"
                />
                <IconButton
                  v-if="vIdx === service.volumes.length - 1"
                  :disabled="volume === ''"
                  icon="heroicons:plus"
                  severity="secondary"
                  outlined
                  class="text-prime-primary"
                  @click="$emit('addVolume', sIdx)"
                />
                <IconButton
                  v-else
                  icon="heroicons:minus"
                  severity="secondary"
                  outlined
                  class="text-prime-danger"
                  @click="$emit('removeVolume', vIdx, sIdx)"
                />
              </InputGroup>
              <small class="text-prime-danger">{{
                getError(sIdx, "volumes", vIdx)?.message
              }}</small>
            </div>
          </div>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Environment variables" />
          <div class="flex flex-col gap-2">
            <div
              v-for="(env, eIdx) in service.env_vars"
              :key="`${env[0]}_${env[1]}`"
              class="flex flex-col"
            >
              <InputGroup>
                <InputText
                  v-model="env[0]"
                  placeholder="Key"
                  @blur="validate(sIdx, 'env_vars', eIdx, 0)"
                />
                <InputText
                  v-model="env[1]"
                  placeholder="Value"
                  @blur="validate(sIdx, 'env_vars', eIdx, 1)"
                />
                <IconButton
                  v-if="eIdx === service.env_vars.length - 1"
                  icon="heroicons:plus"
                  outlined
                  severity="secondary"
                  :disabled="env[0] === '' || env[1] === ''"
                  class="text-prime-primary"
                  @click="() => $emit('addEnv', sIdx)"
                />
                <IconButton
                  v-else
                  icon="heroicons:minus"
                  outlined
                  severity="secondary"
                  class="text-prime-danger"
                  @click="
                    () => $emit('removeEnv', eIdx, sIdx)
                  "
                />
              </InputGroup>
              <small class="text-prime-danger">
                {{
                  getError(sIdx, "env_vars", eIdx, 0)
                    ?.message
                    || getError(sIdx, "env_vars", eIdx, 1)?.message
                }}
              </small>
            </div>
          </div>
        </div>
        <div class="pt-6">
          <Button
            outlined
            severity="danger"
            class="w-full flex justify-center"
            label="Remove service"
            @click="$emit('removeService', sIdx)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { z } from 'zod'
import AddServiceDialog from './dialogs/add-service-dialog.vue'
import type { Project } from '~/schema/schema'
import { serviceSchema } from '~/schema/schema'
import { DialogProps } from '~/config/const'
import { APIService } from '~/api'

const dialog = useDialog()
const config = useRuntimeConfig()

const props = defineProps({
  project: {
    type: Object as PropType<Project>,
    required: true,
  },
  submitted: {
    type: Boolean,
    required: true,
  },
})

const emits = defineEmits<{
  (event: 'addService', serviceID: number): void
  (event: 'addEnv', serviceIndex: number): void
  (event: 'removeEnv', envIndex: number, serviceIndex: number): void
  (event: 'removeService', serviceIndex: number): void
  (event: 'addVolume', serviceIndex: number): void
  (event: 'removeVolume', volumeIndex: number, serviceIndex: number): void
  (event: 'addPort', serviceIndex: number): void
  (event: 'removePort', portIndex: number, serviceIndex: number): void
  (event: 'addHost', hostIndex: number): void
  (event: 'removeHost', hostIndex: number, serviceIndex: number): void
  (event: 'addPostDeployAction', postDeployActionIndex: number): void
  (
    event: 'removePostDeployAction',
    postDeployActionIndex: number,
    serviceIndex: number,
  ): void
}>()

let { validate, getError } = useValidation(
  z.array(serviceSchema),
  props.project.services,
)

const {
  data: serviceStates,
  isLoading: isLoadingServiceStates,
  execute: getServiceStates,
} = useApi((projectID: number) => APIService.GET_serviceStates(projectID))

const selectedValues = ref<string[][]>([])

onMounted(async () => {
  if (props.submitted) {
    validateInputFields()
  }
  props.project.services.forEach((service, index) => {
    selectedValues.value.push([])
    if (service.depends_on) {
      Object.keys(service.depends_on).forEach(key =>
        selectedValues.value[index].push(key),
      )
    }
  })
  await getServiceStates(props.project.id)
})

watch(() => props.project.services.length, updateValidate)
watch(() => props.submitted, validateInputFields)

function updateValidate() {
  const { validate: newValidate, getError: newGetError } = useValidation(
    z.array(serviceSchema),
    props.project.services,
  )

  validate = newValidate
  getError = newGetError
}

function validateInputFields() {
  props.project.services.forEach((service, index) => {
    Object.keys(service).forEach((key) => {
      switch (key) {
        case 'ports':
        case 'volumes':
          service[key].forEach((_v, i) => {
            validate(index, key, i)
          })
          break
        case 'env_vars':
          service[key].forEach((_v, i) => {
            validate(index, key, i, 0)
            validate(index, key, i, 1)
          })
          break
        default:
          validate(index, key)
      }
    })
  })
}

function openAddServiceDialog() {
  dialog.open(AddServiceDialog, {
    props: {
      header: 'Add Service',
      ...DialogProps.BigDialog,
    },
    data: {
      services: props.project.services,
    },
    onClose(options) {
      if (options?.data) {
        emits('addService', options.data)
      }
    },
  })
}
</script>
