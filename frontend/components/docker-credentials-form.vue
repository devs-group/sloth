<script lang="ts" setup>
/* __placeholder__ */
import type { DockerCredentialSchema } from "~/schema/schema";

const props = defineProps({
  credentials: {
    required: true,
    type: Object as PropType<DockerCredentialSchema[]>
  }
})

defineEmits<{
  (event: 'addCredential'): void,
  (event: 'removeCredential', index: number): void
}>()

const config = useRuntimeConfig()
</script>

<template>
  <div>
    <div class="flex flex-row items-center space-x-4 py-6">
      <p class="text-gray-400">Docker registry credentials</p>
      <UButton icon="i-heroicons-plus" :ui="{ rounded: 'rounded-full' }" @click="$emit('addCredential')" />
    </div>

    <div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div v-for="(c, idx) in props.credentials">
            <UFormGroup label="Username" :name="`docker_credentials.${idx}.username`">
              <UInput v-model="c.username" />
            </UFormGroup>

            <UFormGroup label="Password" :name="`docker_credentials.${idx}.password`">
              <UInput v-model="c.password" type="password" />
            </UFormGroup>

            <UFormGroup label="Registry" :name="`docker_credentials.${idx}.registry`">
              <UInput v-model="c.registry" />
            </UFormGroup>
            <div>
              <p
                  class="text-xs text-red-400 cursor-pointer p-2 text-center"
                  @click="$emit('removeCredential', idx as number)">
                Remove credential
              </p>
            </div>
          </div>
        </div>
    </div>
  </div>
</template>