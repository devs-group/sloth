<template>
  <OrganisationHeader :props="{ organisation_name: props.organisation.organisation_name, button: props.button }" />
  <div v-if="props.invitations && props.invitations.length > 0">
    <template
      v-for="invitation in props.invitations"
      :key="invitation.user_id"
    >
      <OrganisationInvitationRow
        :invitation="invitation"
        :organisation-id="props.organisation.id"
        @on-withdraw="props.emits.withdrawInvitation()"
      />
    </template>
  </div>
  <div
    v-else
    class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700"
  >
    <OverlayProgressSpinner :show="props.isLoading" />
    <p v-if="!props.isLoading">
      No invitations found.
    </p>
  </div>
</template>

<script lang="ts" setup>
import type { PropType } from 'vue'
import OrganisationInvitationRow from './rows/organisation-invitation.row.vue'
import type { Invitation, Organisation } from '~/schema/schema'

defineProps({
  props: {
    required: true,
    type: Object as PropType<{ organisation: Organisation, isLoading: boolean, invitations: Invitation[], button: { label: string, icon: string, onClick: () => void }, emits: { withdrawInvitation: () => void } }>,
  },
})
</script>
