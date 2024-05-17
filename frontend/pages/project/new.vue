<script setup lang="ts">
import { useTabs } from "~/composables/useTabs";
import { projectSchema } from "~/schema/schema";
import type { ProjectSchema } from "~/schema/schema";
const p = ref<ProjectSchema | undefined>({
  name: "",
  services: [],
  docker_credentials: [],
});

const { addService, removeService, addPort, addCredential, addEnv ,addHost, addVolume,
        removeCredential, removeEnv, removeHost, removePort, removeVolume } = useService(p);

const isSubmitting = ref(false);
const toast = useToast();
const router = useRouter();
const config = useRuntimeConfig();
const { tabs, activeTabComponent } = useTabs();

async function saveProject() {
  const data = projectSchema.parse(p.value);
  isSubmitting.value = true;
  try {
    await $fetch(`${config.public.backendHost}/v1/project`, {
      method: "POST",
      body: data,
      credentials: "include",
    });
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Your project has been created successfully",
      life: 3000,
    });
    await router.push("/project");
  } catch (e) {
    console.error(e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Something went wrong",
      life: 3000,
    });
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <form class="p-12 flex flex-col flex-1 overflow-hidden">
    <div class="flex flex-row pb-12">
      <InputGroup>
        <InputText v-model="p!.name" class="max-w-[20em]" />
        <IconButton
          label="Create Project"
          icon="heroicons:bolt"
          @click="saveProject"
          :disabled="!p?.name || p.services.length === 0"
          :loading="isSubmitting"
        />
      </InputGroup>
    </div>

    <Menubar :model="tabs" />
    <component
      :is="activeTabComponent"
      :credentials="p!.docker_credentials"
      @add-credential="addCredential"
      @remove-credential="removeCredential"
      :services="p!.services"
      @add-service="addService"
      @add-env="addEnv"
      @remove-env="removeEnv"
      @add-volume="addVolume"
      @remove-volume="removeVolume"
      @remove-service="removeService"
      @add-port="addPort"
      @remove-port="removePort"
      @add-host="addHost"
      @remove-host="removeHost"
    ></component>
  </form>
</template>
