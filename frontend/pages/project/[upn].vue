<script lang="ts" setup>
import {projectSchema, ProjectSchema, ServiceSchema} from "~/schema/schema";
import {useWebSocket} from "@vueuse/core";
import DockerCredentialsForm from "~/components/docker-credentials-form.vue";
import {FormSubmitEvent} from "@nuxt/ui/dist/runtime/types";

const route = useRoute()
const upn = route.params.upn
const config = useRuntimeConfig()
const { showError, showSuccess } = useNotification()

interface ServiceState {
  state: string
  status: string
}

const tabItems = [{
  label: 'Services',
  // __component: DockerCredentialsForm,
}, {
  label: 'Docker credentials',
  __component: DockerCredentialsForm,
}, {
  label: 'Monitoring (coming soon)',
  disabled: true,
}]

const p = ref<ProjectSchema>()
const isUpdatingLoading = ref(false)
const isChangeProjectNameModalOpen = ref(false)
const serviceStates = ref<Record<string, ServiceState>>({})
const isLogsModalOpen = ref(false)
const logsLines = ref<string[]>([])
const isLogsModalFullScreen = ref(false)
const activeTabComponent = ref(tabItems[0].__component)

onMounted(() => {
  fetchProject()
  fetchServiceStates()
})

function onChangeTab(idx: number) {
  activeTabComponent.value = tabItems[idx].__component
}

async function updateProject(event: FormSubmitEvent<ProjectSchema>) {
  const data = projectSchema.parse(event.data)
  isUpdatingLoading.value = true
  try {
    await $fetch<ProjectSchema>(
        `${config.public.backendHost}/v1/project/${upn}`,
        {
          method: "PUT",
          credentials: "include",
          body: data
        },
    )
    await fetchProject()
    await fetchServiceStates()
    showSuccess("Success", "Project has been updated")
  } catch (e) {
    console.error("unable to update project", e)
    showError("Error", "Unable to update project")
  } finally {
    isUpdatingLoading.value = false
  }
}

async function fetchProject() {
  try {
    p.value = await $fetch<ProjectSchema>(
        `${config.public.backendHost}/v1/project/${upn}`,
        { credentials: "include" },
    )
  } catch (e) {
    console.error("unable to fetch project", e)
  }
}

