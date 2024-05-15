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
                    accept: () => removeProjectFromOrganisation(project.upn),
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
            @click="addProjectToOrganisation"
            class="invite-button text-green-500 hover:text-green-700 focus:outline-none"
        />
      </div>
    </div>
  </Dialog>
</template>

<script lang="ts" setup>
import { Routes } from "~/config/routes";
import { type OrganisationProject } from "~/schema/schema";

const confirm = useConfirm();
const route = useRoute();
const config = useRuntimeConfig();
const organisation_name = route.params.organisation_name;
const projectUPN = ref("");
const isAddOrganisationProjectModalOpen = ref(false);

const g = ref<OrganisationProject[]>()

const { addProjectToOrganisation, removeProjectFromOrganisation, fetchOrganisationProjects } = useOrganisation(organisation_name, config, confirm);

onMounted(() => {
  fetchOrganisationProjects();
});
</script>