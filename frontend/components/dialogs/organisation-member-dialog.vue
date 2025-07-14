<template>
  <form
    class="grid grid-cols-1 sm:grid-cols-2 gap-4 w-full h-full"
    @submit.prevent="onSave"
  >
    <div class="flex flex-col gap-2 col-span-full">
      <IftaLabel>
        <InputText
          id="email"
          v-model.trim="member.email"
          class="w-full"
          autofocus
          placeholder="E-Mail"
          disabled
          :invalid="parseResult && !parseResult.success && !!parseResult.error.formErrors.fieldErrors.email"
        />
        <label for="email">E-Mail</label>
      </IftaLabel>

      <small
        v-if="parseResult && !parseResult.success && parseResult.error.formErrors.fieldErrors.email"
        class="text-red-400"
      >
        <p
          v-for="error of parseResult.error.formErrors.fieldErrors.email"
          :key="error"
        >{{ error }}</p>
      </small>
    </div>

    <div class="flex flex-col gap-2">
      <IftaLabel>
        <InputText
          id="username"
          v-model.trim="member.username"
          class="w-full"
          autofocus
          placeholder="E-Mail"
          disabled
          :invalid="parseResult && !parseResult.success && !!parseResult.error.formErrors.fieldErrors.username"
        />
        <label for="username">Username</label>
      </IftaLabel>

      <small
        v-if="parseResult && !parseResult.success && parseResult.error.formErrors.fieldErrors.username"
        class="text-red-400"
      >
        <p
          v-for="error of parseResult.error.formErrors.fieldErrors.username"
          :key="error"
        >{{ error }}</p>
      </small>
    </div>

    <div class="flex flex-col gap-2">
      <IftaLabel>
        <Select
          id="role"
          v-model="member.role"
          :options="OrganisationMemberRoleOptions"
          option-label="label"
          option-value="value"
          placeholder="Select Role*"
          :disabled="isCurrentlyOwner"
          class="w-full"
          :invalid="parseResult && !parseResult.success && !!parseResult.error.formErrors.fieldErrors.role"
        />
        <label for="role">Role</label>
      </IftaLabel>
      <small
        v-if="parseResult && !parseResult.success && parseResult.error.formErrors.fieldErrors.role"
        class="text-red-400"
      >
        <p
          v-for="error of parseResult.error.formErrors.fieldErrors.role"
          :key="error"
        >{{ error }}</p>
      </small>
    </div>

    <Message
      v-if="isCurrentlyOwner"
      severity="warn"
      size="small"
      class="col-span-full"
    >
      You can't change your role because you are the Owner. Select another member to declare a new Owner.
    </Message>

    <Message
      v-if="showNewOwnerHint"
      severity="warn"
      size="small"
      class="col-span-full"
    >
      Attention: Your account will be set back to "Admin" and this user will be the new Owner
    </Message>

    <div class="flex justify-end gap-2 col-span-full">
      <Button
        label="Save"
        type="submit"
        :disabled="parseResult && !parseResult.success"
        @click="onSave"
      />
      <Button
        label="Cancel"
        severity="secondary"
        @click="onCancel"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import type { SafeParseReturnType } from 'zod'
import { computed } from 'vue'
import type { IDialogOrganisationMemberInject } from '~/config/interfaces'
import {
  type OrganisationMember,
  OrganisationMemberRoleEnum,
  OrganisationMemberRoleOptions,
  type OrganisationMemberUpdate,
  organisationMemberUpdateSchema,
} from '~/schema/schema'

const dialogRef = inject<IDialogOrganisationMemberInject>('dialogRef')!

const parseResult = ref<SafeParseReturnType<OrganisationMember, OrganisationMemberUpdate>>()
const member = ref<OrganisationMember>({ ...dialogRef.value.data.member })

watch(member, (currentValue) => {
  parseResult.value = organisationMemberUpdateSchema.safeParse(currentValue)
}, { deep: true, immediate: true })

const isCurrentlyOwner = computed(() => dialogRef.value.data.member.role === OrganisationMemberRoleEnum.Values.owner)
const showNewOwnerHint = computed(() => member.value.role === OrganisationMemberRoleEnum.Values.owner && !isCurrentlyOwner.value)

const onSave = async () => {
  parseResult.value = organisationMemberUpdateSchema.safeParse(member.value)
  if (!parseResult.value.success) {
    return
  }
  dialogRef.value.close(member.value)
}
const onCancel = () => {
  dialogRef.value.close(null)
}
</script>
