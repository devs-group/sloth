<template>
  <div class="pl-5">
    <div class="flex justify-between items-center mb-2">
      <div>
        <h1 class="text-2xl">Projects</h1>
        <p class="text-sm text-gray-400">{{ g?.length }} Projects</p>
      </div>
      <IconButton
          icon="heroicons:pencil-square"
          label="Add Project"
          size="sm"
          color="gray"
          variant="solid"
          :trailing="false"
          @click="isAddOrganisationProjectModalOpen = true"
      />
    </div>
    <div>Projects</div>
    <Card v-for="project in g" :key="project.upn" class="mb-2">
      <template #content>
        <div class="flex justify-between">
          <h3>{{ project.name }}</h3>
          <div class="flex space-x-2">
            <IconButton
                text
                severity="danger"
                icon="heroicons:trash"
                @click="
                () =>
                  confirm.require({
                    header: 'Accept invitation?',
                    message:
                      'You were invited to a new Organisation do you wanna participate to the Organisation?',
                    accept: () => removeProject(project.upn),
                    acceptLabel: 'Remove',
                    rejectLabel: 'Cancel',
                  })
              "
            />
            <NuxtLink :to="{path: Routes.ORGANISATION}">
              <IconButton icon="heroicons:arrow-right-on-rectangle" />
            </NuxtLink>
          </div>
        </div>
      </template>
    </Card>
  </div>
  <Dialog
      v-model:visible="isAddOrganisationProjectModalOpen"
      header="Add Project"
      modal
  >
    <div class="flex flex-col space-y-4 p-6">
      <div class="flex flex-row items-center space-x-4">
        <input class="w-full" v-model="projectUPN" />
        <IconButton
            @click="addProject"
            class="invite-button text-green-500 hover:text-green-700 focus:outline-none"
        />
      </div>
    </div>
  </Dialog>
</template>

<script lang="ts" setup>
import type { OrganisationProjectSchema } from "~/schema/schema";
import {Routes} from "~/config/routes";
import {Constants} from "~/config/const";
const confirm = useConfirm();

const toast = useToast();
const route = useRoute();

const organisation_name = route.params.organisation_name;
const g = ref<OrganisationProjectSchema[]>();
const isAddOrganisationProjectModalOpen = ref(false);
const projectUPN = ref("");

interface State {
  isRemoving?: boolean;
}

const config = useRuntimeConfig();

async function fetchOrganisationProjects() {
  try {
    g.value = await $fetch<OrganisationProjectSchema[]>(
        `${config.public.backendHost}/v1/organisation/${organisation_name}/projects`,
        { credentials: "include" }
    );
  } catch (e) {
    console.error("unable to fetch Organisation", e);
  }
}

onMounted(() => {
  fetchOrganisationProjects();
});

async function addProject() {
  try {
    g.value = await $fetch(
        `${config.public.backendHost}/v1/organisation/project`,
        {
          method: "PUT",
          credentials: "include",
          body: {
            organisation_name: organisation_name,
            upn: projectUPN.value,
          },
        }
    );
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Project added to organisation",
      life: Constants.ToasterDefaultLifeTime,
    });
  } catch (e) {
    console.error("unable to invite", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to add Project",
      life: Constants.ToasterDefaultLifeTime,
    });
  } finally {
    isAddOrganisationProjectModalOpen.value = false;
    fetchOrganisationProjects();
  }
}

async function removeProject(upn: string) {
  try {
    g.value = await $fetch(
        `${config.public.backendHost}/v1/organisation/project`,
        {
          method: "DELETE",
          credentials: "include",
          body: {
            organisation_name: organisation_name,
            upn: upn,
          },
        }
    );
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Project removed from organisation",
      life: Constants.ToasterDefaultLifeTime,
    });
  } catch (e) {
    console.error("unable to invite", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to remove Project",
      life: Constants.ToasterDefaultLifeTime,
    });
  } finally {
    fetchOrganisationProjects();
  }
}
</script>