<script lang="ts" setup>
import { type Invitation, type Organisation } from "~/schema/schema";

const config = useRuntimeConfig();
const { data } = loadGroups();
const { data: invitationsData } = loadInvitations();

const confirm = useConfirm();

interface OrganisationState {
  isDeploying?: boolean;
  isRemoving?: boolean;
}
const state = ref<Record<string, OrganisationState>>({});
const toast = useToast();

function loadGroups() {
  return useFetch<Organisation[]>(`${config.public.backendHost}/v1/organisations`, {
    server: false,
    lazy: true,
    credentials: "include",
  });
}

function loadInvitations() {
  return useFetch<Invitation[]>(
    `${config.public.backendHost}/v1/organisations/invitations`,
    {
      server: false,
      lazy: true,
      credentials: "include",
    }
  );
}

function remove(name: string) {
  state.value[name] = {
    isRemoving: true,
  };
  $fetch(`${config.public.backendHost}/v1/organisation/${name}`, {
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
        detail: "Failed to delete organisation",
      });
    })
    .finally(() => (state.value[name].isRemoving = false));
}

function confirmRemove(name: string) {
  confirm.require({
    header: "Remove Group?",
    message: "Do you wanna delete this organisation, this can not be undone?",
    accept: () => remove(name),
    acceptLabel: "Accept",
    rejectLabel: "Cancel",
  });
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
        <h1 class="text-2xl">Organisation</h1>
        <p class="text-sm text-gray-400">
          Member of {{ data?.length ?? 0 }} organisation
        </p>
      </div>
      <NuxtLink to="/organisation/new">
        <IconButton
          label="Create Organisation"
          icon="heroicons:rocket-launch"
          aria-label="create"
        />
      </NuxtLink>
    </div>

    <div>
      <div
        v-for="d in data as Organisation[]"
        class="p-6 flex flex-row flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700"
      >
        <div class="flex flex-row items-center">
          <Avatar :alt="d.organisation_name" class="mr-3" />
          <div class="w-2/3">
            <p>{{ d.organisation_name }}</p>
          </div>
        </div>
        <div class="space-x-4 flex flex-row items-center">
          <IconButton
            icon="heroicons:trash"
            :loading="state[d.organisation_name]?.isRemoving"
            text
            severity="danger"
            @click="() => confirmRemove(d.organisation_name)"
          />
          <NuxtLink :to="'organisation/' + d.organisation_name">
            <IconButton icon="heroicons:arrow-right-on-rectangle" />
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>
