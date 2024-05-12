<template>
  <div class="relative flex flex-col gap-4 w-full max-w-6xl p-6">
    <NuxtLink class="lg:hidden" :to="{name: Routes.PROJECTS}">
      <IconButton icon="heroicons:arrow-uturn-left"/>
    </NuxtLink>

    <template v-if="p">
      <div class="flex flex-col gap-2">
        <div class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-2">
          <div class="flex flex-col gap-1 max-w-full">
            <p class="text-sm text-prime-secondary-text">Project Name</p>
            <p class="break-all">{{ p.name }}</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button
                :loading="isUpdatingLoading"
                label="Save"
                @click="updateProject"
            />
            <Button
                :loading="isUpdatingLoading"
                label="Save & Restart"
                @click="updateProject"
            />
          </div>
        </div>

        <div class="flex flex-col gap-1 max-w-full">
          <p class="text-sm text-prime-secondary-text">Project Unique Name</p>
          <div class="flex items-center gap-1">
            <p>{{ p.upn }}</p>
            <CopyButton :string="p.upn!" />
          </div>
        </div>

        <div
            v-if="p.services.find((s) => s.public.enabled)"
            class="flex flex-col gap-1 max-w-full"
        >
          <p class="text-sm text-prime-secondary-text">Public URL's</p>
          <div v-for="service in p.services.filter((s) => s.public.enabled)">
            <template v-if="service.public.hosts.some(url => url.trim().length > 0)">
              <div
                  v-for="host in service.public.hosts"
                  class="flex items-center gap-1"
              >
                <Icon icon="heroicons:link" />
                <a :href="`//${host}`" target="_blank">{{ host }}</a>
                <CopyButton :string="host" />
              </div>
            </template>
            <p v-else>No public urls yet</p>
          </div>
        </div>

        <div class="flex flex-col gap-1 max-w-full">
          <p class="text-sm text-prime-secondary-text">Deployment Webhook</p>
          <div class="flex flex-wrap gap-1 items-center">
            <p class="break-words">{{ p.hook }}</p>
            <CopyButton :string="p.hook!" />
          </div>

          <p class="text-sm text-prime-secondary-text">Deployment Access Token</p>
          <div class="flex gap-1 items-center">
            <p class="break-words">{{ p.access_token }}</p>
            <CopyButton :string="p.access_token!" />
          </div>

          <p class="text-sm text-prime-secondary-text">Example Command</p>
          <div class="flex items-center">
            <code class="text-sm text-prime-secondary-text">
              {{ hookCurlCmd(p.hook!, p.access_token!) }}
            </code>
            <CopyButton
                :string="hookCurlCmd(p.hook!, p.access_token!)"
            ></CopyButton>
          </div>
        </div>
      </div>

      <form @submit.prevent>
        <!-- TABS -->
        <Menubar :model="tabItems" @change="onChangeTab" />

        <!-- Service states -->
        <div
            class="flex flex-col gap-2"
            v-if="
        Object.values(p.services).length > 0 &&
        activeTabComponent?.__name == 'services-form'
      "
        >
          <p class="text-prime-secondary-text">Service stats</p>
          <div class="flex gap-6">
            <div
                class="flex flex-col gap-1"
                v-for="(service, sIdx) in Object.values(p.services)"
            >
              <template v-if="service.usn && serviceStates[service.usn]">
                <div>
                  <p class="pb-2">{{ service.name }}</p>
                  <p class="text-xs text-prime-secondary-text">
                    State: {{ serviceStates[service.usn].state }}
                  </p>
                  <p class="text-xs text-prime-secondary-text">
                    Status: {{ serviceStates[service.usn].status }}
                  </p>
                </div>
                <Button
                    label="Show logs"
                    @click="streamServiceLogs(p.id!, service.usn)"
                />
                <Dialog
                    v-model:visible="isLogsModalOpen"
                    :header="service.name + ' Logs'"
                    modal
                >
                  <div class="overflow-auto h-[80vh]">
                    <code class="text-xs" v-for="l in logsLines">
                      <p>{{ l }}</p>
                    </code>
                  </div>
                </Dialog>
              </template>
            </div>
          </div>
        </div>

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
    <Message v-else-if="pageErrorMessage" severity="error" :closable="false">
      {{pageErrorMessage}}, <NuxtLink class="underline" :to="{name: Routes.PROJECTS}">go back</NuxtLink>
    </Message>
    <OverlayProgressSpinner :show="isLoading"/>
  </div>
</template>

<script lang="ts" setup>
import type {ProjectSchema} from "~/schema/schema";
import {useWebSocket} from "@vueuse/core";
import DockerCredentialsForm from "~/components/docker-credentials-form.vue";
import ServicesForm from "~/components/services-form.vue";
import {Constants} from "~/config/const";
import {Routes} from "~/config/routes";

const route = useRoute();
const config = useRuntimeConfig();
const toast = useToast();

const pageErrorMessage = ref('')
const isLoading = ref(true)

interface ServiceState {
  state: string;
  status: string;
}

