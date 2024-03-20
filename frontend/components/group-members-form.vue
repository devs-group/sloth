<script lang="ts" setup>
import type { Group } from "~/schema/schema";

const confirm = useConfirm();
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
      v-for="member in props.group?.members"
      :key="member"
      class="flex justify-between items-center mb-2 pl-5"
    >
      <span class="text-gray-800">{{ member }}</span>

      <IconButton
        text
        severity="danger"
        icon="heroicons:trash"
        :loading="state[member]?.isRemoving"
        @click="
          () =>  confirm.require({
                  header: 'Remove the member?',
                  message: 'Are you sure you wanna remove this user from your group?',
                  accept: () => $emit('deleteMember', member as string),
                  acceptLabel:'Delete',
                  rejectLabel: 'Cancel',
                })
        "
      />
    </li>
  </ul>
</template>
