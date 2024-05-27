<template>
  <div class="flex flex-col gap-2 w-full">
    <TabMenu :model="tabItems" class="w-full" @change="onChangeTab"/>
    <div class="flex flex-row w-full justify-between items-center pl-5 pr-5">
      <div class="flex flex-col">
        <p class="text-lg">{{ organisation?.organisation_name }}</p>
      </div> 
      <div class="flex flex-col">
        <IconButton
              label="Add Project"
              icon="heroicons:rocket-launch"
              aria-label="add"
              @click="onInviteToOrganisation()"
          />
      </div>
    </div>
    <div class="flex flex-col gap-2 px">
      <component v-if="!isLoading" :is="activeTabComponent" :props="activeTabProps" />
    </div>
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
import type { Invitation } from "~/schema/schema";
import InviteToOrganisationDialog from "~/components/dialogs/invite-to-organisation-dialog.vue";

const toast = useToast();
const isLoading = ref(true);
const route = useRoute();
const dialog = useDialog();
const organisationID = parseInt(route.params.id.toString());
const { organisation, organisationProjects, fetchOrganisation, fetchOrganisationProjects  } = useOrganisation(organisationID, toast);
const {  } = useOrganisationInviation()

const tabItems = computed(() => [
  { 
    label: "Projects", 
    component: OrganisationProjectList, 
    props: {  organisation: organisation.value, projects: organisationProjects.value }, 
    command: () => onChangeTab(0)
  },
  { label: "Members", component: OrganisationMembers, props: { organisation: organisation.value}, command: () => onChangeTab(1) },
  { label: "Invitations", component: OrganisationInvitationsForm, props: { organisation: organisation.value, isLoading: false }, command: () => onChangeTab(2) },
  { label: "Monitoring (coming soon)", disabled: true },
] as TabItem[]);
const { activeTabComponent, onChangeTab, activeTabProps } = useTabs(tabItems);

onMounted(async () => {
  isLoading.value = true;
  await fetchOrganisation();
  await fetchOrganisationProjects(organisationID);
  isLoading.value = false; 
});

const invitaions: Invitation[] = [
  {
    organisation_name: 'Google',
    user_id: "2"
  },
  {
    organisation_name: 'Apple',
    user_id: "1"
  },
  {
    organisation_name: 'Microsoft',
    user_id: "3"
  },
  {
    organisation_name: 'Samsung',
    user_id: "4"
  },
  {
    organisation_name: 'Nintendo',
    user_id: "5"
  },
  {
    organisation_name: 'Dell',
    user_id: "6"
  },
  {
    organisation_name: 'Sony',
    user_id: "7"
  }
]

const onAddProjectToOrganisation = () => {
  console.log(organisationID)
  dialog.open(AddProjectToOrganisationDialog, {
    props: {
      header: 'Add Project to Organisation',
      ...DialogProps.BigDialog,
    },
    data: {
      organisation_id: organisationID
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
    }
  })
}

</script>