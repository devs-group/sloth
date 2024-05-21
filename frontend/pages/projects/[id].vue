<template>
  <div class="relative flex flex-col gap-4 w-full max-w-6xl p-6">
    <NuxtLink class="lg:hidden" :to="{name: Routes.PROJECTS}">
      <IconButton icon="heroicons:arrow-uturn-left"/>
    </NuxtLink>

    <template v-if="project">
      <ProjectInfo 
        :project="project" 
        :isUpdatingLoading="isUpdatingLoading"
        @updateProject="updateProject(project)">
      </ProjectInfo>

      <form @submit.prevent>
        <Menubar :model="tabItems" @change="onChangeTab" />
        <div class="flex flex-col gap-2" v-if="hasServices">
          <p class="text-prime-secondary-text">Service stats</p>
          <div class="flex gap-6">
            <div class="flex flex-col gap-1" v-for="(service, _) in Object.values(project.services)">
              <ServiceDetail :service="serviceStates[service.usn!]" :logs-lines="logsLines"
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
import type { Project } from '~/schema/schema';
import { Routes } from '~/config/routes';
import ServicesForm from '~/components/services-form.vue';
import DockerCredentialsForm from '~/components/docker-credentials-form.vue';
import ProjectInfo from '~/components/project-info.vue';

const route = useRoute();
const projectID = parseInt(route.params.id.toString());

const project = ref<Project | null>(null);
const { isLoading, fetchProject, isUpdatingLoading, updateProject } = useProject()

const { addCredential, removeCredential,
        addEnv, removeEnv, addHost, 
        removeHost, addPort, removePort, 
        addService, removeService, addVolume, 
        removeVolume, streamServiceLogs, fetchServiceStates } = useService(project);
 

const serviceStates = ref<Record<string, IServiceState>>({});
const logsLines = ref<string[]>([]);
const pageErrorMessage = ref('');
const isLogsModalOpen = ref(false);

const tabItems = computed(()=> [
  { label: "Services", component: ServicesForm, command: () => onChangeTab(0) },
  { label: "Docker Credentials", component: DockerCredentialsForm, command: () => onChangeTab(1) },
  { label: "Monitoring", disabled: true }
] as TabItem[]);

const { activeTabComponent, onChangeTab } = useTabs(tabItems);
const hasServices = computed(() => Object.values(project.value?.services || {}).length > 0);

onMounted(() => {
  fetchProject(projectID).then(async (fetchedProject) => {
    project.value = fetchedProject;
    try {
      const records = await fetchServiceStates(project.value!.id.toString());
      if (records) {
        Object.keys(records).forEach((key) => {
          const state = records[key];
          serviceStates.value[key] = state;
          console.log(`Service ID: ${key}, State: ${state.state}, Status: ${state.status}`);
        });
      }
    } catch (error) {
      console.error("Failed to fetch states for services in the project", project.value?.id, error);
    }
  }).catch(error => {
    console.error("Failed to fetch project details", error);
  });
});

</script>
