<template>
  <div class="flex flex-col flex-1">
    <div class="flex flex-row items-center gap-4 py-6">
      <p class="text-prime-secondary-text">Services</p>
      <IconButton icon="heroicons:plus" @click="$emit('addService')" outlined />
    </div>
    <div class="flex gap-12 overflow-auto flex-1">
      <div v-for="(service, sIdx) in props.services" class="flex flex-col gap-6 max-w-[14em]">
        <div class="flex flex-col gap-1">
          <Label label="Name" required />
          <InputText v-model="service.name" @blur="validate(sIdx, 'name')" />
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Ports" />
          <div class="flex flex-col gap-2">
            <template v-for="(port, pIdx) in service.ports">
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
          <Label label="Image" required />
          <p class="text-xs text-prime-secondary-text">Valid docker image</p>
          <InputText v-model="service.image" />
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Image tag" required />
          <p class="text-xs text-prime-secondary-text">
            Valid docker image version tag
          </p>
          <InputText v-model="service.image_tag" />
        </div>
        <div class="flex justify-between">
          <Label label="Publicly exposed" />
          <InputSwitch v-model="service.public.enabled" />
        </div>
        <template v-if="service.public.enabled">
          <div class="flex flex-col gap-1">
            <Label label="Hosts" />
            <p class="text-xs text-prime-secondary-text">
              For custom domains DNS A-Record is required
            </p>
            <p class="text-xs text-prime-secondary-text py-2">
              Leave empty to auto generate
            </p>
            <div class="flex flex-col gap-2">
              <InputGroup v-for="(host, hIdx) in service.public.hosts">
                <InputText v-model="service.public.hosts[hIdx]" />
                <IconButton
                  v-if="hIdx === service.public.hosts.length - 1"
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
            <Label label="Port" required />
            <Dropdown
              :options="service.ports.filter((port) => port)"
              v-model="service.public.port"
            />
          </div>
          <div class="flex flex-col gap-4">
            <div class="flex justify-between">
              <p>SSL</p>
              <div
                v-tooltip.bottom="'Currently only SSL endpoints are supported'"
              >
                <InputSwitch v-model="service.public.ssl" disabled />
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
            <div v-for="(env, eIdx) in service.env_vars" class="flex flex-col">
              <InputGroup>
                <InputText
                  placeholder="Key"
                  v-model="env[0]"
                  @blur="validate(sIdx, 'env_vars', eIdx, 0)"
                />
                <InputText
                  placeholder="Value"
                  v-model="env[1]"
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
                  @click="() => $emit('removeEnv', eIdx, sIdx)"
                />
              </InputGroup>
              <small class="text-prime-danger">
                {{
                  getError(sIdx, "env_vars", eIdx, 0)?.message ||
                  getError(sIdx, "env_vars", eIdx, 1)?.message
                }}
              </small>
            </div>
          </div>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Healthcheck variables" />
          <div class="flex flex-col gap-2">
            <div v-for="(healthcheckValue, healthcheckKey) in service.healthcheck" class="flex flex-col">
              <InputGroup>
                <InputText
                  placeholder="Key"
                  :value="healthcheckKey"
                  disabled
                />
                <InputText
                  v-model="healthcheckValue as string"
                  :placeholder="setHealthCheckPlaceholders(healthcheckKey)"
                />
              </InputGroup>
            </div>
          </div>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Wait for" />
          <div class="flex flex-col gap-2">
            <div class="flex flex-col">
              <MultiSelect
              v-model="selectedValues[sIdx]"
              :options="filterServices(service)"
              optionLabel="name"
              optionValue="value"
              placeholder="Select Services"
              class="w-full md:w-20rem"
              @update:modelValue="handleChange(sIdx, $event)"
              display="chip"
              />
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
          >
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { serviceSchema } from "~/schema/schema";
import { z } from "zod";
import type { ServiceSchema } from "~/schema/schema";

const props = defineProps<{
  services: ServiceSchema[];
}>();

const { validate, getError } = useValidation(
  z.array(serviceSchema),
  props.services
);

const selectedValues = ref<Record<string, { condition: string }>[]>([]);

defineEmits<{
  (event: "addService"): void;
  (event: "addEnv", serviceIndex: number): void;
  (event: "removeEnv", envIndex: number, serviceIndex: number): void;
  (event: "removeService", serviceIndex: number): void;
  (event: "addVolume", serviceIndex: number): void;
  (event: "removeVolume", volumeIndex: number, serviceIndex: number): void;
  (event: "addPort", serviceIndex: number): void;
  (event: "removePort", portIndex: number, serviceIndex: number): void;
  (event: "addHost", hostIndex: number): void;
  (event: "removeHost", hostIndex: number, serviceIndex: number): void;
}>();

const filterServices = (currentService: ServiceSchema) => {
  return props.services
  .filter((service: ServiceSchema) => {
    if (!service.usn) return false
    if (service.usn === currentService.usn) return false
    if (!service.depends_on) {
      return true
    } else {
      return deepFilterForServices(service.depends_on, currentService.usn)
    } 
  })
  .map((service) => {
    return {
      name: service.name,
      value: service.usn
    }
  })
}

const deepFilterForServices = (depandsOn: Record<string, { condition: string }>, currentUsn: string | undefined) => {
  let show = true;
  const usns = Object.keys(depandsOn);
  for (const usn of usns) {
    if (currentUsn === usn) {
      show = false;
      break;
    }
    const nextService = props.services.find((s) => s.usn === usn);
    if (nextService && nextService.depends_on) {
      if (!deepFilterForServices(nextService.depends_on, currentUsn)) {
        show = false;
        break;
      }
    }
  }
  return show
}

const setHealthCheckPlaceholders = (key: string) => {
  switch (key) {
    case 'test':
      return 'CMD-SHELL,curl -f http://localhost/ || exit 1';
    case 'interval':
      return '30s';
    case 'timeout':
      return '10s';
    case 'retries':
      return '3'
    case 'start_period':
      return '15s';
    default:
      return ''
  }

}


const handleChange = (serivceIdx: number, value: string[]) => {
  props.services[serivceIdx].depends_on =  value.reduce((acc, v) => {
    return { ...acc, [v]: { condition: "service_healthy" } }
  }, {})
};
</script>