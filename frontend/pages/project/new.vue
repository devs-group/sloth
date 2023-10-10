<script setup lang="ts">
import { z } from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui/dist/runtime/types'

const schema = z.object({
  name: z.string().refine(s => !s.includes(' '), 'Spaces are not allowed'),
  services: z.array(
    z.object({
      name: z.string(),
      port: z.string().min(2, "Minimum of 2 numbers").max(6, "Max 6 numbers").regex(/^\d+$/, "Only numbers are allowed").transform(Number),
      image: z.string(),
      image_tag: z.string(),
      public: z.object({
        enabled: z.boolean(),
        host: z.string(),
        ssl: z.boolean(),
        compress: z.boolean()
      })
    })
  )
})

type Schema = z.output<typeof schema>

interface Service {
  name: string
  port: string
  image: string
  image_tag: string
  public: {
    enabled: boolean
    host: string
    ssl: boolean
    compress: boolean
  }
}

const state = reactive({
  name: "",
  services: [] as Service[]
})

const isSubmitting = ref(false)
const { showError, showSuccess } = useNotification()
const router = useRouter()
const config = useRuntimeConfig()

function submit (event: FormSubmitEvent<Schema>) {
  isSubmitting.value = true
  $fetch(`${config.public.backendHost}/v1/project`, {method: "POST", body: event.data})
    .catch((e) => {
      console.error(e)
      showError("Error", "Something went wrong")
    })
    .then(() => {
      showSuccess("Success", "Your project has been created succesfully")
      router.push("/")
    })
    .finally(() => {
      isSubmitting.value = false
    })
}

function addService() {
  state.services.push({
    name: "",
    port: "",
    image: "",
    image_tag: "",
    public: {
      enabled: false,
      host: "",
      ssl: false,
      compress: false,
    },
  })
}

function removeService(idx: number) {
  state.services.splice(idx, 1)
}
</script>

<template>
  <UForm
    :schema="schema"
    :state="state"
    @submit="submit"
    class="p-12"
  >
    <div class="flex flex-row items-end space-x-12">
      <UFormGroup label="Name" name="name">
        <UInput v-model="state.name" class="w-72"/>
      </UFormGroup>
      <UButton type="submit" icon="i-heroicons-bolt" :disabled="!state.name || state.services.length === 0" :loading="isSubmitting">
        Create Project
      </UButton>
    </div>

    <div class="pt-12 flex flex-row items-center space-x-4">
      <p class="text-gray-400">Services</p>
      <UButton icon="i-heroicons-plus" :ui="{ rounded: 'rounded-full' }" @click="addService" :disabled="state.services.length === 10"/>
    </div>
  
    <div class="pt-6 grid grid-cols-5 gap-12">
      <div v-for="(s, idx) in state.services" class="space-y-4 py-3">
        <UFormGroup label="Name" :name="`services[${idx}].name`">
          <UInput v-model="s.name" type="text" />
        </UFormGroup>
        <UFormGroup label="Port" :name="`services[${idx}].port`">
          <UInput v-model="s.port" type="text" />
        </UFormGroup>
        <UFormGroup label="Image" :name="`services[${idx}].image`">
          <UInput v-model="s.image" type="text" />
        </UFormGroup>
        <UFormGroup label="Image tag" :name="`services[${idx}].image_tag`">
          <UInput v-model="s.image_tag" type="text" />
        </UFormGroup>
        <UFormGroup>
          <div class="flex flex-row justify-between items-center">
            <p class="text-sm">Publicly exposed</p>
            <UToggle v-model="s.public.enabled" />
          </div>
        </UFormGroup>
        <div v-if="s.public.enabled" class="space-y-4">
          <UFormGroup label="Host" :name="`services[${idx}].public.host`">
            <UInput v-model="s.public.host" type="text" />
          </UFormGroup>
          <UFormGroup>
            <div class="flex flex-row justify-between items-center">
              <p class="text-sm">SSL</p>
              <UToggle v-model="s.public.ssl" />
            </div>
          </UFormGroup>
          <UFormGroup>
            <div class="flex flex-row justify-between items-center">
              <p class="text-sm">Compress</p>
              <UToggle v-model="s.public.compress" />
            </div>
          </UFormGroup>
        </div>
        <div>
          <p class="text-xs text-red-400 cursor-pointer p-2 text-center" @click="removeService(idx)">Remove</p>
        </div>
      </div>
    </div>
  </UForm>
</template>
