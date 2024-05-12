<template>
  <form
      @submit.prevent="saveOrganisation"
      class="p-12 w-full"
  >
    <div class="flex flex-row pb-12">
      <InputGroup>
        <InputText v-model="orgName" class="max-w-[20em]" />
        <IconButton
            label="Create Organisation"
            type="submit"
            icon="heroicons:bolt"
            :loading="isSubmitting"
            @click="saveOrganisation"
        />
      </InputGroup>
    </div>
  </form>
</template>

<script setup lang="ts">
import {Routes} from "~/config/routes";
import {Constants} from "~/config/const";

const isSubmitting = ref(false);
const toast = useToast();
const router = useRouter();
const config = useRuntimeConfig();

const orgName = ref('')

async function saveOrganisation() {
  if (!orgName.value.trim().length) {
    return
  }
  isSubmitting.value = true;
  try {
    await $fetch(`${config.public.backendHost}/v1/organisation`, {
      method: "POST",
      body: {
        organisation_name: orgName.value,
      } as CreateOrganisationRequest,
      credentials: "include",
    });
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Your Organisation has been created successfully",
      life: Constants.ToasterDefaultLifeTime,
    });
    await router.push(Routes.ORGANISATIONS);
  } catch (e) {
    console.error(e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Something went wrong",
      life: Constants.ToasterDefaultLifeTime,
    });
  } finally {
    isSubmitting.value = false;
  }
}
</script>