<script lang="ts" setup>
import type { Group } from "~/schema/schema";

const { showConfirmation } = useConfirmation();

interface State {
  isRemoving?: boolean;
}
const state = ref<Record<string, State>>({});

const props = defineProps({
  group: {
    required: true,
    type: Object as PropType<Group>,
  },
});

const config = useRuntimeConfig();
defineEmits<{
  (event: "deleteMember", member: string): void;
}>();
</script>
<template>
  <ul class="list-disc pl-5">
    <li
      v-for="member in group?.members"
      :key="member"
      class="flex justify-between items-center mb-2 pl-5"
    >
      <span class="text-gray-800">{{ member }}</span>
      <UButton
        icon="i-heroicons-trash"
        :loading="state[member]?.isRemoving"
        variant="ghost"
        color="red"
        @click="
          () =>
            showConfirmation(
              'Remove the member?',
              'Are you sure you wanna remove this user from your group?',
              () => $emit('deleteMember', member as string)
            )
        "
      >
      </UButton>
    </li>
  </ul>
</template>
