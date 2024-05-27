<template>

<div v-if="invitations && invitations.length > 0">
    <template v-for="invitation in invitations" :key="invitation.user_id">
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
import type { Organisation } from '~/schema/schema';
import OrganisationInvitationRow from './rows/organisation-invitation.row.vue';

const isLoading = ref<boolean>(false)
const toast = useToast()

const { loadInvitations, invitations } = useOrganisations(toast)

onMounted(async () => {
  isLoading.value = true;
  await loadInvitations();
  isLoading.value = false; 
});

defineProps({
    props: {
        required: true,
        type: Object as PropType<{ organisation: Organisation, isLoading: boolean }>,
    },
});
</script>