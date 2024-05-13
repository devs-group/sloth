<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import GroupInvitationsForm from "~/components/organisation/organisation-invitations-form.vue";
import GroupMembersForm from "~/components/organisation/organisation-members-form.vue";
import GroupProjects from "~/components/organisation/organisation-projects.vue";

// Basic configuration and state setup
const route = useRoute();
const toast = useToast();
const config = useRuntimeConfig();
const organisationName = route.params.organisation_name;
const isAddMemberModalOpen = ref(false);
const memberID = ref("");

// Composable for organisation management
import { useOrganisation } from '~/composables/useOrganisation';
const { organisation, fetchOrganisation, deleteMember, inviteMember } = useOrganisation(organisationName, config, toast);

const tabItems = [
  { label: "Projects", __component: GroupProjects },
  { label: "Members", __component: GroupMembersForm },
  { label: "Invitations", __component: GroupInvitationsForm },
  { label: "Monitoring (coming soon)", disabled: true },
];
const activeTabComponent = ref(tabItems[0].__component);

</script>

<template>
  <div class="flex flex-col">
    <TabView :items="tabItems" class="mb-4">
      <TabPanel v-for="tab in tabItems" :key="tab.label" :header="tab.label" />
    </TabView>

    <div v-if="organisation && organisation.members" class="mt-6">
      <div class="p-6 flex flex-row items-end justify-between">
        <div>
          <h1 class="text-2xl">{{ organisation.organisation_name }}</h1>
          <p class="text-sm text-gray-400">
            {{ organisation.members.length }} Organisation members
          </p>
        </div>
        <IconButton
          icon="heroicons:pencil-square"
          variant="solid"
          :trailing="false"
          @click="isAddMemberModalOpen = true"
          label="Invite Member"
        />
      </div>
    </div>
    <component :is="activeTabComponent" :organisation="organisation" @delete-member="deleteMember"></component>
  </div>
  <Dialog v-model:visible="isAddMemberModalOpen" header="Add Member" modal>
    <div class="flex flex-col space-y-4 p-6">
      <div class="flex flex-row items-center space-x-4">
        <input class="w-full" v-model="memberID" />
        <IconButton
          @click="inviteMember"
          class="invite-button text-green-500 hover:text-green-700 focus:outline-none"
        />
      </div>
    </div>
  </Dialog>
</template>
