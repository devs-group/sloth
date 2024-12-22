<template>
    <WrappersListPage
        title="Organisations"
        :description="`You are a member of ${
            organisations?.length ?? 0
        } organisations`"
    >
        <template #actions>
            <IconButton
                label="Add organisation"
                icon="heroicons:plus"
                aria-label="create"
                @click="onCreateOrganisation"
            />
        </template>
        <template #content>
            <OverlayProgressSpinner
                :show="isLoadingOrganisations"
            ></OverlayProgressSpinner>
            <OrganisationRow
                v-for="organisation in organisations"
                @on-delete="onDeleteOrganisation"
                :organisation="organisation"
            />
        </template>
    </WrappersListPage>
</template>

<script lang="ts" setup>
import { type Organisation } from "~/schema/schema";
import { DialogProps } from "~/config/const";
import CreateOrganisationDialog from "~/components/dialogs/create-organisation-dialog.vue";
import OrganisationRow from "~/components/rows/organisation-row.vue";
import { APIService } from "~/api";

const dialog = useDialog();
const {
    data: organisations,
    isLoading: isLoadingOrganisations,
    execute: getOrganisations,
} = useApi(() => APIService.GET_organisations(), {});

onMounted(async () => {
    await getOrganisations();
});

const onCreateOrganisation = () => {
    dialog.open(CreateOrganisationDialog, {
        props: {
            header: "Create new organisation",
            ...DialogProps.BigDialog,
        },
        onClose(options) {
            const data = options?.data as Organisation;
            (organisations.value ??= []).push(data);
        },
    });
};

function onDeleteOrganisation(id: number) {
    organisations.value =
        organisations.value?.filter((o) => o.id !== id) || null;
}
</script>
