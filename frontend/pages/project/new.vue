<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui/dist/runtime/types'
import {projectSchema, ProjectSchema, Service, ServiceSchema} from "~/schema/schema";

const state = reactive({
  name: "",
  services: [] as ServiceSchema[]
})

const isSubmitting = ref(false)
const { showError, showSuccess } = useNotification()
const router = useRouter()
const config = useRuntimeConfig()

async function submit (event: FormSubmitEvent<ProjectSchema>) {
  const data = projectSchema.parse(event.data)
  isSubmitting.value = true
  try {
    await $fetch(`${config.public.backendHost}/v1/project`, { method: "POST", body: data, credentials: "include" })
    showSuccess("Success", "Your project has been created succesfully")
    await router.push("/")
  } catch (e) {
    console.error(e)
    showError("Error", "Something went wrong")
  } finally {
    isSubmitting.value = false
  }
}

function addService() {
  state.services.push({
    name: "",
    ports: [""],
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
    ],
    volumes: [""],
  })
}

function addEnv(serviceIdx: number) {
  state.services[serviceIdx].env_vars.push(["",""])
}

function removeEnv(envIdx: number, serviceIdx: number) {
  state.services[serviceIdx].env_vars.splice(envIdx, 1)
}

function addVolume(serviceIdx: number) {
  state.services[serviceIdx].volumes.push("")
}

function removeVolume(volIdx: number, serviceIdx: number) {
  state.services[serviceIdx].volumes.splice(volIdx, 1)
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
      <ServiceForm
          v-for="(s, idx) in state.services"
          :service="s as Service"
          :index="idx"
          @add-env="addEnv"
          @remove-env="removeEnv"
          @add-volume="addVolume"
          @remove-volume="removeVolume"
          @remove-service="removeService"
      ></ServiceForm>
    </div>
  </UForm>
</template>
