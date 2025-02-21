<template>
  <div
    class="p-6 flex flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700"
  >
    <div class="flex items-center w-full gap-2">
      <Avatar :alt="props.organisation.organisation_name" />
      <div class="w-full">
        <p>{{ props.organisation.organisation_name }}</p>
      </div>
      <p
        v-if="isActiveOrganisation(props.organisation)"
        class="bg-green-200 p-4"
      >
        Active
      </p>
      <Button
        v-else
        severity="secondary"
        label="Switch to organisation"
        @click="onSwitchOrganisation(props.organisation)"
      />
    </div>
    <div
      v-if="isActiveOrganisation(props.organisation)"
      class="space-x-4 flex items-center"
    >
      <IconButton
        icon="heroicons:trash"
        :loading="isDeletingOrganisation"
        text
        severity="danger"
        @click="openDeleteOrganisationDialog(props.organisation)"
      />
      <NuxtLink
        :to="{
          name: Routes.ORGANISATION,
          params: { id: organisation.id },
        }"
      >
        <IconButton icon="heroicons:arrow-right-on-rectangle" />
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Organisation } from '~/schema/schema'
import { Routes } from '~/config/routes'
import { Constants, DialogProps } from '~/config/const'
import CustomConfirmationDialog from '~/components/dialogs/custom-confirmation-dialog.vue'
import type { ICustomConfirmDialog, IPutMemberToOrganisationResponse } from '~/config/interfaces'
import { APIService } from '~/api'

const props = defineProps({
  organisation: {
    type: Object as PropType<Organisation>,
    required: true,
  },
})

const emits = defineEmits<{ (event: 'onDelete', id: number): void }>()

const { user } = useAuth()
const config = useRuntimeConfig()
const toast = useToast()

const dialog = useDialog()
const { isLoading: isDeletingOrganisation, execute: deleteOrganisation }
  = useApi(
    (organisationID: number) => APIService.DELETE_organisation(organisationID),
    {
      showSuccessToast: true,
      successMessage: 'Organisation has been deleted succesfully',
    },
  )

const isActiveOrganisation = (organisation: Organisation) => {
  return organisation.id == user.value?.current_organisation_id
}

function openDeleteOrganisationDialog(organisation: Organisation) {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: 'Delete organisation',
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Do you want to delete "${organisation.organisation_name}"? This action cannot be undone.`,
      confirmText: 'Delete',
      rejectText: 'Cancel',
    } as ICustomConfirmDialog,
    async onClose(options) {
      if (options?.data === true) {
        await deleteOrganisation(organisation.id)
        emits('onDelete', organisation.id)
      }
    },
  })
}

const onSwitchOrganisation = (organisation: Organisation) => {
  $fetch<IPutMemberToOrganisationResponse>(`${config.public.backendHost}/v1/user/set-current-organisation`, {
    method: 'PUT',
    body: {
      id: organisation.id,
    },
    credentials: 'include',
  })
    .then(async () => {
      reloadNuxtApp({ force: true })
    })
    .catch(() => {
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: 'Can\'t switch organisation',
        life: Constants.ToasterDefaultLifeTime,
      })
    })
}
</script>
