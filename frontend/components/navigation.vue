<template>
  <div
    class="hidden w-64 lg:flex flex-col border border-gray-200 dark:border-gray-700 border-t-0 border-b-0 relative pt-5 px-2"
  >
    <template
      v-for="link in menuItems"
      :key="link.label"
    >
      <hr
        v-if="link.divider"
        class="my-2 dark:border-gray-800"
      >
      <NuxtLink
        v-else-if="link.to"
        class="menu-item"
        :to="{ name: link.to }"
      >
        <Button
          text
          severity="secondary"
          class="w-full"
        >
          <div class="flex gap-2 items-center w-full">
            <Icon
              v-if="link.icon"
              :icon="link.icon"
            />
            <p>{{ link.label }}</p>
          </div>
        </Button>
      </NuxtLink>
      <Button
        v-else
        text
        severity="secondary"
        class="w-full"
        @click="link.click"
      >
        <div class="flex gap-2 items-center w-full">
          <Icon
            v-if="link.icon"
            :icon="link.icon"
          />
          <p>{{ link.label }}</p>
        </div>
      </Button>
    </template>
    <div
      v-if="user"
      class="flex items-center gap-2 cursor-default"
    >
      <Avatar
        :image="user.avatar_url"
        shape="circle"
        class="p-1"
      />
      <p class="text-xs dark:text-neutral-500">
        {{ user.email }}
      </p>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Icon } from '@iconify/vue'
import { Constants, DialogProps } from '~/config/const'
import CustomConfirmationDialog from '~/components/dialogs/custom-confirmation-dialog.vue'
import type { ICustomConfirmDialog, NavigationItems } from '~/config/interfaces'

const { user, logout } = useAuth()
const { getMainMenuItems } = useMenu()
const { showGlobalSpinner, hideGlobalSpinner } = useGlobalSpinner()
const confirm = useConfirm()
const toast = useToast()
const dialog = useDialog()

const menuItems = ref<NavigationItems[]>(
  getMainMenuItems({
    onLogout: () => {
      dialog.open(CustomConfirmationDialog, {
        props: {
          header: 'Logout',
          ...DialogProps.SmallDialog,
        },
        data: {
          question: `Are you sure you want to logout?`,
          confirmText: 'Logout',
          rejectText: 'Cancel',
        } as ICustomConfirmDialog,
        onClose(options) {
          if (options?.data === true) {
            showGlobalSpinner()
            logout()
              .then(() => {
                reloadNuxtApp({ force: true })
              })
              .catch(() => {
                hideGlobalSpinner()
                toast.add({
                  severity: 'error',
                  summary: 'Error',
                  detail: 'Unable to log out user',
                  life: Constants.ToasterDefaultLifeTime,
                })
              })
          }
        },
      })
    },
  }),
)

const { checkInvitation, acceptInvitation } = useOrganisationInviation(toast)

onMounted(() => {
  const invitations = checkInvitation()
  if (invitations) {
    confirm.require({
      header: 'Accept invitation?',
      message:
        'You were invited to a new Organisation do you wanna participate to the Organisation?',
      accept: () => acceptInvitation(invitations),
      acceptLabel: 'Accept',
      rejectLabel: 'Cancel',
    })
  }
})
</script>
