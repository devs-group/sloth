<template>
  <div class="flex flex-col gap-2 w-full">
    <TabMenu :model="tabItems" class="w-full" @change="onChangeTab"/>
    <component v-if="!isLoading" :is="activeTabComponent" :props="activeTabProps" />
    <OverlayProgressSpinner :show="isLoading"/>
  </div>
</template>

<script setup lang="ts">
import type { TabItem } from "~/config/interfaces";
import { DialogProps } from "~/config/const";
import OrganisationInvitationsForm from "~/components/organisation-invitations-form.vue";
import OrganisationMembers from "~/components/organisation-members-form.vue";
import OrganisationProjectList from "~/components/organisation-project-list.vue";
import AddProjectToOrganisationDialog from "~/components/dialogs/add-project-to-organisation-dialog.vue";
import AddMemberToOrganisationDialog from "~/components/dialogs/add-member-to-organisation-dialog.vue";
import InviteToOrganisationDialog from "~/components/dialogs/invite-to-organisation-dialog.vue";

const toast = useToast();
const isLoading = ref(true);
const route = useRoute();
const dialog = useDialog();
const organisationID = parseInt(route.params.id.toString());
const { organisation, organisationProjects, fetchOrganisation, fetchOrganisationProjects  } = useOrganisation(organisationID, toast);

const tabItems = computed(() => [
  { 
    label: "Projects", 
    component: OrganisationProjectList, 
    props: {  organisation: organisation.value, projects: organisationProjects.value, button: { label: "Add Project", icon: "heroicons:rocket-launch", onClick: () => onAddProjectToOrganisation() } }, 
    command: () => onChangeTab(0)
  },
  { label: "Members",
    component: OrganisationMembers,
    props: { organisation: organisation.value, button: { label: "Add Member", icon: "heroicons:user-group", onClick: () => onAddMemberToOrganisation() }},
    command: () => onChangeTab(1) },
    { label: "Invitations", component: OrganisationInvitationsForm, props: { organisation: organisation.value, isLoading: false, invitations: invitations.value }, command: () => onChangeTab(2) },
  { label: "Monitoring (coming soon)", disabled: true },
] as TabItem[]);

const { activeTabComponent, onChangeTab, activeTabProps } = useTabs(tabItems);
const { invitations, loadInvitations } = useOrganisations(toast)

onMounted(async () => {
  isLoading.value = true;
  await fetchOrganisation();
  await loadOrganisationProjects()
});

const loadOrganisationProjects = async () => {
  await fetchOrganisationProjects(organisationID);
  await loadInvitations(organisationID)
  isLoading.value = false; 
}

const onAddProjectToOrganisation = () => {
  console.log(organisationID)
  dialog.open(AddProjectToOrganisationDialog, {
    props: {
      header: 'Add Project to Organisation',
      ...DialogProps.BigDialog,
    },
    data: {
      organisation_id: organisationID,
      organisationProjects: organisationProjects.value
    },
    onClose: async () => {
      await loadOrganisationProjects()
    }
  })
}

const onAddMemberToOrganisation = () => {
  console.log(organisationID)
  dialog.open(AddMemberToOrganisationDialog, {
    props: {
      header: 'Add Member to Organisation',
      ...DialogProps.BigDialog,
    },
    data: {
      organisation_id: organisationID
    },
    onClose: async () => {
      await loadOrganisationProjects()
    }
  })
}
const onInviteToOrganisation = () => {
  console.log(organisationID)
  dialog.open(InviteToOrganisationDialog, {
    props: {
      header: 'Invite to Organisation',
      ...DialogProps.BigDialog,
    },
    data: {
      organisation_name: organisation.value?.organisation_name
    },
    onClose: async () => {
      await loadInvitations(organisationID)
    }
  })
}
</script>