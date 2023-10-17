<script setup lang="ts">
import { z } from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui/dist/runtime/types'

const serviceSchema = z.object({
  name: z.string(),
      port: z.string().min(2, "Minimum of 2 numbers").max(6, "Max 6 numbers").regex(/^\d+$/, "Only numbers are allowed"),
      image: z.string(),
      image_tag: z.string(),
      public: z.object({
        enabled: z.boolean(),
        host: z.string(),
        ssl: z.boolean(),
        compress: z.boolean()
      }),
      env_vars: z.array(
        z.tuple([
          z.string().refine(s => !s.includes(' '), 'Spaces are not allowed'),
          z.string().refine(s => !s.includes(' '), 'Spaces are not allowed')
        ])).transform(Object.fromEntries)
})

const projectSchema = z.object({
  name: z.string().refine(s => !s.includes(' '), 'Spaces are not allowed'),
  services: z.array(serviceSchema)
})

type ProjectSchema = z.output<typeof projectSchema>
type ServiceSchema = z.output<typeof serviceSchema>

const state = reactive({
  name: "",
  services: [] as ServiceSchema[]
})

const isSubmitting = ref(false)
const { showError, showSuccess } = useNotification()
const router = useRouter()
const config = useRuntimeConfig()

function submit (event: FormSubmitEvent<ProjectSchema>) {
  const data = projectSchema.parse(event.data)

  isSubmitting.value = true
  $fetch(`${config.public.backendHost}/v1/project`, { method: "POST", body: data, credentials: "include" })
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
    env_vars: [
      ["",""]
    ]
  })
}

function addEnv(serviceIdx: number) {
  state.services[serviceIdx].env_vars.push(["",""])
}

function removeEnv(serviceIdx: number, envIdx: number) {
  state.services[serviceIdx].env_vars.splice(envIdx, 1)
}

function removeService(idx: number) {
  state.services.splice(idx, 1)
}

</script>

<template>
  <UForm
    :schema="projectSchema"
    :state="state"
    @submit="submit"
    class="p-12"
  >
    <div class="flex flex-row items-end space-x-6">
      <UFormGroup label="Name" name="name">
        <UInput v-model="state.name" class="w-full md:w-72"/>
      </UFormGroup>
      <UButton type="submit" icon="i-heroicons-bolt" :disabled="!state.name || state.services.length === 0" :loading="isSubmitting">
        Create Project
      </UButton>
    </div>

    <div class="pt-12 flex flex-row items-center space-x-4">
      <p class="text-gray-400">Services</p>
      <UButton icon="i-heroicons-plus" :ui="{ rounded: 'rounded-full' }" @click="addService" :disabled="state.services.length === 10"/>
    </div>
  
    <div class="pt-6 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-12">
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
        <UFormGroup label="Environment variables" class="pt-4">
          <div class="flex flex-col space-y-2">
            <div v-for="env, envIdx in s.env_vars" class="flex space-x-2">
              <UInput placeholder="Key" v-model="env[0]"></UInput>
              <UInput placeholder="Value" v-model="env[1]"></UInput>
                <UButton 
                  v-if="envIdx===s.env_vars.length-1"
                  icon="i-heroicons-plus"
                  variant="ghost"
                  :ui="{ rounded: 'rounded-full' }"
                  @click="() => addEnv(idx)"
                  :disabled="env[0] === '' || env[1] === ''"
                />
                <UButton 
                  v-else
                  icon="i-heroicons-minus"
                  variant="ghost"
                  color="red"
                  :ui="{ rounded: 'rounded-full' }"
                  @click="() => removeEnv(idx, envIdx)"
                />
            </div>
          </div>
        </UFormGroup>
        <div>
          <p class="text-xs text-red-400 cursor-pointer p-2 text-center" @click="removeService(idx)">Remove</p>
        </div>
      </div>
    </div>
  </UForm>
</template>
