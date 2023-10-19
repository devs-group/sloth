<script lang="ts" setup>
import {PropType} from "vue";
import {Service} from "~/schema/schema";

const props = defineProps({
  service: {
    type: Object as PropType<Service>,
    required: true,
  },
  index: {
    type: Number,
    required: true,
  }
})

defineEmits<{
  (event: 'addEnv', serviceIndex: number): void,
  (event: 'removeEnv', envIndex: number, serviceIndex: number): void
  (event: 'removeService', serviceIndex: number): void
  (event: 'addVolume', serviceIndex: number): void,
  (event: 'removeVolume', volumeIndex: number, serviceIndex: number): void
  (event: 'addPort', serviceIndex: number): void,
  (event: 'removePort', portIndex: number, serviceIndex: number): void
}>()
</script>

<template>
  <div class="space-y-4 py-3">
    <UFormGroup label="Name" :name="`services[${index}].name`">
      <UInput v-model="service.name" type="text" />
    </UFormGroup>
    <UFormGroup label="Ports" class="pt-4">
      <div class="flex flex-col space-y-2">
        <div v-for="(port, portIdx) in service.ports as string[]" class="flex space-x-2">
          <UInput class="w-full" placeholder="Port" v-model="service.ports[portIdx]"></UInput>
          <UButton
              v-if="portIdx === (service.ports as string[]).length-1"
              icon="i-heroicons-plus"
              variant="ghost"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('addPort', index)"
              :disabled="port === ''"
          />
          <UButton
              v-else
              icon="i-heroicons-minus"
              variant="ghost"
              color="red"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('removePort', portIdx, index)"
          />
        </div>
      </div>
    </UFormGroup>
    <UFormGroup label="Image" :name="`services[${index}].image`">
      <UInput v-model="service.image" type="text" />
    </UFormGroup>
    <UFormGroup label="Image tag" :name="`services[${index}].image_tag`">
      <UInput v-model="service.image_tag" type="text" />
    </UFormGroup>
    <UFormGroup>
      <div class="flex flex-row justify-between items-center">
        <p class="text-sm">Publicly exposed</p>
        <UToggle v-model="service.public.enabled" />
      </div>
    </UFormGroup>
    <div v-if="service.public.enabled" class="space-y-4">
      <UFormGroup label="Host" :name="`services[${index}].public.host`">
        <UInput v-model="service.public.host" type="text" />
      </UFormGroup>
      <UFormGroup label="Port" :name="`services[${index}].public.port`">
        <USelectMenu v-model="service.public.port" :options="service.ports" />
      </UFormGroup>
      <UFormGroup>
        <div class="flex flex-row justify-between items-center">
          <p class="text-sm">SSL</p>
          <UToggle v-model="service.public.ssl" />
        </div>
      </UFormGroup>
      <UFormGroup>
        <div class="flex flex-row justify-between items-center">
          <p class="text-sm">Compress</p>
          <UToggle v-model="service.public.compress" />
        </div>
      </UFormGroup>
    </div>
    <UFormGroup label="Volumes" class="pt-4">
      <div class="flex flex-col space-y-2">
        <div v-for="(volume, volIdx) in service.volumes as string[]" class="flex space-x-2">
          <UInput class="w-full" placeholder="Path" v-model="service.volumes[volIdx]"></UInput>
          <UButton
              v-if="volIdx === (service.volumes as string[]).length-1"
              icon="i-heroicons-plus"
              variant="ghost"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('addVolume', index)"
              :disabled="volume === ''"
          />
          <UButton
              v-else
              icon="i-heroicons-minus"
              variant="ghost"
              color="red"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('removeVolume', volIdx, index)"
          />
        </div>
      </div>
    </UFormGroup>
    <UFormGroup label="Environment variables" class="pt-4">
      <div class="flex flex-col space-y-2">
        <div v-for="(env, envIdx) in service.env_vars as string[][]" class="flex space-x-2">
          <UInput placeholder="Key" v-model="env[0]"></UInput>
          <UInput placeholder="Value" v-model="env[1]"></UInput>
          <UButton
              v-if="envIdx === (service.env_vars as string[]).length-1"
              icon="i-heroicons-plus"
              variant="ghost"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('addEnv', index)"
              :disabled="env[0] === '' || env[1] === ''"
          />
          <UButton
              v-else
              icon="i-heroicons-minus"
              variant="ghost"
              color="red"
              :ui="{ rounded: 'rounded-full' }"
              @click="() => $emit('removeEnv', envIdx, index)"
          />
        </div>
      </div>
    </UFormGroup>
    <div>
      <p class="text-xs text-red-400 cursor-pointer p-2 text-center" @click="() => $emit('removeService', index)">Remove</p>
    </div>
  </div>
</template>