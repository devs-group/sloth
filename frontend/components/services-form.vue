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
  <div class="flex flex-row items-center space-x-4 py-6">
    <p class="text-gray-400">Services</p>
    <UButton icon="i-heroicons-plus" :ui="{ rounded: 'rounded-full' }" @click="$emit('addService')" />
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-12">
    <div class="space-y-4" v-for="(s, idx) in services">
      <UFormGroup label="Name" :name="`services.${idx}.name`" required>
        <UInput v-model="s.name" type="text" required />
      </UFormGroup>
      <UFormGroup
          v-for="(port, portIdx) in s.ports as string[]"
          label="Port"
          :name="`services.${idx}.ports.${portIdx}`">
        <div  class="flex space-x-2">
          <UInput
              class="w-full"
              placeholder="Port"
              v-model="s.ports[portIdx]"
          ></UInput>
          <UButton
              v-if="portIdx === (s.ports as string[]).length-1"
              icon="i-heroicons-plus"
              variant="ghost"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('addPort', idx)"
              :disabled="port === ''"
          />
          <UButton
              v-else
              icon="i-heroicons-minus"
              variant="ghost"
              color="red"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('removePort', portIdx, idx)"
          />
        </div>
      </UFormGroup>
      <UFormGroup label="Command" :name="`services.${idx}.command`" description="Command will be executed on container start">
        <UInput v-model="s.command" type="text" />
      </UFormGroup>
      <UFormGroup label="Image" :name="`services.${idx}.image`" description="Valid docker image" required>
        <UInput v-model="s.image" type="text" required />
      </UFormGroup>
      <UFormGroup label="Image tag" :name="`services.${idx}.image_tag`" description="Valid docker image version tag" required>
        <UInput v-model="s.image_tag" type="text" required />
      </UFormGroup>
      <UFormGroup>
        <div class="flex flex-row justify-between items-center">
          <p class="text-sm">Publicly exposed</p>
          <UToggle v-model="s.public.enabled" />
        </div>
      </UFormGroup>
      <div v-if="s.public.enabled" class="space-y-4">
        
        <UFormGroup label="Hosts">
          <template #description>
            For custom domains DNS A-Record is required
            <UTooltip text="IP: 45.83.105.86">
              <UIcon name="i-heroicons-information-circle" />
            </UTooltip>
          </template>
        </UFormGroup>

        <UFormGroup
          v-for="(host, hostIdx) in s.public.hosts as string[]"
          :name="`services.${idx}.hosts.${hostIdx}`"
          :description="'Leave empty to auto-generate'"
          label="Host">
            <div class="flex space-x-2">
              <UInput class="w-full" placeholder="Host" v-model="s.public.hosts[hostIdx]"></UInput>
              <UButton
                  v-if="hostIdx === (s.public.hosts as string[]).length-1"
                  icon="i-heroicons-plus"
                  variant="ghost"
                  :ui="{ rounded: 'rounded-full' }"
                  @click="() => $emit('addHost', idx)"
                  :disabled="host === ''"
              />
              <UButton
                  v-else
                  icon="i-heroicons-minus"
                  variant="ghost"
                  color="red"
                  :ui="{ rounded: 'rounded-full' }"
                  @click="() => $emit('removeHost', hostIdx, idx)"
              />
            </div>
          </UFormGroup>

        <UFormGroup label="Port" :name="`services.${idx}.public.port`" required >
          <USelectMenu v-model="s.public.port" :options="s.ports" required />
        </UFormGroup>
        <UFormGroup>
          <div class="flex flex-row justify-between items-center">
            <p class="text-sm">SSL</p>
            <UTooltip text="Currently only SSL endpoints are supported">
              <UToggle v-model="s.public.ssl" disabled />
            </UTooltip>
          </div>
        </UFormGroup>
        <UFormGroup>
          <div class="flex flex-row justify-between items-center">
            <p class="text-sm">Compress</p>
            <UToggle v-model="s.public.compress" />
          </div>
        </UFormGroup>
      </div>
      <UFormGroup
          v-for="(volume, volIdx) in s.volumes as string[]"
          :name="`services.${idx}.volumes.${volIdx}`"
          label="Volume" description="Path within the container">
        <div class="flex space-x-2">
          <UInput class="w-full" placeholder="Path" v-model="s.volumes[volIdx]"></UInput>
          <UButton
              v-if="volIdx === (s.volumes as string[]).length-1"
              icon="i-heroicons-plus"
              variant="ghost"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('addVolume', idx)"
              :disabled="volume === ''"
          />
          <UButton
              v-else
              icon="i-heroicons-minus"
              variant="ghost"
              color="red"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('removeVolume', volIdx, idx)"
          />
        </div>
      </UFormGroup>
      <UFormGroup label="Environment variables" class="pt-4">
        <div class="flex flex-col space-y-2">
          <div v-for="(env, envIdx) in s.env_vars as string[][]" class="flex space-x-2">
            <UInput placeholder="Key" v-model="env[0]"></UInput>
            <UInput placeholder="Value" v-model="env[1]"></UInput>
            <UButton
                v-if="envIdx === (s.env_vars as string[][]).length-1"
                icon="i-heroicons-plus"
                variant="ghost"
                :ui="{ rounded: 'rounded-full' }"
                @click="() => $emit('addEnv', idx)"
                :disabled="env[0] === '' || env[1] === ''"
            />
            <UButton
                v-else
                icon="i-heroicons-minus"
                variant="ghost"
                color="red"
                :ui="{ rounded: 'rounded-full' }"
                @click="() => $emit('removeEnv', envIdx, idx)"
            />
          </div>
        </div>
      </UFormGroup>
      <div>
        <p class="text-xs text-red-400 cursor-pointer p-2 text-center" @click="() => $emit('removeService', idx)">Remove service</p>
      </div>
    </div>
  </div>
</template>