<script lang="ts" setup>
import { type Invitation, type Group } from "~/schema/schema";

const config = useRuntimeConfig();
const { data } = loadGroups();
const { data: invitationsData } = loadInvitations();

const confirm = useConfirm();

interface GroupState {
  isDeploying?: boolean;
  isRemoving?: boolean;
}
const state = ref<Record<string, GroupState>>({});
const toast = useToast();

function loadGroups() {
  return useFetch<Group[]>(`${config.public.backendHost}/v1/groups`, {
    server: false,
    lazy: true,
    credentials: "include",
  });
}

function loadInvitations() {
  return useFetch<Invitation[]>(
    `${config.public.backendHost}/v1/groups/invitations`,
    {
      server: false,
      lazy: true,
      credentials: "include",
    }
  );
}

function confirmRemove(group_name: string) {
  confirm.require({
    header: "Remove Group?",
    message: "Do you wanna delete this group, this can not be undone?",
    accept: () => remove(group_name),
    acceptLabel: "Accept",
    rejectLabel: "Cancel",
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
      toast.add({
        severity: "success",
        summary: "Success",
        detail: "Group has been removed successfully",
      });
    })
    .catch((e) => {
      console.error(e);
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Failed to delete group",
      });
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
          {{ invitationsData?.length ?? 0 }} Invitations
        </p>
      </div>
    </div>
    <div class="p-6 flex flex-row items-end justify-between">
      <div>
        <h1 class="text-2xl">Groups</h1>
        <p class="text-sm text-gray-400">
          Member of {{ data?.length ?? 0 }} Groups
        </p>
      </div>
      <NuxtLink to="/group/new">
        <IconButton
          label="Create Group"
          icon="heroicons:rocket-launch"
          aria-label="create"
        />
      </NuxtLink>
    </div>

    <div>
      <div
        v-for="d in data as Group[]"
        class="p-6 flex flex-row flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700"
      >
        <div class="flex flex-row items-center">
          <Avatar :alt="d.group_name" class="mr-3" />
          <div class="w-2/3">
            <p>{{ d.group_name }}</p>
          </div>
        </div>
        <div class="space-x-4 flex flex-row items-center">
          <IconButton
            icon="heroicons:trash"
            :loading="state[d.group_name]?.isRemoving"
            text
            severity="danger"
            @click="() => confirmRemove(d.group_name)"
          />
          <NuxtLink :to="'group/' + d.group_name">
            <IconButton icon="heroicons:arrow-right-on-rectangle" />
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>
