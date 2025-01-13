<template>
  <OrganisationHeader :props="{ organisation_name: props.props.organisation.organisation_name, button: props.props.button }" />

  <div
    v-if="props && props.props.organisation.members && props.props.organisation.members.length > 0"
  >
    <div
      v-for="member in props.props.organisation.members"
      :key="member.user_id"
      class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700"
    >
      <div class="flex flex-col gap-1">
        <p class="break-all">
          {{ member.email }}
        </p>
        <p
          v-if="member.username"
          class="text-xs text-prime-secondary-text"
        >
          Username: {{ member.username }}
        </p>
      </div>

      <IconButton
        text
        severity="danger"
        icon="heroicons:trash"
        :loading="isDeleting"
        @click="onDelete(member)"
      />
    </div>
  </div>
  <div
    v-else
    class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700"
  >
    <p>No members found.</p>
  </div>
</template>

<script lang="ts" setup>
import type { Organisation, OrganisationMember } from '~/schema/schema'
import CustomConfirmationDialog from '~/components/dialogs/custom-confirmation-dialog.vue'
import { DialogProps } from '~/config/const'
import type { ICustomConfirmDialog } from '~/config/interfaces'

const isDeleting = ref(false)
const toast = useToast()
const dialog = useDialog()

const props = defineProps({
  props: {
    required: true,
    type: Object as PropType<{ organisation: Organisation, button: { label: string, icon: string, onClick: () => void }, emits: { deleteMember: () => void } }>,
  },
})

const onDelete = (member: OrganisationMember) => {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: 'Delete Project',
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Are you sure you want to remove the member "${member.email}" from this organisation?`,
      confirmText: 'Delete',
      rejectText: 'Cancel',
    } as ICustomConfirmDialog,
    onClose(options) {
      if (options?.data === true) {
        isDeleting.value = true
        const organisation = useOrganisation(props.props.organisation.id ?? '', toast)
        organisation.deleteMember(member.user_id)
          .then(async () => {
            props.props.emits.deleteMember()
          })
          .finally(() => {
            isDeleting.value = false
          })
      }
    },
  })
}
</script>
