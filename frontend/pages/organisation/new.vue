<script setup lang="ts">
import { type OrgaisationSchema, organisationSchema } from "~/schema/schema";

const isSubmitting = ref(false);
const toast = useToast();
const router = useRouter();
const config = useRuntimeConfig();

const g = ref<OrgaisationSchema>({
  organisation_name: "",
});

async function saveOrganisation() {
  const data = organisationSchema.parse(g.value);
  isSubmitting.value = true;
  try {
    await $fetch(`${config.public.backendHost}/v1/organisation`, {
      method: "POST",
      body: data,
      credentials: "include",
    });
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Your Organisation has been created successfully",
    });
    await router.push("/organisation");
  } catch (e) {
    console.error(e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Something went wrong",
    });
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <form
    :schema="organisationSchema"
    :state="g"
    @submit="saveOrganisation"
    class="p-12 w-full"
  >
    <div class="flex flex-row pb-12">
      <InputGroup>
        <InputText v-model="g!.organisation_name" class="max-w-[20em]" />
        <IconButton
          label="Create Group"
          icon="heroicons:bolt"
          :loading="isSubmitting"
          @click="saveOrganisation"
        />
      </InputGroup>
    </div>
  </form>
</template>
