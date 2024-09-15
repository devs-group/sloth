<template>
  <div
    class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700"
  >
    <div class="flex flex-col gap-1">
      <p class="break-all">{{ project.name }}</p>
      <p class="text-xs text-prime-secondary-text">UPN: {{ project.upn }}</p>
      <p class="text-xs text-prime-secondary-text">
        Hook URL: {{ project.hook }}
      </p>
      <p class="text-xs text-prime-secondary-text">
        Services: {{ project.services.length }}
        {{
          project.services.length > 0
            ? `(${project.services.map((s) => s.name).join(", ")})`
            : ""
        }}
      </p>
      <p v-if="publicHosts" class="text-xs text-prime-secondary-text">
        Hosts: {{ publicHosts }}
      </p>
    </div>
    <div class="flex items-center gap-2">
      <IconButton
        :loading="isDeleting"
        text
        severity="danger"
        icon="heroicons:trash"
        @click="onDelete"
      />
      <NuxtLink :to="{ name: Routes.PROJECT, params: { id: project.id } }">
        <Button
          icon-pos="left"
          icon="pi pi-pencil"
          size="small"
          label="Details"
          rounded
          outlined
        ></Button>
      </NuxtLink>
      <Button
        icon-pos="left"
        icon="pi pi-cloud"
        size="small"
        label="Deploy"
        rounded
        outlined
        :loading="isDeploying"
        @click="onDeploy"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Project } from "~/schema/schema";
import { Routes } from "~/config/routes";
import { Constants, DialogProps } from "~/config/const";
import CustomConfirmationDialog from "../dialogs/custom-confirmation-dialog.vue";
import type { ICustomConfirmDialog } from "~/config/interfaces";

const props = defineProps({
  project: {
    type: Object as PropType<Project>,
    required: true,
  },
});

const emits = defineEmits<{
  (e: "on-delete", id: number): void;
  (e: "on-deploy", id: number): void;
}>();

const config = useRuntimeConfig();
const dialog = useDialog();
const toast = useToast();

const isDeleting = ref(false);
const isDeploying = ref(false);

const publicHosts = computed(() =>
  props.project.services
    .filter((s) => s.public.enabled)
    .map((s) => s.public.hosts)
    .join(",")
);

const onDelete = () => {
  dialog.open(CustomConfirmationDialog, {
    props: {
      header: "Delete Project",
      ...DialogProps.SmallDialog,
    },
    data: {
      question: `Are you sure you want to delete the project "${props.project.name}" entirely?`,
      confirmText: "Delete",
      rejectText: "Cancel",
    } as ICustomConfirmDialog,
    onClose(options) {
      if (options?.data === true) {
        isDeleting.value = true;
        $fetch(`${config.public.backendHost}/v1/project/${props.project.id}`, {
          method: "DELETE",
          credentials: "include",
        })
          .then(() => {
            emits("on-delete", props.project.id);
            toast.add({
              severity: "success",
              summary: "Success",
              detail: `Project "${props.project.name}" has been deleted successfully`,
              life: Constants.ToasterDefaultLifeTime,
            });
          })
          .catch((e) => {
            toast.add({
              severity: "error",
              summary: "Error",
              detail: `Failed to delete project "${props.project.name}"`,
              life: Constants.ToasterDefaultLifeTime,
            });
          })
          .finally(() => {
            isDeleting.value = false;
          });
      }
    },
  });
};

const onDeploy = () => {
  isDeploying.value = true;
  $fetch(props.project.hook, {
    method: "GET",
    headers: {
      "X-Access-Token": props.project.access_token,
    },
  })
    .then(() => {
      emits("on-deploy", props.project.id);
      toast.add({
        severity: "success",
        summary: "Success",
        detail: "Project has been deployed successfully",
        life: Constants.ToasterDefaultLifeTime,
      });
    })
    .catch((e) => {
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Failed to deploy project",
        life: Constants.ToasterDefaultLifeTime,
      });
    })
    .finally(() => {
      isDeploying.value = false;
    });
};
</script>
