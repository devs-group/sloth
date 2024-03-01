<script lang="ts" setup>
import type {PropType} from "vue";
import type {ServiceSchema} from "~/schema/schema";

const props = defineProps({
  services: {
    required: true,
    type: Object as PropType<ServiceSchema[]>,
  },
})

defineEmits<{
  (event: 'addService'): void,
  (event: 'addEnv', serviceIndex: number): void,
  (event: 'removeEnv', envIndex: number, serviceIndex: number): void
  (event: 'removeService', serviceIndex: number): void
  (event: 'addVolume', serviceIndex: number): void,
  (event: 'removeVolume', volumeIndex: number, serviceIndex: number): void
  (event: 'addPort', serviceIndex: number): void,
  (event: 'removePort', portIndex: number, serviceIndex: number): void
  (event: 'addHost', hostIndex: number): void,
  (event: 'removeHost', hostIndex: number, serviceIndex: number): void
}>()
</script>

<template>
  <div class="flex flex-col flex-1">
    <div class="flex flex-row items-center space-x-4 py-6">
      <p class="text-gray-400">Services</p>
      <IconButton icon="heroicons:plus" @click="$emit('addService')" />
    </div>
    <div class="flex gap-6 overflow-auto flex-1">
      <div v-for="service, sIdx in services" class="flex flex-col gap-6 w-[16em]">
        <div class="flex flex-col gap-1">
          <label>Name</label>
          <InputText />
        </div>
        <div class="flex flex-col gap-1">
          <label>Ports</label>
          <div class="flex flex-col gap-2">
            <InputGroup v-for="port, pIdx in service.ports">
              <InputText v-model="service.ports[pIdx]"/>
              <IconButton 
                v-if="pIdx === service.ports.length -1"
                icon="heroicons:plus"
                severity="secondary"
                outlined
                @click="$emit('addPort', sIdx)"
              />
              <IconButton 
                v-else
                icon="heroicons:minus"
                severity="secondary"
                outlined
                @click="$emit('removePort', pIdx, sIdx)"
              />
            </InputGroup>
          </div>
        </div>
        <div class="flex flex-col gap-1">
          <label>Command</label>
          <p class="text-xs">Command will be executed on container start</p>
          <InputText />
        </div>
        <div class="flex flex-col gap-1">
          <label>Image</label>
          <p class="text-xs">Valid docker image</p>
          <InputText />
        </div>
        <div class="flex flex-col gap-1">
          <label>Image tag</label>
          <p class="text-xs">Valid docker image version tag</p>
          <InputText />
        </div>
        <div class="flex gap-4">
          <p>Publicly exposed</p>
          <InputSwitch v-model="service.public.enabled"/>
        </div>
        <template v-if="service.public.enabled">
          <div class="flex flex-col gap-1">
            <p>Hosts</p>
            <p class="text-xs">For custom domains DNS A-Record is required</p>
          </div>
          <div class="flex flex-col gap-1">
            <InputGroup v-for="host, hIdx in service.public.hosts">
              <InputText v-model="service.public.hosts[hIdx]"/>
              <IconButton 
                v-if="hIdx === service.public.hosts.length -1"
                icon="heroicons:plus"
                severity="secondary"
                outlined
                @click="$emit('addHost', sIdx)"
              />
              <IconButton 
                v-else
                icon="heroicons:minus"
                severity="secondary"
                outlined
                @click="$emit('removeHost', hIdx, sIdx)"
              />
            </InputGroup>
          </div>
        </template>
        <div class="flex flex-col gap-1">
          <label>Volumes</label>
          <div class="flex flex-col gap-2">
            <InputGroup v-for="volumes, vIdx in service.volumes">
              <InputText v-model="service.volumes[vIdx]"/>
              <IconButton 
                v-if="vIdx === service.volumes.length -1"
                icon="heroicons:plus"
                severity="secondary"
                outlined
                @click="$emit('addVolume', sIdx)"
              />
              <IconButton 
                v-else
                icon="heroicons:minus"
                severity="secondary"
                outlined
                @click="$emit('removeVolume', vIdx, sIdx)"
              />
            </InputGroup>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>