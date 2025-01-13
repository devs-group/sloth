<template>
  <div
    v-if="props && props.invitation"
    class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700"
  >
    <div class="flex flex-col gap-1">
      <p class="break-all">
        {{ props.invitation.email }}
      </p>
    </div>
    <div class="flex items-center gap-2">
      <IconButton
        :loading="isWithdrawn"
        text
        severity="danger"
        icon="heroicons:backspace"
        @click="onWithdraw()"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
import type { PropType } from 'vue'
import type { Invitation } from '~/schema/schema'
import { DialogProps } from '~/config/const'
import CustomConfirmationDialog from '~/components/dialogs/custom-confirmation-dialog.vue'
import type { ICustomConfirmDialog } from '~/config/interfaces'

const dialog = useDialog()
const isWithdrawn = ref(false)

const props = defineProps({
  invitation: {
    required: true,
    type: Object as PropType<Invitation | undefined>,
  },
  organisationId: {
    required: true,
    type: Number,
  },
})

const toast = useToast()

const emits = defineEmits<{
  (e: 'on-withdraw'): void
}>()

const onWithdraw = () => {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: 'Withdraw Invitation',
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Are you sure you want to withdraw the invitation for "${props.invitation?.email}"`,
      confirmText: 'Withdraw',
      rejectText: 'Cancel',
    } as ICustomConfirmDialog,
    onClose(options) {
      if (options?.data === true) {
        isWithdrawn.value = true
        const organisation = useOrganisationInviation(toast)
        organisation
          .withdrawInvitation(props.invitation!.email, props.organisationId) // TODO: invitation code
          .then(() => {
            emits('on-withdraw')
          })
          .finally(() => {
            isWithdrawn.value = false
          })
      }
    },
  })
}
</script>
