<template>
  <div
    class="hidden lg:flex flex-col border border-gray-200 dark:border-gray-700 border-t-0 border-b-0 relative pt-5 px-2"
  >
    <template v-for="link in menuItems">
      <hr v-if="link.divider" class="my-2 dark:border-gray-800" />
      <NuxtLink
        class="menu-item"
        v-else-if="link.to"
        :to="{ name: link.to }"
        v-tooltip="link.label"
      >
        <Button text severity="secondary" class="flex gap-2 items-center">
          <Icon v-if="link.icon" :icon="link.icon" />
        </Button>
      </NuxtLink>
      <Button v-else text severity="secondary" @click="link.click">
        <Icon v-if="link.icon" :icon="link.icon" />
      </Button>
    </template>
    <div v-if="user" class="flex justify-center gap-2 pb-2">
      <Avatar
        :image="user.avatar_url"
        shape="circle"
        class="p-1"
        v-tooltip="user.email || user.nickname"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Constants, DialogProps } from "~/config/const";
import CustomConfirmationDialog from "~/components/dialogs/custom-confirmation-dialog.vue";
import type {
  ICustomConfirmDialog,
  NavigationItems,
} from "~/config/interfaces";

const { user, logout } = useAuth();
const { getMainMenuItems } = useMenu();
const { showGlobalSpinner, hideGlobalSpinner } = useGlobalSpinner();
const confirm = useConfirm();
const toast = useToast();
const dialog = useDialog();

const menuItems = ref<NavigationItems[]>(
  getMainMenuItems({
    onLogout: () => {
      dialog.open(CustomConfirmationDialog, {
        props: {
          header: "Logout",
          ...DialogProps.SmallDialog,
        },
        data: {
          question: `Are you sure you want to logout?`,
          confirmText: "Logout",
          rejectText: "Cancel",
        } as ICustomConfirmDialog,
        onClose(options) {
          if (options?.data === true) {
            showGlobalSpinner();
            logout()
              .then(() => {
                reloadNuxtApp({ force: true });
              })
              .catch(() => {
                hideGlobalSpinner();
                toast.add({
                  severity: "error",
                  summary: "Error",
                  detail: "Unable to log out user",
                  life: Constants.ToasterDefaultLifeTime,
                });
              });
          }
        },
      });
    },
  })
);

const { checkInvitation, acceptInvitation } = useOrganisationInviation(toast);

onMounted(() => {
  const inviations = checkInvitation();
  if (inviations) {
    confirm.require({
      header: "Accept invitation?",
      message:
        "You were invited to a new Organisation do you wanna participate to the Organisation?",
      accept: () => acceptInvitation(inviations),
      acceptLabel: "Accept",
      rejectLabel: "Cancel",
    });
  }
});
</script>
