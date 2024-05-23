<template>
  <OrganisationHeader :props="{ organisation_name: props.props.organisation.organisation_name, button: props.props.button }" ></OrganisationHeader>

  <div class="flex flex-col gap-2 px">
  <ul class="list-disc pl-5">
    <li
      v-for="member in props.props?.organisation.members"
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
                  message: 'Are you sure you wanna remove this user from your organisation?',
                  accept: () => $emit('deleteMember', member as string),
                  acceptLabel:'Delete',
                  rejectLabel: 'Cancel',
                })
        "
      />
    </li>
  </ul>
  </div>
</template>
<script lang="ts" setup>
import type { Organisation } from "~/schema/schema";

const confirm = useConfirm();
interface State {
  isRemoving?: boolean;
}
const state = ref<Record<string, State>>({});

const props = defineProps({
  props: {
    required: true,
    type: Object as PropType<{ organisation: Organisation, button: { label: string, icon: string, onClick: () => void }}>,
  },
});

defineEmits<{
  (event: "deleteMember", member: string): void;
}>();
</script>

