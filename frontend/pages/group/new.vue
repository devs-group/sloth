<script setup lang="ts">
import type { FormSubmitEvent } from "@nuxt/ui/dist/runtime/types";
import { GroupSchema, groupSchema } from "~/schema/schema";

const isSubmitting = ref(false);
const { showError, showSuccess } = useNotification();
const router = useRouter();
const config = useRuntimeConfig();

const g = ref<GroupSchema>({
  group_name: "",
});

async function saveGroup(event: FormSubmitEvent<GroupSchema>) {
  const data = groupSchema.parse(event.data);
  isSubmitting.value = true;
  try {
    await $fetch(`${config.public.backendHost}/v1/group`, {
      method: "POST",
      body: data,
      credentials: "include",
    });
    showSuccess("Success", "Your Group has been created successfully");
    await router.push("/group");
  } catch (e) {
    console.error(e);
    showError("Error", "Something went wrong");
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <UForm
    :schema="groupSchema"
    :state="g"
    @submit="saveGroup"
    class="p-12 w-full"
  >
    <div class="flex flex-row items-end space-x-6 pb-12">
      <UFormGroup label="Name" name="name" required>
        <UInput v-model="g!.group_name" class="w-full md:w-72" required />
      </UFormGroup>
      <UButton type="submit" icon="i-heroicons-bolt" :loading="isSubmitting">
        Create Group
      </UButton>
    </div>
  </UForm>
</template>
