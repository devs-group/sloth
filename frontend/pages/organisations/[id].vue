<template>
  <div class="flex flex-col gap-2 w-full">
    <TabMenu :model="tabItems" class="w-full"/>
    <div class="flex flex-col gap-2 px-4">
      <p class="text-lg">{{ organisation?.organisation_name }}</p>
      <NuxtPage/>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TabItem } from "~/config/interfaces";
import GroupInvitationsForm from "~/components/organisation-invitations-form.vue";
import GroupMembersForm from "~/components/organisation-members-form.vue";
import GroupProjects from "~/components/organisation-projects.vue";

const route = useRoute();
const organisationID = parseInt((route.params.id.length > 0  ? route.params.id[0] : "0"), 10 );
const { organisation, fetchOrganisation, fetchOrganisationProjects  } = useOrganisation(organisationID)

const tabItems = [
  { label: "Projects", component: GroupProjects, props: { organisation } },
  { label: "Members", component: GroupMembersForm, props: { organisation } },
  { label: "Invitations", component: GroupInvitationsForm, props: { organisation } },
  { label: "Monitoring (coming soon)", disabled: true },
] as TabItem[];

onMounted(() => {
  fetchOrganisation()
  fetchOrganisationProjects(organisationID)
});
</script>