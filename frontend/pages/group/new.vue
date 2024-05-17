<script setup lang="ts">
import { type GroupSchema, organizationSchema } from "~/schema/schema";

const isSubmitting = ref(false);
const toast = useToast();
const router = useRouter();
const config = useRuntimeConfig();

const g = ref<GroupSchema>({
  organization_name: "",
});

async function saveOrganization() {
  const data = organizationSchema.parse(g.value);
  isSubmitting.value = true;
  try {
    await $fetch(`${config.public.backendHost}/v1/organization`, {
      method: "POST",
      body: data,
      credentials: "include",
    });
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Your Organization has been created successfully",
      life: 3000,
    });
    await router.push("/group");
  } catch (e) {
    console.error(e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Something went wrong", // Show correct error message
    });
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <form
    :schema="organizationSchema"
    :state="g"
    @submit="saveOrganization"
    class="p-12 w-full"
  >
    <div class="flex flex-row pb-12">
      <InputGroup>
        <InputText v-model="g!.organization_name" class="max-w-[20em]" />
        <IconButton
          label="Create Group"
          icon="heroicons:bolt"
          :loading="isSubmitting"
          @click="saveOrganization"
        />
      </InputGroup>
    </div>
  </form>
</template>
