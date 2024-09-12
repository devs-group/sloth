<template>
    <div class="flex flex-col gap-6 py-6 w-full max-w-6xl">
        <div class="flex flex-col gap-2 px-6">
            <h1 class="text-2xl">Organisations</h1>
            <p class="text-sm text-gray-400">
                {{ invitationsData?.length ?? 0 }} open invitations
            </p>
        </div>
        <div class="flex items-end justify-between px-6">
            <div>
                <p class="text-sm text-gray-400">
                    Member of {{ organisations?.length ?? 0 }} Organisations
                </p>
            </div>
            <IconButton
                label="Create Organisation"
                icon="heroicons:rocket-launch"
                aria-label="create"
                @click="onCreateOrganisation()"
            />
        </div>

        <div>
            <div
                v-for="organisation in organisations"
                class="p-6 flex flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700"
            >
                <div class="flex items-center">
                    <Avatar
                        :alt="organisation.organisation_name"
                        class="mr-3"
                    />
                    <div class="w-2/3">
                        <p>{{ organisation.organisation_name }}</p>
                    </div>
                </div>
                <div class="space-x-4 flex items-center">
                    <IconButton
                        icon="heroicons:trash"
                        :loading="
                            state[organisation.organisation_name]?.isRemoving
                        "
                        text
                        severity="danger"
                        @click="onDeleteOrganisation(organisation)"
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
        </div>
    </div>
</template>
<script lang="ts" setup>
import { ref } from "vue";
import { type Organisation } from "~/schema/schema";
import { Routes } from "~/config/routes";
import { useOrganisations } from "~/composables/useOrganisations";
import { DialogProps } from "~/config/const";
import CreateOrganisationDialog from "~/components/dialogs/create-organisation-dialog.vue";
import CustomConfirmationDialog from "~/components/dialogs/custom-confirmation-dialog.vue";
import type {
    ICustomConfirmDialog,
    ICreateOrganisationResponse,
} from "~/config/interfaces";

const toast = useToast();
const { loadOrganisations, loadInvitations, deleteOrganisation } =
    useOrganisations(toast);
const { data: organisations } = loadOrganisations();
const { data: invitationsData } = loadInvitations();

const dialog = useDialog();
const confirm = useConfirm();
const state = ref<Record<string, OrganisationState>>({});

interface OrganisationState {
    isRemoving?: boolean;
}

function onDeleteOrganisation(organisation: Organisation) {
    dialog.open(CustomConfirmationDialog, {
        props: {
            header: "Logout",
            ...DialogProps.SmallDialog,
        },
        data: {
            question: `Do you want to delete "${organisation.organisation_name}"? This action cannot be undone.`,
            confirmText: "Delete",
            rejectText: "Cancel",
        } as ICustomConfirmDialog,
        onClose(options) {
            if (options?.data === true) {
                try {
                    deleteOrganisation(organisation);
                } catch (err) {
                    console.error("Failed to delete organisation:", err);
                }
            }
        },
    });
}

const onCreateOrganisation = () => {
    dialog.open(CreateOrganisationDialog, {
        props: {
            header: "Create New Organisation",
            ...DialogProps.BigDialog,
        },
        onClose(options) {
            const data = options?.data as ICreateOrganisationResponse;
            organisations.value?.push(data);
        },
    });
};
</script>
