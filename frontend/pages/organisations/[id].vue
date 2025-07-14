<template>
  <WrappersListPage
    :title="currentOrganisation?.organisationName ?? '...'"
  >
    <template #actions>
      <IconButton
        v-if="currentOrganisation && canEditOrganisation()"
        v-tooltip="'Organisation Settings'"
        icon="heroicons:cog"
        severity="contrast"
        size="small"
        @click="onChangeOrganisationName"
      />
    </template>
    <template #content>
      <div
        v-if="currentOrganisation"
        class="flex flex-col gap-8"
      >
        <div>
          <div class="flex items-center gap-2">
            <h2 class="text-lg sm:text-xl">
              Members
            </h2>
            <IconButton
              v-if="canEditOrganisation()"
              v-tooltip="'Add Member'"
              icon="heroicons:plus"
              size="small"
              @click="onInviteToOrganisation"
            />
          </div>
          <div
            v-for="member of currentOrganisation.members"
            :key="member.userID"
            class="flex items-center gap-2 bg-gray-50 dark:bg-neutral-900 p-2"
          >
            <p class="flex-1 sm:flex-initial">
              {{ member.email }} {{ member.userID == user?.id ? '(You)' : '' }}
            </p>
            <IconButton
              v-if="canEditOrganisation()"
              v-tooltip="'Edit Member'"
              icon="heroicons:pencil"
              severity="contrast"
              size="small"
              @click="onEditOrganisationMember(member)"
            />
            <IconButton
              v-if="canEditOrganisation() && member.userID != user?.id"
              v-tooltip="'Remove Member'"
              icon="heroicons:user-minus"
              severity="danger"
              size="small"
              @click="onRemoveOrganisationMember(member)"
            />
            <IconButton
              v-if="!isOwnerOfOrganisation() && member.userID == user?.id"
              v-tooltip="'Leave Organisation'"
              icon="heroicons:arrow-right-end-on-rectangle"
              severity="danger"
              size="small"
              @click="onLeaveOrganisation(member)"
            />
          </div>
        </div>

        <div v-if="invitations.length > 0">
          <h3 class="text-lg sm:text-xl">
            Pending Invitations
          </h3>
          <div
            v-for="invitation of invitations"
            :key="invitation.email"
            class="flex flex-col gap-2"
          >
            <div class="flex items-center gap-2 bg-orange-50 p-2">
              <div class="flex flex-col flex-1 sm:flex-initial">
                <p>
                  {{ invitation.email }}
                </p>
                <p>
                  <small>Expires: {{ DateStringToFormattedDateTime(invitation.validUntil, false) }}</small>
                </p>
              </div>
              <IconButton
                v-tooltip="'Cancel Invitation'"
                icon="heroicons:trash"
                severity="danger"
                @click="onCancelInvitation(invitation)"
              />
            </div>
          </div>
        </div>
      </div>

      <OverlayError v-if="organisationStatus == 'error'">
        <p>Organisation not found</p>
      </OverlayError>
    </template>
    <template #loader>
      <OverlayProgressSpinner
        :show="organisationStatus == 'pending'"
        :is-fixed="false"
        text="Loading organisation..."
      />
    </template>
  </WrappersListPage>
</template>

<script setup lang="ts">
import { Constants, DialogProps } from '~/config/const'
import {
  type Invitation,
  inviteToOrganisationSchema,
  type Organisation,
  type OrganisationMember,
  organisationNameSchema,
} from '~/schema/schema'
import SingleValueDialog from '~/components/dialogs/single-value-dialog.vue'
import type {
  ICustomConfirmDialog,
  IDialogOrganisationMemberInput,
  IDialogOrganisationMemberOutput,
  IDialogSingleValueInput,
  IDialogSingleValueOutput,
} from '~/config/interfaces'
import CustomConfirmationDialog from '~/components/dialogs/custom-confirmation-dialog.vue'
import OrganisationMemberDialog from '~/components/dialogs/organisation-member-dialog.vue'

const toast = useToast()
const route = useRoute()
const dialog = useDialog()
const organisationID = parseInt(route.params.id.toString())

const { user } = useAuth()
const {
  currentOrganisation,
  updateOrganisation,
  canEditOrganisation,
  isOwnerOfOrganisation,
  deleteMember,
  invitations,
  createInvitation,
  deleteInvitation,
} = useOrganisation(toast)

const config = useRuntimeConfig()
const { data: organisationData, error: organisationError, status: organisationStatus, refresh: organisationRefresh } = await useFetch<Organisation>(`/v1/organisation/${organisationID}`, {
  method: 'GET',
  lazy: true,
  baseURL: config.public.backendHost,
})
watch(organisationStatus, (value) => {
  if (value == 'success') {
    currentOrganisation.value = organisationData.value
  }
  else if (value == 'error') {
    toast.add({
      severity: 'error',
      summary: 'Fetch Failed',
      detail: organisationError.value,
    })
  }
})

const { data: invitationsData, error: invitationsError, status: invitationsStatus, refresh: invitationsRefresh } = await useFetch<Invitation[]>(`/v1/organisation/${organisationID}/invitations`, {
  method: 'GET',
  lazy: true,
  baseURL: config.public.backendHost,
})
watch(invitationsStatus, (value) => {
  if (value == 'success') {
    invitations.value = invitationsData.value ?? []
  }
  else if (value == 'error') {
    toast.add({
      severity: 'error',
      summary: 'Fetch Failed',
      detail: invitationsError.value,
    })
  }
})