async function fetchServiceStates() {
  try {
    serviceStates.value = await $fetch<Record<string, ServiceState>>(
        `${config.public.backendHost}/v1/project/state/${upn}`,
        {
          method: "GET",
          credentials: "include",
        },
    )
  } catch (e) {
    console.error("unable to fetch project state", e)
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

function streamServiceLogs(upn: string, service: string) {
  isLogsModalOpen.value = true
  logsLines.value = []

  const wsBackendHost = config.public.backendHost.replace("http", "ws")
  const { status, data, close } = useWebSocket(`${wsBackendHost}/v1/ws/project/logs/${service}/${upn}`, {
    autoReconnect: {
      retries: 5,
      delay: 1000,
      onFailed() {
        showError("Error", "unable to stream logs")
      },
    },
  })

  watchEffect(() => {
    logsLines.value?.push(data.value)
  })
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

function removeService(idx: number) {
  p.value?.services.splice(idx, 1)
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

function hookCurlCmd(url: string, accessToken: string) {
  return `curl -X GET "${url}" -H "X-Access-Token: ${accessToken}"`
}
</script>

<template>
  <div>
    <div v-if="p">
      <div class="py-12 px-6 space-y-4">
        <UForm
            :schema="projectSchema"
            :state="p"
            @submit="updateProject"
        >
        <div class="flex flex-row justify-between items-center">
          <div>
            <p class="text-sm text-gray-500">Project name</p>
            <div class="flex flex-row items-center space-x-4">
              <p>{{ p.name }}</p>
              <UBadge class="cursor-pointer" @click="isChangeProjectNameModalOpen = true">Change</UBadge>

              <UModal v-model="isChangeProjectNameModalOpen">
                <UFormGroup name="name">
                  <div class="flex flex-row items-center w-full space-x-4 p-6" >
                    <UInput class="w-full" v-model="p.name"/>
                    <UButton icon="i-heroicons-check" @click="isChangeProjectNameModalOpen = false"></UButton>
                  </div>
                </UFormGroup>
              </UModal>
          </div>

          </div>
          <div>
            <UButton type="submit" :loading="isUpdatingLoading">Save & restart</UButton>
          </div>
        </div>

        <div>
          <p class="text-sm text-gray-500">Project unique name</p>
          <p>{{ p.upn }}</p>
        </div>

        <div v-if="p.services.find((s) => s.public.enabled)">
          <p class="text-sm text-gray-500">Project URL's</p>
          <div v-for="s in p.services.filter((s) => s.public.enabled)" class="flex flex-row items-center space-x-2">
            <UIcon name="i-heroicons-link"></UIcon>
            <ULink :to="'//' + s.public.host" target="_blank">{{ s.public.host }}</ULink>
            <CopyButton :string="s.public.host as string"></CopyButton>
          </div>
        </div>


        <div>
          <p class="text-sm text-gray-500">Deployment webhook</p>
          <div class="flex flex-row items-center space-x-2">
            <p class="text-sm text-gray-500">URL:</p>
            <p>{{ p.hook }}</p>
            <CopyButton :string="p.hook as string"></CopyButton>
          </div>
          <div class="flex flex-row items-center space-x-2">
            <p class="text-sm text-gray-500">
              Access Token:
            </p>
            <p>{{ p.access_token }}</p>
            <CopyButton :string="p.access_token as string"></CopyButton>
          </div>
          <div class="flex flex-row items-center space-x-2">
            <code class="text-sm text-gray-500">
              {{ hookCurlCmd(p.hook as string, p.access_token as string) }}
            </code>
            <CopyButton :string="hookCurlCmd(p.hook as string, p.access_token as string)"></CopyButton>
          </div>
        </div>

          <!-- TABS -->
          <UTabs :items="tabItems" @change="onChangeTab" />
          <component
              :is="activeTabComponent as string"
              :credentials="p.docker_credentials"
              @add-credential="addCredential"
              @remove-credential="removeCredential"
          ></component>

          <!-- Service states -->
          <div v-if="Object.values(p.services).length > 0">
            <p class="text-gray-400 py-2">Services stats</p>
            <div class="space-x-4">
              <div v-for="(s, idx) in Object.values(p.services)" class="inline-block">
                <div v-if="serviceStates[s.name]" class="flex flex-col">
                  <p>{{ s.name }}</p>
                  <div class="space-y-2">
                    <div>
                      <p class="text-sm text-gray-500">State: {{ serviceStates[s.name].state }}</p>
                      <p class="text-sm text-gray-500">Status: {{ serviceStates[s.name].status }}</p>
                    </div>

                    <div>
                      <UButton size="xs" @click="streamServiceLogs(p.upn as string, s.name)">Show logs</UButton>
                      <UModal v-model="isLogsModalOpen" :fullscreen="isLogsModalFullScreen">
                        <div class="p-3">
                          <div class="flex flex-row space-between items-center w-full">
                            <p class="w-full text-sm text-gray-500">
                              {{ s.name }} Logs
                            </p>
                            <div class="w-full flex flex-row justify-end space-x-2 pb-3">
                              <UButton v-if="isLogsModalFullScreen" icon="i-heroicons-arrows-pointing-in" type="ghost" @click="isLogsModalFullScreen = false"></UButton>
                              <UButton v-else icon="i-heroicons-arrows-pointing-out" type="ghost" @click="isLogsModalFullScreen = true"></UButton>
                              <UButton icon="i-heroicons-x-mark" type="ghost" @click="isLogsModalOpen = false"></UButton>
                            </div>
                          </div>

                          <div class="h-[80vh] overflow-auto">
                            <code class="text-xs" v-for="l in logsLines">
                              <p>{{ l }}</p>
                            </code>
                          </div>

                        </div>
                      </UModal>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>


            <ServicesForm
                :services="p.services"
                @add-service="addService"
                @add-env="addEnv"
                @remove-env="removeEnv"
                @add-volume="addVolume"
                @remove-volume="removeVolume"
                @remove-service="removeService"
                @add-port="addPort"
                @remove-port="removePort"
            ></ServicesForm>
        </UForm>
      </div>
    </div>
  </div>
</template>