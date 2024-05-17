<template>
  <div class="flex flex-col gap-2 w-full">
    
    <TabMenu :model="tabItems" class="w-full" @change="onChangeTab"/>
    <div class="flex flex-col gap-2 px-4">
      <p class="text-lg">{{ organisation?.organisation_name }}</p>
      <IconButton
            label="Add Project"
            icon="heroicons:rocket-launch"
            aria-label="add"
            @click="onAddProjectToOrganisation()"
        />
    </div> 
    <div class="flex flex-col gap-2 px" >
      <component :is="activeTabComponent" :props="props"/>
    </div>
    
  </div>
</template>

<script setup lang="ts">
import type { TabItem } from "~/config/interfaces";
import type { OrganisationProject } from "~/schema/schema";
import { DialogProps } from "~/config/const";
import OrganisationInvitationsForm from "~/components/organisation-invitations-form.vue";
import OrganisationMembers from "~/components/organisation-members-form.vue";
import OrganisationProjects from "~/components/organisation-projects.vue";
import AddProjectToOrganisationDialog from "~/components/dialogs/add-project-to-organisation-dialog.vue";

const route = useRoute();
const dialog = useDialog();
const organisationID = parseInt((route.params.id.length > 0  ? route.params.id[0] : "0"), 10 );
const { organisation, fetchOrganisation, fetchOrganisationProjects, addProjectToOrganisation  } = useOrganisation(organisationID)
const organisationProjects = ref<OrganisationProject[]>();

const tabItems = [
  { label: "Projects", component: OrganisationProjects, props: { projects: organisationProjects }, command: () => onChangeTab(0)},
  { label: "Members", component: OrganisationMembers,props: { organisation: organisation }, command: () => onChangeTab(1) },
  { label: "Invitations", component: OrganisationInvitationsForm,props: { organisation: organisation }, command: () => onChangeTab(2) },
  { label: "Monitoring (coming soon)", disabled: true },
] as TabItem[];
const { activeTabComponent, onChangeTab, props } = useTabs(tabItems);

onMounted(async () => {
  await fetchOrganisation();
  organisationProjects.value = await fetchOrganisationProjects(organisationID);
});

const onAddProjectToOrganisation = () => {
  dialog.open(AddProjectToOrganisationDialog, {
    props: {
      header: 'Add Project to Organisation',
      ...DialogProps.BigDialog,
    },
  })
}

</script>