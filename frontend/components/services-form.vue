<script lang="ts" setup>
import {serviceSchema } from "~/schema/schema" 
import {type ServiceSchema} from "~/schema/schema";

const props = defineProps<{
  services: ServiceSchema[]
}>()

const errors = {}

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
    <div class="flex flex-row items-center gap-4 py-6">
        <p class="text-prime-secondary-text">Services</p>
        <IconButton icon="heroicons:plus" @click="$emit('addService')" outlined/>
    </div>
    <div class="flex gap-12 overflow-auto flex-1">
      <div v-for="service, sIdx in props.services" class="flex flex-col gap-6 max-w-[14em]">
        <div class="flex flex-col gap-1">
          <Label label="Name" required/>
          <InputText v-model="service.name"/>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Ports"/>
          <div class="flex flex-col gap-2">
            <InputGroup v-for="port, pIdx in service.ports">
              <InputText v-model="service.ports[pIdx]"/>
              <IconButton 
                v-if="pIdx === service.ports.length -1"
                :disabled="port===''"
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
          </div>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Command" />
          <p class="text-xs text-prime-secondary-text">Command will be executed on container start</p>
          <InputText v-model="service.command"/>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Image" required/>
          <p class="text-xs text-prime-secondary-text">Valid docker image</p>
          <InputText v-model="service.image"/>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Image tag" required/>
          <p class="text-xs text-prime-secondary-text">Valid docker image version tag</p>
          <InputText v-model="service.image_tag"/>
        </div>
        <div class="flex justify-between">
          <Label label="Publicly exposed"/>
          <InputSwitch v-model="service.public.enabled"/>
        </div>
        <template v-if="service.public.enabled">
          <div class="flex flex-col gap-1">
            <Label label="Hosts"/>
            <p class="text-xs text-prime-secondary-text">
              For custom domains DNS A-Record is required
            </p>
            <p class="text-xs text-prime-secondary-text py-2">Leave empty to auto generate</p>
            <div class="flex flex-col gap-2">
              <InputGroup v-for="host, hIdx in service.public.hosts">
                <InputText v-model="service.public.hosts[hIdx]"/>
                <IconButton 
                  v-if="hIdx === service.public.hosts.length -1"
                  :disabled="host===''"
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
            <Label label="Port" required/>
            <Dropdown :options="service.ports.filter(port => port)" v-model="service.public.port" />
          </div>
          <div class="flex flex-col gap-4">
            <div class="flex justify-between">
              <p>SSL</p>
              <div v-tooltip.bottom="'Currently only SSL endpoints are supported'">
                <InputSwitch v-model="service.public.ssl" disabled/>
              </div>
            </div>
            <div class="flex justify-between">
              <p>Compress</p>
              <InputSwitch v-model="service.public.compress"/>
            </div>
          </div>
        </template>
        <div class="flex flex-col gap-1">
          <Label label="Volumes" />
          <div class="flex flex-col gap-2">
            <InputGroup v-for="volume, vIdx in service.volumes">
              <InputText v-model="service.volumes[vIdx]"/>
              <IconButton
                v-if="vIdx === service.volumes.length -1"
                :disabled="volume===''"
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
          </div>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Environment variables"/>
          <div class="flex flex-col gap-2">
            <div v-for="env, eIdx in service.env_vars" class="flex gap-2">
              <InputGroup>
                <InputText placeholder="Key" v-model="env[0]"/>
                <InputText placeholder="Value" v-model="env[1]"/>
                <IconButton v-if="eIdx === service.env_vars.length -1"
                  icon="heroicons:plus"
                  outlined severity="secondary"
                  :disabled="env[0] === '' || env[1] === ''"
                  class="text-prime-primary"
                  @click="() => $emit('addEnv', sIdx)"
                />
                <IconButton v-else 
                  icon="heroicons:minus"
                  outlined severity="secondary"
                  class="text-prime-danger"
                  @click="() => $emit('removeEnv', eIdx, sIdx)"
                />
              </InputGroup>
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