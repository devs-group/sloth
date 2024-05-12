<template>
  <div class="flex flex-col">
    <div v-if="organisation?.members">
      <div class="flex flex-row items-end justify-between">
        <div>
          <p class="text-sm text-gray-400">
            {{ organisation.members?.length }} Organisation members
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

<script lang="ts" setup>
import { type Organisation } from "~/schema/schema";
import OrganisationInvitationsForm from "~/components/organisation-invitations-form.vue";
import OrganisationMembersForm from "~/components/organisation-members-form.vue";
import OrganisationProjects from "~/components/organisation-projects.vue";
import {Constants} from "~/config/const";

const toast = useToast();
const route = useRoute();
const config = useRuntimeConfig();

const organisationId = route.params.id;
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
    component: OrganisationProjects,
  },
  {
    label: "Members",
    component: OrganisationMembersForm,
  },
  {
    label: "Invitations",
    component: OrganisationInvitationsForm,
  },
  {
    label: "Monitoring (coming soon)",
    disabled: true,
  },
] as TabItem[];

const organisation = ref<Organisation>();
const activeTabComponent = ref(tabItems[0].component);

onMounted(() => {
  fetchOrganisation();
});

function onChangeTab(idx: number) {
  activeTabComponent.value = tabItems[idx].component;
}

async function fetchOrganisation() {
  try {
    organisation.value = await $fetch<Organisation>(
        `${config.public.backendHost}/v1/organisation/${organisationId}`,
        { credentials: "include" }
    );
  } catch (e) {
    console.error("unable to fetch Organisation", e);
  }
}

async function deleteMember(memberID: string) {
  try {
    organisation.value = await $fetch(
        `${config.public.backendHost}/v1/organisation/member/${organisationId}/${memberID}`,
        {
          method: "DELETE",
          credentials: "include",
        }
    );
  } catch (e) {
    console.error("unable to delete member", e);
  } finally {
    fetchOrganisation();
  }
}

async function inviteMember() {
  try {
    organisation.value = await $fetch(
        `${config.public.backendHost}/v1/organisation/member`,
        {
          method: "PUT",
          credentials: "include",
          body: {
            organisation_name: organisation.value?.organisation_name,
            email: memberID.value,
          },
        }
    );
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Invitation has been sent",
      life: Constants.ToasterDefaultLifeTime,
    });
  } catch (e) {
    console.error("unable to invite", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to invite",
      life: Constants.ToasterDefaultLifeTime,
    });
  } finally {
    isAddMemberModalOpen.value = false;
    fetchOrganisation();
  }
}
</script>