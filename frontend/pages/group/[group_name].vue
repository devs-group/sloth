<script lang="ts" setup>
import { Group, groupSchema } from "~/schema/schema";
import DockerCredentialsForm from "~/components/docker-credentials-form.vue";
import ServicesForm from "~/components/services-form.vue";

const { showConfirmation } = useConfirmation();
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
interface State {
  isRemoving?: boolean;
}
const state = ref<Record<string, State>>({});

const tabItems = [
  {
    label: "Services",
    __component: ServicesForm,
  },
  {
    label: "Docker credentials",
    __component: DockerCredentialsForm,
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
      `${config.public.backendHost}/v1/group/${group_name}/${memberID}`,
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

async function lookupUser(memberID: string) {
  try {
    search.value = await $fetch(
      `${config.public.backendHost}/v1/group/${group_name}/${memberID}`,
      {
        method: "GET",
        credentials: "include",
      }
    );
  } catch (e) {
    console.error("unable to add member", e);
  }
}

async function inviteMember(memberID: string) {
  try {
    g.value = await $fetch(
      `${config.public.backendHost}/v1/group/${group_name}/${memberID}`,
      {
        method: "PUT",
        credentials: "include",
      }
    );
  } catch (e) {
    console.error("unable to add member", e);
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
      <ul class="list-disc pl-5">
        <li
          v-for="member in g.members"
          :key="member"
          class="flex justify-between items-center mb-2 pl-5"
        >
          <span class="text-gray-800">{{ member }}</span>
          <UButton
            icon="i-heroicons-trash"
            :loading="state[member]?.isRemoving"
            variant="ghost"
            color="red"
            @click="
              () =>
                showConfirmation(
                  'Remove the project?',
                  'Are you sure you wanna remove this user from your group?',
                  () => deleteMember(member)
                )
            "
          >
          </UButton>
        </li>
      </ul>
    </div>
  </div>
  <UModal v-model="isAddMemberModalOpen">
    <UFormGroup name="name">
      <div class="flex flex-col space-y-4 p-6">
        <div class="flex flex-row items-center space-x-4">
          <UInput
            class="w-full"
            v-model="memberID"
            @input="lookupUser(memberID)"
          />
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
              <UButton
                @click="inviteMember(user.id)"
                class="invite-button text-green-500 hover:text-green-700 focus:outline-none"
              >
                <i class="icon heroicons-check text-lg"></i>
              </UButton>
            </li>
          </ul>
        </div>
      </div>
    </UFormGroup>
  </UModal>
</template>
