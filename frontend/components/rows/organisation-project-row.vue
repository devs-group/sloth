<template>
  <div v-if="props && props.project" class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700">
    <div class="flex flex-col gap-1">
      <p class="break-all">{{ props.project?.name }}</p>
      <p class="text-xs text-prime-secondary-text">UPN: {{ props.project?.upn }}</p>
    </div>
    <div class="flex items-center gap-2">
        <IconButton     
            :loading="isDeleting"
            text
            severity="danger"
            icon="heroicons:trash"
            @click="onDelete()"
        />
        <NuxtLink :to="{name: Routes.PROJECT, params: {id: props.project?.id}}">
            <IconButton icon="heroicons:arrow-right-on-rectangle" />
        </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, ref } from 'vue';
import { Routes } from "~/config/routes";
import { type Organisation, type OrganisationProject } from "~/schema/schema";
import CustomConfirmationDialog from '~/components/dialogs/custom-confirmation-dialog.vue';
import type { ICustomConfirmDialog } from '~/config/interfaces';
import { Constants, DialogProps } from '~/config/const';

const props = defineProps({
  organisation:{
    type: Object as PropType<Organisation | undefined>,
  },
  project: {
    type: Object as PropType<OrganisationProject | undefined>,
    required: true,
  }
});

const emits = defineEmits<{
  (e: "on-delete"): void
}>()

const toast = useToast()
const dialog = useDialog()
const isDeleting = ref(false)

const onDelete = () => {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: 'Delete Project',
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Are you sure you want to remove the project "${props.project!.name}" from this organisation?`,
      confirmText: 'Delete',
      rejectText: 'Cancel',
    } as ICustomConfirmDialog,
    onClose(options) {
      if (options?.data === true) {
        isDeleting.value = true
        const organisation = useOrganisation(props.organisation!.id, toast)
        organisation.removeProjectFromOrganisation(props.project!.upn, props.project!.name)
        .then(() => {
          emits('on-delete')
          console.log("remoev")
        })
        .finally(() => {
          isDeleting.value = false
        })
      }
    },
  })
}
</script>
