<template>
  <div class="flex justify-between items-center min-h-12 border  border-gray-200 dark:border-gray-700 px-6">
    <p class="font-bold">
      SLOTH
    </p>

    <Inbox />

    <div class="block lg:hidden">
      <IconButton
        icon="heroicons:bars-3"
        outlined
        aria-haspopup="true"
        aria-controls="overlay_menu"
        @click="onToggleMenu"
      />
      <Menu
        id="overlay_menu"
        ref="menu"
        :model="links"
        :popup="true"
      >
        <template #item="{ item }">
          <hr
            v-if="item.divider"
            class="my-2"
          >
          <NuxtLink
            v-else-if="item.to"
            class="menu-item"
            :to="{ name: item.to }"
          >
            <Button
              text
              severity="secondary"
              class="flex gap-2 items-center w-full"
            >
              <Icon
                v-if="item.icon"
                :icon="item.icon"
                style="font-size: 20px"
              />
              <span>{{ item.label }}</span>
            </Button>
          </NuxtLink>
          <Button
            v-else
            text
            severity="secondary"
            class="flex gap-2 items-center w-full"
            @click="item.click"
          >
            <Icon
              v-if="item.icon"
              :icon="item.icon"
              style="font-size: 20px"
            />
            <span>{{ item.label }}</span>
          </Button>
        </template>
      </Menu>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Constants, DialogProps } from '~/config/const'
import CustomConfirmationDialog from '~/components/dialogs/custom-confirmation-dialog.vue'
import type { ICustomConfirmDialog, NavigationItems } from '~/config/interfaces'

const dialog = useDialog()
const toast = useToast()
const { showGlobalSpinner, hideGlobalSpinner } = useGlobalSpinner()
const { logout, getMenuItems } = useAuth()

const menu = ref()
const links = ref<NavigationItems[]>(getMenuItems({
  onLogout: () => {
    // TODO: Mach weiter indem du confirm.require( ersetzt durch das hier drunter

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
}))

const onToggleMenu = (event: PointerEvent) => {
  menu.value.toggle(event)
}
</script>
