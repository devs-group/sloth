<script lang="ts" setup>
import { Group, groupSchema } from "~/schema/schema";
import GroupInvitationsForm from "~/components/group-invitations-form.vue";
import GroupMembersForm from "~/components/group-members-form.vue";
import GroupProjects from "~/components/group-projects.vue";
const { showError, showSuccess } = useNotification();

const route = useRoute();
const group_name = route.params.group_name;
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
  fetchGroup();
});

function onChangeTab(idx: number) {
  activeTabComponent.value = tabItems[idx].__component;
}

async function fetchGroup() {
  try {
    g.value = await $fetch<Group>(
      `${config.public.backendHost}/v1/group/${group_name}`,
      { credentials: "include" }
    );
  } catch (e) {
    console.error("unable to fetch Group", e);
  }
}

async function deleteMember(memberID: string) {
  try {
    g.value = await $fetch(
      `${config.public.backendHost}/v1/group/member/${group_name}/${memberID}`,
      {
        method: "DELETE",
        credentials: "include",
      }
    );
  } catch (e) {
    console.error("unable to delete member", e);
  } finally {
    fetchGroup();
  }
}

async function inviteMember(memberID: string) {
  try {
    g.value = await $fetch(`${config.public.backendHost}/v1/group/member`, {
      method: "PUT",
      credentials: "include",
      body: {
        group_name: g.value?.group_name,
        email: memberID,
      },
    });
    showSuccess("Success", "Invitation has been sent");
  } catch (e) {
    console.error("unable to invite", e);
    showError("Error", "Uunable to invite");
  } finally {
    isAddMemberModalOpen.value = false;
    fetchGroup();
  }
}
</script>
<template>
  <div class="flex flex-col">
    <UTabs :items="tabItems" @change="onChangeTab" class="mb-4" />

    <div v-if="g && g.members" class="mt-6">
      <div class="p-6 flex flex-row items-end justify-between">
        <div>
          <h1 class="text-2xl">{{ g.group_name }}</h1>
          <p class="text-sm text-gray-400">
            {{ g.members?.length }} Group members
          </p>
        </div>
        <UButton
          icon="i-heroicons-pencil-square"
          size="sm"
          color="gray"
          variant="solid"
          :trailing="false"
          @click="isAddMemberModalOpen = true"
        >
          Invite Member
        </UButton>
      </div>
    </div>
    <component
      :is="activeTabComponent"
      :group="g"
      @delete-member="deleteMember"
    ></component>
  </div>
  <UModal v-model="isAddMemberModalOpen">
    <UFormGroup name="name">
      <div class="flex flex-col space-y-4 p-6">
        <div class="flex flex-row items-center space-x-4">
          <UInput class="w-full" v-model="memberID" />
          <UButton
            @click="inviteMember(memberID)"
            class="invite-button text-green-500 hover:text-green-700 focus:outline-none"
          >
            <i class="icon heroicons-check text-lg"></i>
          </UButton>
        </div>
        <!-- User List Display -->
        <div v-if="search?.items.length" class="search-results space-y-2">
          <ul class="list-disc pl-4 overflow-auto h-64">
            <li
              v-for="user in search?.items"
              :key="user.id"
              class="item bg-gray-100 hover:bg-gray-200 py-2 px-4 rounded-md flex justify-between items-center"
            >
              <span class="username text-sm font-medium">{{ user.login }}</span>
            </li>
          </ul>
        </div>
      </div>
    </UFormGroup>
  </UModal>
</template>
