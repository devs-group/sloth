<template>
  <div>
    <div class="p-6 flex flex-row items-end justify-between">
      <div>
        <h1 class="text-2xl">Organisations</h1>
        <p class="text-sm text-gray-400">
          {{ invitationsData?.length ?? 0 }} Invitations
        </p>
      </div>
    </div>
    <div class="p-6 flex flex-row items-end justify-between">
      <div>
        <h1 class="text-2xl">Organisations</h1>
        <p class="text-sm text-gray-400">
          Member of {{ organisations?.length ?? 0 }} Organisations
        </p>
      </div>
      <NuxtLink :to="{name: Routes.ORGANISATIONS_NEW}">
        <IconButton
          label="Create Organisation"
          icon="heroicons:rocket-launch"
          aria-label="create"
        />
      </NuxtLink>
    </div>

    <div>
      <div
        v-for="d in organisations as Organisation[]"
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
          <NuxtLink :to="{name: Routes.ORGANISATIONS, params: {id: d.id}}">
            <IconButton icon="heroicons:arrow-right-on-rectangle" />
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { type Invitation, type Organisation } from "~/schema/schema";
import {Routes} from "~/config/routes";
import {Constants} from "~/config/const";

const config = useRuntimeConfig();
const { data: organisations } = loadOrganisations();
const { data: invitationsData } = loadInvitations();

const confirm = useConfirm();

interface OrganisationState {
  isDeploying?: boolean;
  isRemoving?: boolean;
}
const state = ref<Record<string, OrganisationState>>({});
const toast = useToast();

function loadOrganisations() {
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

function confirmRemove(organisation_name: string) {
  confirm.require({
    header: "Remove Organisation?",
    message: "Do you wanna delete this organisation, this can not be undone?",
    accept: () => remove(organisation_name),
    acceptLabel: "Accept",
    rejectLabel: "Cancel",
  });
}

function remove(organisation_name: string) {
  state.value[organisation_name] = {
    isRemoving: true,
  };
  $fetch(`${config.public.backendHost}/v1/organisation/${organisation_name}`, {
    method: "DELETE",
    credentials: "include",
  })
      .then(() => {
        // Re-fetch projects after delete
        const { data: d } = loadOrganisations();
        organisations.value = d.value;
        toast.add({
          severity: "success",
          summary: "Success",
          detail: "Organisation has been removed successfully",
          life: Constants.ToasterDefaultLifeTime,
        });
      })
      .catch((e) => {
        console.error(e);
        toast.add({
          severity: "error",
          summary: "Error",
          detail: "Failed to delete organisation",
          life: Constants.ToasterDefaultLifeTime,
        });
      })
      .finally(() => (state.value[organisation_name].isRemoving = false));
}
</script>