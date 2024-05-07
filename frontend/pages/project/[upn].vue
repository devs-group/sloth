<script lang="ts" setup>
import { useWebSocket } from "@vueuse/core";
import { useTabs } from "~/composables/useTabs";
import type { ProjectSchema } from "~/schema/schema";

const route = useRoute();
const upn = route.params.upn;
const config = useRuntimeConfig();
const toast = useToast();

interface ServiceState {
  state: string;
  status: string;
}

const { tabs, activeTabComponent } = useTabs();

const p = ref<ProjectSchema | undefined>();
const { addService, removeService, addPort, addCredential, addEnv ,addHost, addVolume,
        removeCredential, removeEnv, removeHost, removePort, removeVolume } = useService(p);
const isUpdatingLoading = ref(false);
const serviceStates = ref<Record<string, ServiceState>>({});
const isLogsModalOpen = ref(false);
const logsLines = ref<string[]>([]);

onMounted(() => {
  fetchProject();
  fetchServiceStates();
});

async function updateProject() {
  const data = p.value;
  isUpdatingLoading.value = true;
  try {
    await $fetch<ProjectSchema>(
      `${config.public.backendHost}/v1/project/${upn}`,
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
      life: 3000,
    });
  } catch (e) {
    console.error("unable to update project", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to update project",
      life: 3000,
    });
  } finally {
    isUpdatingLoading.value = false;
  }
}

async function fetchProject() {
  try {
    p.value = await $fetch<ProjectSchema>(
      `${config.public.backendHost}/v1/project/${upn}`,
      { credentials: "include" }
    );
  } catch (e) {
    console.error("unable to fetch project", e);
  }
}

async function fetchServiceStates() {
  try {
    serviceStates.value = await $fetch<Record<string, ServiceState>>(
      `${config.public.backendHost}/v1/project/state/${upn}`,
      {
        method: "GET",
        credentials: "include",
      }
    );
  } catch (e) {
    console.error("unable to fetch project state", e);
  }
}

function streamServiceLogs(upn: string, usn: string) {
  isLogsModalOpen.value = true;
  logsLines.value = [];

  const wsBackendHost = config.public.backendHost.replace("http", "ws");
  const { status, data, close } = useWebSocket(
    `${wsBackendHost}/v1/ws/project/logs/${usn}/${upn}`,
    {
      autoReconnect: {
        retries: 5,
        delay: 1000,
        onFailed() {
          toast.add({
            severity: "error",
            summary: "Error",
            detail: "Unable to stream logs",
            life: 3000,
          });
        },
      },
    }
  );

  watchEffect(() => {
    logsLines.value?.push(data.value);
  });
}

function hookCurlCmd(url: string, accessToken: string) {
  return `curl -X GET "${url}" -H "X-Access-Token: ${accessToken}"`;
}
</script>

<template>
  <form class="w-full p-12" v-if="p">
    <div class="flex flex-col gap-4 mb-12">
      <div class="flex justify-between">
        <div class="flex flex-col gap-1">
          <p class="text-sm text-prime-secondary-text">Project name</p>
          <p>{{ p.name }}</p>
        </div>
        <Button
          :loading="isUpdatingLoading"
          label="Save & restart"
          @click="updateProject"
        />
      </div>
      <div class="flex flex-col gap-1">
        <p class="text-sm text-prime-secondary-text">Project unique name</p>
        <p>{{ p.upn }}</p>
      </div>
      <div
        v-if="p.services.find((s) => s.public.enabled)"
        class="flex flex-col gap-1"
      >
        <p class="text-sm text-prime-secondary-text">Project URL's</p>
        <div v-for="service in p.services.filter((s) => s.public.enabled)">
          <div
            v-for="host in service.public.hosts"
            class="flex items-center gap-2"
          >
            <Icon icon="heroicons:link" />
            <NuxtLink :to="'//' + host" target="_blank">{{ host }}</NuxtLink>
            <CopyButton :string="host" />
          </div>
        </div>
      </div>
      <div class="flex flex-col gap-1">
        <p class="text-sm text-prime-secondary-text">Deployment webhook</p>
        <div class="flex gap-4 items-center">
          <p>URL:</p>
          <p class="whitespace-nowrap">{{ p.hook }}</p>
          <CopyButton :string="p.hook!" />
        </div>
        <div class="flex gap-4 items-center">
          <p>Access Token:</p>
          <p class="whitespace-nowrap">{{ p.access_token }}</p>
          <CopyButton :string="p.access_token!" />
        </div>
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

    <!-- TABS -->
    <Menubar :model="tabs" />

    <!-- Service states -->
    <div
      class="flex flex-col gap-2 my-8"
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
              @click="streamServiceLogs(p.upn!, service.usn)"
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
