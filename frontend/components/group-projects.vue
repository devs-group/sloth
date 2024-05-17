<script lang="ts" setup>
import type { GroupProject } from "~/schema/schema";
const confirm = useConfirm();

const toast = useToast();
const route = useRoute();

const organization_name = route.params.organization_name;
const g = ref<GroupProject[]>();
const isAddGroupProjectModalOpen = ref(false);
const projectUPN = ref("");

interface State {
  isRemoving?: boolean;
}

const config = useRuntimeConfig();

async function fetchGroupProjects() {
  try {
    g.value = await $fetch<GroupProject[]>(
      `${config.public.backendHost}/v1/organization/${organization_name}/projects`,
      { credentials: "include" }
    );
    console.log(g.value);
  } catch (e) {
    console.error("unable to fetch Group", e);
  }
}

onMounted(() => {
  fetchGroupProjects();
});

async function addProject() {
  try {
    g.value = await $fetch(
      `${config.public.backendHost}/v1/organization/project`,
      {
        method: "PUT",
        credentials: "include",
        body: {
          organization_name: organization_name,
          upn: projectUPN.value,
        },
      }
    );
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Project added to group",
        life: 3000
    });
  } catch (e) {
    console.error("unable to invite", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to add Project",
    });
  } finally {
    isAddGroupProjectModalOpen.value = false;
    fetchGroupProjects();
  }
}

async function removeProject(upn: string) {
  try {
    g.value = await $fetch(
      `${config.public.backendHost}/v1/organization/project`,
      {
        method: "DELETE",
        credentials: "include",
        body: {
          organization_name: organization_name,
          upn: upn,
        },
      }
    );
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Project removed from group",
        life: 3000
    });
  } catch (e) {
    console.error("unable to invite", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to remove Project",
    });
  } finally {
    fetchGroupProjects();
  }
}
</script>
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
        @click="isAddGroupProjectModalOpen = true"
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
                      'You were invited to a new Group do you wanna participate to the Group?',
                    accept: () => removeProject(project.upn),
                    acceptLabel: 'Remove',
                    rejectLabel: 'Cancel',
                  })
              "
            />
            <NuxtLink :to="`/project/${project.upn}`">
              <IconButton icon="heroicons:arrow-right-on-rectangle" />
            </NuxtLink>
          </div>
        </div>
      </template>
    </Card>
  </div>
  <Dialog
    v-model:visible="isAddGroupProjectModalOpen"
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
