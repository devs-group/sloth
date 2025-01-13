<template>
  <div class="flex flex-col gap-4 w-full h-full">
    <p class="break-words">
      {{ question }}
    </p>
    <div class="flex justify-end gap-2">
      <Button
        :label="confirmText"
        @click="onAccept"
      />
      <Button
        :label="rejectText"
        severity="secondary"
        @click="onReject"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ICustomConfirmDialog, IDialogInjectRef } from '~/config/interfaces'

const dialogRef = inject<IDialogInjectRef<ICustomConfirmDialog>>('dialogRef')

const question = dialogRef?.value.data.question ?? '<Sorry we forgot to put the question here 🙈>'
const confirmText = dialogRef?.value.data.confirmText ?? 'Yes'
const rejectText = dialogRef?.value.data.rejectText ?? 'No'

const onAccept = async () => {
  dialogRef?.value.close(true)
}
const onReject = () => {
  dialogRef?.value.close(false)
}
</script>
