<template>
  <form
    class="flex flex-col gap-4 w-full h-full"
    @submit.prevent="onSave"
  >
    <div class="flex flex-col gap-2">
      <InputText
        v-model.trim="value"
        autofocus
        :placeholder="placeholder"
        :invalid="parseResult && !parseResult.success"
        @blur="onBlur"
      />
      <small
        v-if="lostFocus && parseResult && !parseResult.success && parseResult.error.formErrors.formErrors.length > 0"
        id="username-help"
        class="text-red-400"
      >
        <p
          v-for="error of parseResult.error.formErrors.formErrors"
          :key="error"
        >{{ error }}</p>
      </small>
    </div>
    <div class="flex justify-end gap-2">
      <Button
        label="Save"
        type="submit"
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
import type { IDialogSingleValueInject } from '~/config/interfaces'

const dialogRef = inject<IDialogSingleValueInject>('dialogRef')!

const parseResult = ref<SafeParseReturnType<unknown, unknown>>()
const value = ref(dialogRef.value.data.value ?? '')
const placeholder = dialogRef.value.data.placeholder ?? ''
const lostFocus = ref(false)

watch(value, (currentValue) => {
  parseResult.value = dialogRef.value.data.schema.safeParse(currentValue)
})

const onSave = async () => {
  parseResult.value = dialogRef.value.data.schema.safeParse(value.value)
  if (!parseResult.value.success) {
    return
  }
  dialogRef.value.close(value.value)
}
const onCancel = () => {
  dialogRef.value.close(null)
}
const onBlur = () => {
  if (value.value.trim().length > 0) {
    lostFocus.value = true
  }
}
</script>
