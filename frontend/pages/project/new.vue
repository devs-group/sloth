<script setup lang="ts">
import {projectSchema} from "~/schema/schema";
import DockerCredentialsForm from "~/components/docker-credentials-form.vue";
import ServicesForm from "~/components/services-form.vue";

import type {ProjectSchema, Service, ServiceSchema} from "~/schema/schema"

const tabItems = ref([{
  label: 'Services',
  command: () => onChangeTab(0),
  __component: ServicesForm,
}, {
  label: 'Docker credentials',
  command: () => onChangeTab(1),
  __component: DockerCredentialsForm,
}, {
  label: 'Monitoring (coming soon)',
  command: () => onChangeTab(2),
  disabled: true,
}])

const isSubmitting = ref(false)
const toast = useToast()
const router = useRouter()
const config = useRuntimeConfig()

const p = ref<ProjectSchema>({
  name: "",
  services: [],
  docker_credentials: [],
})
const activeTabComponent = ref(tabItems.value[0].__component)

function onChangeTab(idx: number) {
  activeTabComponent.value = tabItems.value[idx].__component
}

async function saveProject () {
  const data = p.value
  isSubmitting.value = true
  try {
    await $fetch(`${config.public.backendHost}/v1/project`, { method: "POST", body: data, credentials: "include" })
    toast.add({
      severity: 'success',
      summary: "Success",
      detail: "Your project has been created successfully",
      life: 3000
    })
    await router.push("/")
  } catch (e) {
    console.error(e)
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Something went wrong",
      life: 3000
    })
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
      hosts: [""],
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

function addHost(serviceIdx: number) {
  p.value?.services[serviceIdx].public.hosts.push("")
}

function removeHost(hostIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].public.hosts.splice(hostIdx, 1)
}

</script>

<template>
  <form
    class="p-12 flex flex-col flex-1 overflow-hidden"
  >
    <div class="flex flex-row pb-12">
      <InputGroup>
        <InputText v-model="p.name" class="max-w-[20em]" />
        <IconButton
          label="Create Project"
          icon="heroicons:bolt"
          @click="saveProject"
          :disabled="!p?.name || p.services.length === 0"
          :loading="isSubmitting"
        />
      </InputGroup>
    </div>

    <Menubar :model="tabItems"/>
    <component
        :is="activeTabComponent"
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
        @add-host="addHost"
        @remove-host="removeHost"
    ></component>
  </form>
</template>
