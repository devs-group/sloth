<template>
  <NuxtLink
    :to="{
      name: Routes.ORGANISATION,
      params: { id: organisation.id },
    }"
    class="p-6 flex flex-1 items-center justify-between border border-1 rounded-xl border-neutral-200 dark:border-neutral-700 hover:bg-gray-50 dark:hover:bg-neutral-800"
  >
    <div class="flex items-center w-full gap-2">
      <div class="flex flex-col w-full">
        <p>{{ props.organisation.organisationName }}</p>
        <p>Members: {{ props.organisation.members?.length ?? 0 }}</p>
        <p>Your Role: {{ props.organisation.currentRole }}</p>
      </div>
      <Message
        v-if="isActiveOrganisation(props.organisation)"
        severity="success"
        class="p-4"
      >
        Active
      </Message>
      <Button
        v-else
        severity="secondary"
        label="Switch to organisation"
        @click="onSwitchOrganisation(props.organisation)"
      />
    </div>
  </NuxtLink>
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
  return organisation.id == user.value?.currentOrganisationID
}

function openDeleteOrganisationDialog(organisation: Organisation) {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: 'Delete organisation',
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Do you want to delete "${organisation.organisationName}"? This action cannot be undone.`,
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
