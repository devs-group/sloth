<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui/dist/runtime/types'
import {projectSchema, ProjectSchema, Service, ServiceSchema} from "~/schema/schema";
import DockerCredentialsForm from "~/components/docker-credentials-form.vue";
import ServicesForm from "~/components/services-form.vue";

const tabItems = [{
  label: 'Services',
  __component: ServicesForm,
}, {
  label: 'Docker credentials',
  __component: DockerCredentialsForm,
}, {
  label: 'Monitoring (coming soon)',
  disabled: true,
}]

const isSubmitting = ref(false)
const { showError, showSuccess } = useNotification()
const router = useRouter()
const config = useRuntimeConfig()

const p = ref<ProjectSchema>({
  name: "",
  services: [],
  docker_credentials: [],
})
const activeTabComponent = ref(tabItems[0].__component)

function onChangeTab(idx: number) {
  activeTabComponent.value = tabItems[idx].__component
}

async function saveProject (event: FormSubmitEvent<ProjectSchema>) {
  const data = projectSchema.parse(event.data)
  isSubmitting.value = true
  try {
    await $fetch(`${config.public.backendHost}/v1/project`, { method: "POST", body: data, credentials: "include" })
    showSuccess("Success", "Your project has been created successfully")
    await router.push("/")
  } catch (e) {
    console.error(e)
    showError("Error", "Something went wrong")
  } finally {
    isSubmitting.value = false
  }
}

function addService() {
  p.value?.services.push({
    name: "",
    ports: [""],
    image: "",
    image_tag: "",
    public: {
      enabled: false,
      host: "",
      port: "",
      ssl: true,
      compress: false,
    },
    env_vars: [
      ["",""]
    ],
    volumes: [""],
  })
}

function addEnv(serviceIdx: number) {
  p.value?.services[serviceIdx].env_vars.push(["",""])
}

function removeEnv(envIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].env_vars.splice(envIdx, 1)
}

function addVolume(serviceIdx: number) {
  p.value?.services[serviceIdx].volumes.push("")
}

function removeVolume(volIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].volumes.splice(volIdx, 1)
}

function addPort(serviceIdx: number) {
  p.value?.services[serviceIdx].ports.push("")
}

function removePort(portIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].ports.splice(portIdx, 1)
}

function removeService(idx: number) {
  p.value?.services.splice(idx, 1)
}

function addCredential() {
  p.value?.docker_credentials.push({
    username: "",
    password: "",
    registry: "",
  })
}

function removeCredential(idx: number) {
  p.value?.docker_credentials.splice(idx, 1)
}

</script>

<template>
  <UForm
    :schema="projectSchema"
    :state="p"
    @submit="saveProject"
    class="p-12 w-full"
  >
    <div class="flex flex-row items-end space-x-6 pb-12">
      <UFormGroup label="Name" name="name" required >
        <UInput v-model="p!.name" class="w-full md:w-72" required />
      </UFormGroup>
      <UButton type="submit" icon="i-heroicons-bolt" :disabled="!p?.name || p.services.length === 0" :loading="isSubmitting">
        Create Project
      </UButton>
    </div>

    <!-- TABS -->
    <UTabs :items="tabItems" @change="onChangeTab" />
    <component
        :is="activeTabComponent as string"
        :credentials="p.docker_credentials"
        @add-credential="addCredential"
        @remove-credential="removeCredential"

        :services="p.services"
        @add-service="addService"
        @add-env="addEnv"
        @remove-env="removeEnv"
        @add-volume="addVolume"
        @remove-volume="removeVolume"
        @remove-service="removeService"
        @add-port="addPort"
        @remove-port="removePort"
    ></component>
  </UForm>
</template>
