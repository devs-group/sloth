<script setup lang="ts">
import { type GroupSchema, groupSchema } from "~/schema/schema";

const isSubmitting = ref(false);
const toast = useToast();
const router = useRouter();
const config = useRuntimeConfig();

const g = ref<GroupSchema>({
  group_name: "",
});

async function saveGroup() {
  const data = groupSchema.parse(g.value);
  isSubmitting.value = true;
  try {
    await $fetch(`${config.public.backendHost}/v1/group`, {
      method: "POST",
      body: data,
      credentials: "include",
    });
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Your Group has been created successfully",
    });
    await router.push("/group");
  } catch (e) {
    console.error(e);
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Something went wrong",
    });
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <form
    :schema="groupSchema"
    :state="g"
    @submit="saveGroup"
    class="p-12 w-full"
  >
    <div class="flex flex-row pb-12">
      <InputGroup>
        <InputText v-model="g!.group_name" class="max-w-[20em]" />
        <IconButton
          label="Create Group"
          icon="heroicons:bolt"
          :loading="isSubmitting"
          @click="saveGroup"
        />
      </InputGroup>
    </div>
  </form>
</template>
