<template>
  <div class="flex flex-col gap-6 py-6 w-full max-w-6xl">
    <div class="flex flex-col gap-2 px-6">
      <h1 class="text-2xl">Organisations</h1>
      <p class="text-sm text-gray-400">
        {{ invitationsData?.length ?? 0 }} open invitations
      </p>
    </div>
    <div class="flex items-end justify-between px-6">
      <div>
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
          v-for="organisation in organisations"
          class="p-6 flex flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700"
      >
        <div class="flex items-center">
          <Avatar :alt="organisation.organisation_name" class="mr-3"/>
          <div class="w-2/3">
            <p>{{ organisation.organisation_name }}</p>
          </div>
        </div>
        <div class="space-x-4 flex items-center">
          <IconButton
              icon="heroicons:trash"
              :loading="state[organisation.organisation_name]?.isRemoving"
              text
              severity="danger"
              @click="onDeleteOrganisation(organisation)"
          />
          <NuxtLink :to="{name: Routes.ORGANISATION, params: {id: organisation.id}}">
            <IconButton icon="heroicons:arrow-right-on-rectangle"/>
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import {type Invitation, type Organisation} from "~/schema/schema";
import {Routes} from "~/config/routes";
import {Constants} from "~/config/const";

const config = useRuntimeConfig();
const {data: organisations} = loadOrganisations();
const {data: invitationsData} = loadInvitations();

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

function onDeleteOrganisation(organisation: Organisation) {
  confirm.require({
    header: "Remove Organisation?",
    message: `Do you want to delete "${organisation.organisation_name}", this can not be undone?`,
    accept: () => deleteOrganisation(organisation),
    acceptLabel: "Accept",
    rejectLabel: "Cancel",
  });
}

function deleteOrganisation(organisation: Organisation) {
  state.value[organisation.id] = {
    isRemoving: true,
  };
  $fetch(`${config.public.backendHost}/v1/organisation/${organisation.id}`, {
    method: "DELETE",
    credentials: "include",
  })
      .then(() => {
        // Re-fetch projects after delete
        const {data: d} = loadOrganisations();
        organisations.value = d.value;
        toast.add({
          severity: "success",
          summary: "Success",
          detail: `Organisation "${organisation.organisation_name}" has been removed successfully`,
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
      .finally(() => (state.value[organisation.id].isRemoving = false));
}
</script>