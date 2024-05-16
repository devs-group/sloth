<template>
  <div class="relative flex flex-col gap-4 w-full max-w-6xl p-6">
    <NuxtLink class="lg:hidden" :to="{name: Routes.PROJECTS}">
      <IconButton icon="heroicons:arrow-uturn-left"/>
    </NuxtLink>

    <template v-if="project">
      <ProjectInfo :project="project"></ProjectInfo>
      <form @submit.prevent>
        <Menubar :model="tabItems" @change="onChangeTab" />
        <div
            class="flex flex-col gap-2"
            v-if="
        Object.values(project.services).length > 0 &&
        activeTabComponent?.__name == 'services-form'
      "
        >
          <p class="text-prime-secondary-text">Service stats</p>
          <div class="flex gap-6">
            <div
                class="flex flex-col gap-1"
                v-for="(service, sIdx) in Object.values(project.services)"
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
                    @click="streamServiceLogs(project.id!, service.usn, logsLines)"
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
            :credentials="project.docker_credentials"
            @add-credential="addCredential"
            @remove-credential="removeCredential"
            :services="project.services"
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
import { ref, computed } from 'vue';
import type { IServiceState, TabItem } from '~/config/interfaces';
import type { ProjectSchema } from '~/schema/schema';
import { Routes } from '~/config/routes';
import ServicesForm from '~/components/services-form.vue';
import DockerCredentialsForm from '~/components/docker-credentials-form.vue';
import ProjectInfo from '~/components/project/project-info.vue';

const p = ref<ProjectSchema | null>(null);
const serviceStates = ref<Record<string, IServiceState>>({});
const logsLines = ref<string[]>([]);
const pageErrorMessage = ref('');
const isLogsModalOpen = ref(false);

const tabItems = [
  { label: "Services", component: ServicesForm },
  { label: "Docker Credentials", component: DockerCredentialsForm },
  { label: "Monitoring", disabled: true }
] as TabItem[];

const { activeTabComponent, onChangeTab } = useTabs(tabItems);
const route = useRoute();
const projectID = route.params.id;
const { project, isLoading, fetchProject } = useProject(projectID[0])
const { addCredential, removeCredential,
        addEnv, removeEnv, addHost, 
        removeHost, addPort, removePort, 
        addService, removeService, addVolume, 
        removeVolume, hookCurlCmd, streamServiceLogs } = useService(project);

const hasServices = computed(() => Object.values(project.value?.services || {}).length > 0);
const tabProps = computed(() => ({ credentials: project.value?.docker_credentials, services: project.value?.services }));

onMounted(async () => {
  await fetchProject()
  console.log(project.value)
})
</script>