useHead({
  title() {
    return currentOrganisation.value?.organisationName ?? 'Organisation not found'
  },
})

const onChangeOrganisationName = () => {
  dialog.open(SingleValueDialog, {
    props: {
      header: 'Change Organisation Name',
      ...DialogProps.BigDialog,
    },
    data: {
      value: currentOrganisation.value?.organisationName,
      placeholder: 'Enter Organisation Name*',
      schema: organisationNameSchema,
    } as IDialogSingleValueInput,
    onClose: async (dialogOptions) => {
      const value = dialogOptions?.data as IDialogSingleValueOutput
      if (value) {
        await updateOrganisation(organisationID, { organisationName: value })
          .then(() => {
            toast.add({
              severity: 'success',
              summary: 'Nice 👍',
              detail: 'Successfully updated organisation',
              life: Constants.ToasterDefaultLifeTime,
            })
            organisationRefresh()
          })
          .catch((reason) => {
            toast.add({
              severity: 'error',
              summary: 'Oh no 😱',
              detail: reason,
              life: Constants.ToasterDefaultLifeTime,
            })
          })
      }
    },
  })
}
const onInviteToOrganisation = () => {
  dialog.open(SingleValueDialog, {
    props: {
      header: 'Add Member to Organisation',
      ...DialogProps.BigDialog,
    },
    data: {
      value: '',
      placeholder: 'Enter Members E-Mail*',
      schema: inviteToOrganisationSchema,
    } as IDialogSingleValueInput,
    onClose: async (dialogOptions) => {
      const value = dialogOptions?.data as IDialogSingleValueOutput
      if (value) {
        await createInvitation(value)
          .then(() => {
            toast.add({
              severity: 'success',
              summary: 'Nice 👍',
              detail: 'Invitation has been sent successfully',
              life: Constants.ToasterDefaultLifeTime,
            })
            invitationsRefresh()
          })
          .catch((reason) => {
            toast.add({
              severity: 'error',
              summary: 'Oh no 😱',
              detail: reason,
              life: Constants.ToasterDefaultLifeTime,
            })
          })
      }
    },
  })
}
const onEditOrganisationMember = (member: OrganisationMember) => {
  dialog.open(OrganisationMemberDialog, {
    props: {
      header: 'Edit Organisation Member',
      ...DialogProps.SmallDialog,
    },
    data: {
      member: member,
    } as IDialogOrganisationMemberInput,
    onClose(dialogOptions) {
      const value = dialogOptions?.data as IDialogOrganisationMemberOutput
      if (value) {
        console.log(value)
      }
    },
  })
}
const onCancelInvitation = (invitation: Invitation) => {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: 'Delete Invitation',
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Are you sure you want to delete the invitation for "${invitation.email}"?`,
      confirmText: 'Yes',
      rejectText: 'No',
    } as ICustomConfirmDialog,
    onClose(options) {
      if (options?.data === true) {
        deleteInvitation(invitation.id)
          .then(() => {
            toast.add({
              severity: 'success',
              summary: 'Nice 👍',
              detail: 'Invitation has been deleted',
              life: Constants.ToasterDefaultLifeTime,
            })
            invitationsRefresh()
          })
          .catch((reason) => {
            toast.add({
              severity: 'error',
              summary: 'Oh no 😱',
              detail: reason,
              life: Constants.ToasterDefaultLifeTime,
            })
          })
      }
    },
  })
}
const onRemoveOrganisationMember = (member: OrganisationMember) => {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: 'Remove User From Organisation',
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Are you sure you want to remove the user "${member.email}" from "${currentOrganisation.value?.organisationName}"? All content of the user will be kept and transferred to the Owner.`,
      confirmText: 'Yes',
      rejectText: 'No',
    } as ICustomConfirmDialog,
    onClose(options) {
      if (options?.data === true) {
        deleteMember(member.id)
          .then(async () => {
            toast.add({
              severity: 'success',
              summary: 'Nice 👍',
              detail: 'You have successfully remove the member',
              life: Constants.ToasterDefaultLifeTime,
            })
            organisationRefresh()
          })
          .catch((reason) => {
            toast.add({
              severity: 'error',
              summary: 'Oh no 😱',
              detail: reason,
              life: Constants.ToasterDefaultLifeTime,
            })
          })
      }
    },
  })
}
const onLeaveOrganisation = (member: OrganisationMember) => {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: 'Leave Organisation',
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Are you sure you want to leave the organisation "${currentOrganisation.value?.organisationName}"? You will need to be reinvited. All your content will be kept and transferred to the Owner.`,
      confirmText: 'Yes',
      rejectText: 'No',
    } as ICustomConfirmDialog,
    onClose(options) {
      if (options?.data === true) {
        deleteMember(member.id)
          .then(async () => {
            toast.add({
              severity: 'success',
              summary: 'Nice 👍',
              detail: 'You have successfully left the organisation',
              life: Constants.ToasterDefaultLifeTime,
            })
            await navigateTo({ name: 'organisations' })
          })
          .catch((reason) => {
            toast.add({
              severity: 'error',
              summary: 'Oh no 😱',
              detail: reason,
              life: Constants.ToasterDefaultLifeTime,
            })
          })
      }
    },
  })
}
</script>
