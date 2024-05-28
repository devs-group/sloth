<template>
<div v-if="props.invitations && props.invitations.length > 0">
    <template v-for="invitation in props.invitations" :key="invitation.user_id">
        <OrganisationInvitationRow :invitation="invitation" :organisation_id="props.organisation.id"</OrganisationInvitationRow>
    </template>
    </div>
    <div v-else class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700">
        <OverlayProgressSpinner :show="props.isLoading" />
        <p v-if="!props.isLoading">No invitations found.</p>
    </div>

</template>
<script lang="ts" setup>
import type { PropType } from 'vue';
import type { Invitation, Organisation } from '~/schema/schema';
import OrganisationInvitationRow from './rows/organisation-invitation.row.vue';

defineProps({
    props: {
        required: true,
        type: Object as PropType<{ organisation: Organisation, isLoading: boolean, invitations: Invitation[] }>,
    },
});
</script>