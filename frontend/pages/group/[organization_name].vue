<script lang="ts" setup>
import { type Group } from "~/schema/schema";
import GroupInvitationsForm from "~/components/group-invitations-form.vue";
import GroupMembersForm from "~/components/group-members-form.vue";
import GroupProjects from "~/components/group-projects.vue";
const toast = useToast();

const route = useRoute();
const organization_name = route.params.organization_name;
const config = useRuntimeConfig();
const isAddMemberModalOpen = ref(false);

interface FoundUser {
  id: string;
  login: string;
}

interface Search {
  total_count: number;
  incomplete_results: boolean;
  items: FoundUser[];
}

const search = ref<Search>();

const memberID = ref("");

const tabItems = [
  {
    label: "Projects",
    __component: GroupProjects,
  },
  {
    label: "Members",
    __component: GroupMembersForm,
  },
  {
    label: "Invitations",
    __component: GroupInvitationsForm,
  },
  {
    label: "Monitoring (coming soon)",
    disabled: true,
  },
];

const g = ref<Group>();
const activeTabComponent = ref(tabItems[0].__component);

onMounted(() => {
  fetchOrganization();
});

function onChangeTab(idx: number) {
  activeTabComponent.value = tabItems[idx].__component;
}

async function fetchOrganization() {
  try {
    g.value = await $fetch<Group>(
      `${config.public.backendHost}/v1/organization/${organization_name}`,
      { credentials: "include" }
    );
  } catch (e) {
    console.error("unable to fetch Group", e);
  }
}

async function deleteMember(memberID: string) {
  try {
    g.value = await $fetch(
      `${config.public.backendHost}/v1/organization/member/${organization_name}/${memberID}`,
      {
        method: "DELETE",
        credentials: "include",
      }
    );
  } catch (e) {
    console.error("unable to delete member", e);
  } finally {
    fetchOrganization();
  }
}

async function inviteMember() {
  try {
    g.value = await $fetch(
      `${config.public.backendHost}/v1/organization/member`,
      {
        method: "PUT",
        credentials: "include",
        body: {
          organization_name: g.value?.organization_name,
          email: memberID.value,
        },
      }
    );
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Invitation has been sent",
        life: 3000
    });
  } catch (e) {
    console.error("unable to invite", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to invite",
    });
  } finally {
    isAddMemberModalOpen.value = false;
    fetchOrganization();
  }
}
</script>
<template>
  <div class="flex flex-col">
    <TabView :items="tabItems" class="mb-4">
      <TabPanel v-for="tab in tabItems" :key="tab.label" :header="tab.label" />
    </TabView>

    <div v-if="g && g.members" class="mt-6">
      <div class="p-6 flex flex-row items-end justify-between">
        <div>
          <h1 class="text-2xl">{{ g.organization_name }}</h1>
          <p class="text-sm text-gray-400">
            {{ g.members?.length }} Group members
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
    <component
      :is="activeTabComponent"
      :group="g"
      @delete-member="deleteMember"
    ></component>
  </div>
  <Dialog v-model:visible="isAddMemberModalOpen" header="Add Member" modal>
    <div class="flex flex-col space-y-4 p-6">
      <div class="flex flex-row items-center space-x-4">
        <input class="w-full" v-model="memberID" />
        <IconButton
          @click="inviteMember()"
          class="invite-button text-green-500 hover:text-green-700 focus:outline-none"
        />
      </div>
    </div>
  </Dialog>
</template>
