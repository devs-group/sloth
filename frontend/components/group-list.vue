<script lang="ts" setup>
import { Group } from "~/schema/schema";

const config = useRuntimeConfig();
const { data } = loadGroups();
const { showConfirmation } = useConfirmation();

interface GroupState {
  isDeploying?: boolean;
  isRemoving?: boolean;
}
const state = ref<Record<string, GroupState>>({});
const { showError, showSuccess } = useNotification();

function loadGroups() {
  return useFetch<Group[]>(`${config.public.backendHost}/v1/groups`, {
    server: false,
    lazy: true,
    credentials: "include",
  });
}

function remove(group_name: string) {
  state.value[group_name] = {
    isRemoving: true,
  };
  $fetch(`${config.public.backendHost}/v1/group/${group_name}`, {
    method: "DELETE",
    credentials: "include",
  })
    .then(() => {
      // Re-fetch projects after delete
      const { data: d } = loadGroups();
      data.value = d.value;

      showSuccess("Success", "Group has been removed successfully");
    })
    .catch((e) => {
      console.error(e);
      showError("Error", "Failed to delete group");
    })
    .finally(() => (state.value[group_name].isRemoving = false));
}
</script>

<template>
  <div>
    <div class="p-6 flex flex-row items-end justify-between">
      <div>
        <h1 class="text-2xl">Groups</h1>
        <p class="text-sm text-gray-400">
          Member of {{ data?.length ?? 0 }} Groups
        </p>
      </div>
      <UButton
        icon="i-heroicons-pencil-square"
        size="sm"
        color="gray"
        variant="solid"
        :trailing="false"
      >
        <NuxtLink to="/group/new">New Group</NuxtLink>
      </UButton>
    </div>

    <div>
      <div
        v-for="d in data as Group[]"
        class="p-6 flex flex-row flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700"
      >
        <div class="flex flex-row items-center">
          <UAvatar :alt="d.group_name" size="sm" class="mr-3" />
          <div class="w-2/3">
            <p>{{ d.group_name }}</p>
            <div class="relative">
              <UPopover class="mt-2">
                <UButton
                  color="white"
                  :label="`${d.members?.length ?? 0} members`"
                  trailing-icon="i-heroicons-chevron-down-20-solid"
                />
              </UPopover>
            </div>
          </div>
        </div>
        <div class="space-x-4 flex flex-row items-center">
          <UButton
            icon="i-heroicons-trash"
            :loading="state[d.group_name]?.isRemoving"
            variant="ghost"
            color="red"
            @click="() => showConfirmation(
                        'Remove the Group?',
                        'If you remove this group any related information and projects will be removed too and you won\'t be able to restore it.',
                         () => remove( d.group_name as string)
                        )
                      "
          >
          </UButton>
          <NuxtLink :to="'project/' + d.group_name">
            <UButton icon="i-heroicons-arrow-right-on-rectangle"></UButton>
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>