const tabItems = [
  {
    label: "Services",
    command: () => onChangeTab(0),
    __component: ServicesForm,
  },
  {
    label: "Docker credentials",
    command: () => onChangeTab(1),
    __component: DockerCredentialsForm,
  },
  {
    label: "Monitoring (coming soon)",
    command: () => onChangeTab(2),
    disabled: true,
  },
];

const id = route.params.id;
const p = ref<ProjectSchema>();
const isUpdatingLoading = ref(false);
const serviceStates = ref<Record<string, ServiceState>>({});
const isLogsModalOpen = ref(false);
const logsLines = ref<string[]>([]);
const activeTabComponent = shallowRef(tabItems[0].__component);

fetchProject()

function onChangeTab(idx: number) {
  activeTabComponent.value = tabItems[idx].__component;
}

async function updateProject() {
  const data = p.value;
  isUpdatingLoading.value = true;
  try {
    await $fetch<ProjectSchema>(
        `${config.public.backendHost}/v1/project/${id}`,
        {
          method: "PUT",
          credentials: "include",
          body: data,
        }
    );
    await fetchProject();
    await fetchServiceStates();
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Project has been updated",
      life: Constants.ToasterDefaultLifeTime,
    });
  } catch (e) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to update project",
      life: Constants.ToasterDefaultLifeTime,
    });
  } finally {
    isUpdatingLoading.value = false;
  }
}

async function fetchProject() {
  isLoading.value = true
  $fetch<ProjectSchema>(
      `${config.public.backendHost}/v1/project/${id}`,
      { credentials: "include" }
  )
      .then((payload) => {
        p.value = payload;
        fetchServiceStates()
            .then((payload) => {
              serviceStates.value = payload
            })
            .catch(() => {
              toast.add({
                severity: "error",
                summary: "Error",
                detail: "Unable to receive project state",
                life: Constants.ToasterDefaultLifeTime,
              });
            })
            .finally(() => {
              isLoading.value = false
            })
      })
      .catch(() => {
        isLoading.value = false
        pageErrorMessage.value = "Sorry we can't find this project"
      })
}

async function fetchServiceStates() {
  return $fetch<Record<string, ServiceState>>(
      `${config.public.backendHost}/v1/project/state/${id}`,
      {
        method: "GET",
        credentials: "include",
      }
  )
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
    env_vars: [["", ""]],
    volumes: [""],
    healthcheck: {
      test: "curl -f http://localhost/ || exit 1",
      interval: "30s",
      timeout: "10s",
      retries: 3,
      start_period: "15s",
    },
    depends_on: {
      //"autumn-frost": { condition: "service_healthy" },
    },
    // TODO: @4ddev Das sollte nicht aus dem Frontend kommen, zumindest erstmal regeln wir das nur Backendseitig
    deploy: {
      mode:"replicated",
      replicas: 3,
      endpoint_mode: "vip",
      resources: {
        limits: {
          cpus: "2.0",
          memory: "8g",
          pids: 100,
        },
        reservations: {
          cpus: "1.0",
          memory: "500m",
        }
      },
      restart_policy: {
        condition: "on-failure",
        delay: "5s",
        max_attempts: 3,
        window: "120s",
      }
    }
  });
}

function streamServiceLogs(id: number, usn: string) {
  isLogsModalOpen.value = true;
  logsLines.value = [];

  const wsBackendHost = config.public.backendHost.replace("http", "ws");
  const { status, data, close } = useWebSocket(
      `${wsBackendHost}/v1/ws/project/logs/${usn}/${id}`,
      {
        autoReconnect: {
          retries: 5,
          delay: 1000,
          onFailed() {
            toast.add({
              severity: "error",
              summary: "Error",
              detail: "Unable to stream logs",
              life: Constants.ToasterDefaultLifeTime,
            });
          },
        },
      }
  );

  watchEffect(() => {
    logsLines.value?.push(data.value);
  });
}

function addCredential() {
  p.value?.docker_credentials.push({
    username: "",
    password: "",
    registry: "",
  });
}

function removeCredential(idx: number) {
  p.value?.docker_credentials.splice(idx, 1);
}

function removeService(idx: number) {
  p.value?.services.splice(idx, 1);
}

function addEnv(serviceIdx: number) {
  p.value?.services[serviceIdx].env_vars.push(["", ""]);
}

function removeEnv(envIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].env_vars.splice(envIdx, 1);
}

function addVolume(serviceIdx: number) {
  p.value?.services[serviceIdx].volumes.push("");
}

function removeVolume(volIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].volumes.splice(volIdx, 1);
}

function addPort(serviceIdx: number) {
  p.value?.services[serviceIdx].ports.push("");
}

function removePort(portIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].ports.splice(portIdx, 1);
}

function hookCurlCmd(url: string, accessToken: string) {
  return `curl -X GET "${url}" -H "X-Access-Token: ${accessToken}"`;
}

function addHost(serviceIdx: number) {
  p.value?.services[serviceIdx].public.hosts.push("");
}

function removeHost(hostIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].public.hosts.splice(hostIdx, 1);
}
</script>