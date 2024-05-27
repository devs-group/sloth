<template>

<div v-if="props.invitaions && props.invitaions.length > 0">
    <template v-for="invitation in props.invitaions" :key="invitation.user_id">
        <OrganisationInvitationRow :invitation="invitation" :organisation_id="props.organisation.id"</OrganisationInvitationRow>
    </template>
    </div>
    <div v-else>
        <OverlayProgressSpinner :show="props.isLoading" />
        <p v-if="!props.isLoading">No Invitations</p>
    </div>

</template>
<script lang="ts" setup>
import type { PropType } from 'vue';
import type { Invitation, Organisation } from '~/schema/schema';
import OrganisationInvitationRow from './rows/organisation-invitation.row.vue';



defineProps({
    props: {
        required: true,
        type: Object as PropType<{ organisation: Organisation, isLoading: boolean, invitaions: Invitation[] }>,
    },
});
</script>