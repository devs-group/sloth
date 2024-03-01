<script lang="ts" setup>
import type { Group, GroupProject } from "~/schema/schema";
const { showConfirmation } = useConfirmation();

const { showError, showSuccess } = useNotification();
const route = useRoute();

const group_name = route.params.group_name;
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
      `${config.public.backendHost}/v1/group/${group_name}/projects`,
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

async function addProject(upn: string) {
  console.log("Add project");
  try {
    g.value = await $fetch(`${config.public.backendHost}/v1/group/project`, {
      method: "PUT",
      credentials: "include",
      body: {
        group_name: group_name,
        upn: upn,
      },
    });
    showSuccess("Success", "Project added to group");
  } catch (e) {
    console.error("unable to invite", e);
    showError("Error", "Unable to add Project");
  } finally {
    isAddGroupProjectModalOpen.value = false;
    fetchGroupProjects();
  }
}

async function removeProject(upn: string) {
  try {
    g.value = await $fetch(`${config.public.backendHost}/v1/group/project`, {
      method: "DELETE",
      credentials: "include",
      body: {
        group_name: group_name,
        upn: upn,
      },
    });
    showSuccess("Success", "Project removed from group");
  } catch (e) {
    console.error("unable to invite", e);
    showError("Error", "Unable to remove Project");
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

      <UButton
        icon="i-heroicons-pencil-square"
        size="sm"
        color="gray"
        variant="solid"
        :trailing="false"
        @click="isAddGroupProjectModalOpen = true"
      >
        Add Project
      </UButton>
    </div>
    <div>Projects</div>
    <ul class="list-disc pl-5">
      <li
        v-for="project in g"
        :key="project.upn"
        class="flex justify-between items-center mb-2"
      >
        <span class="text-gray-800">{{ project.name }}</span>
        <div>
          <UButton
            icon="i-heroicons-trash"
            variant="ghost"
            color="red"
            @click="
              () =>
                showConfirmation(
                  'Remove the Project?',
                  'Are you sure you wanna remove this project from your group?',
                  () => removeProject(project.upn)
                )
            "
          >
          </UButton>
          <NuxtLink :to="`/project/${project.upn}`">
            <UButton icon="i-heroicons-arrow-right-on-rectangle"></UButton>
          </NuxtLink>
        </div>
      </li>
    </ul>
  </div>
  <UModal v-model="isAddGroupProjectModalOpen">
    <UFormGroup name="name">
      <div class="flex flex-col space-y-4 p-6">
        <div class="flex flex-row items-center space-x-4">
          <UInput class="w-full" v-model="projectUPN" />
          <UButton
            @click="addProject(projectUPN)"
            class="invite-button text-green-500 hover:text-green-700 focus:outline-none"
          >
            <i class="icon heroicons-check text-lg"></i>
          </UButton>
        </div>
      </div>
    </UFormGroup>
  </UModal>
</template>
