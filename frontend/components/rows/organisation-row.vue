<template>
  <div
    class="p-6 flex flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700"
  >
    <div class="flex items-center">
      <Avatar :alt="props.organisation.organisation_name" class="mr-3" />
      <div class="w-2/3">
        <p>{{ props.organisation.organisation_name }}</p>
      </div>
    </div>
    <div class="space-x-4 flex items-center">
      <IconButton
        icon="heroicons:trash"
        :loading="isDeletingOrganisation"
        text
        severity="danger"
        @click="openDeleteOrganisationDialog(props.organisation)"
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
</template>

<script setup lang="ts">
import type { Organisation } from "~/schema/schema";
import { Routes } from "~/config/routes";
import { DialogProps } from "~/config/const";
import CustomConfirmationDialog from "~/components/dialogs/custom-confirmation-dialog.vue";
import type { ICustomConfirmDialog } from "~/config/interfaces";
import { APIService } from "~/api";

const props = defineProps({
  organisation: {
    type: Object as PropType<Organisation>,
    required: true,
  },
});

const emits = defineEmits<{ (event: "onDelete", id: number): void }>();

const dialog = useDialog();
const { isLoading: isDeletingOrganisation, execute: deleteOrganisation } =
  useApi(
    (organisationID: number) => APIService.DELETE_organisation(organisationID),
    {
      showSuccessToast: true,
      successMessage: "Organisation has been deleted succesfully",
    }
  );

function openDeleteOrganisationDialog(organisation: Organisation) {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: "Delete organisation",
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Do you want to delete "${organisation.organisation_name}"? This action cannot be undone.`,
      confirmText: "Delete",
      rejectText: "Cancel",
    } as ICustomConfirmDialog,
    async onClose(options) {
      if (options?.data === true) {
        await deleteOrganisation(organisation.id);
        emits("onDelete", organisation.id);
      }
    },
  });
}
</script>
