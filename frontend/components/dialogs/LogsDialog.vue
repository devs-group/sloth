<template>
  <div
    ref="dialogContentRef"
    class="overflow-y-auto flex-1 border border-zinc-400"
  >
    <div
      v-for="(line, i) in logLines"
      :key="line"
      class="font-mono text-xs"
    >
      <p class="p-1 hover:bg-gray-200">
        {{ i }} {{ line }}
      </p>
    </div>
  </div>

  <div class="flex justify-end pt-4">
    <ToggleButton
      v-model="autoscrollLogs"
      on-icon="pi pi-play"
      on-label="Auto scroll"
      off-icon="pi pi-pause"
      off-label="Manual scroll"
      severity="secondary"
      @change="onToogleAutoscroll"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import type { IBaseDialog, ILogsDialogData } from '~/interfaces/dialog-interfaces'

const dialogRef = inject<IBaseDialog<ILogsDialogData>>('dialogRef')!

const project = ref(dialogRef.value.data.project)
const service = ref(dialogRef.value.data.service)

const { streamServiceLogs } = useService(project)

const logLines = ref<string[]>([])
const dialogContentRef = ref<HTMLElement>()
const autoscrollLogs = ref(true)
const lastLogScrollTop = ref(0)
const isProgrammaticScroll = ref(false)
const scrollMargin = 64

onMounted(() => {
  if (dialogContentRef.value && dialogContentRef.value instanceof HTMLElement) {
    const contentElement = dialogContentRef.value
    contentElement.addEventListener('scroll', onManualScroll)
  }

  const { data } = streamServiceLogs(project.value.upn!, service.value.usn!)
  watch(data, (value) => {
    if (value) {
      logLines.value.push(value)
      executeAutoscroll()
    }
  })
})

onUnmounted(() => {
  if (dialogContentRef.value && dialogContentRef.value instanceof HTMLElement) {
    const contentElement = dialogContentRef.value
    contentElement.removeEventListener('scroll', onManualScroll)
  }
})

const executeAutoscroll = () => {
  if (autoscrollLogs.value && dialogContentRef.value && dialogContentRef.value instanceof HTMLElement) {
    const contentElement = dialogContentRef.value

    isProgrammaticScroll.value = true
    nextTick(() => {
      contentElement.scrollTo(0, contentElement.scrollHeight)
      setTimeout(() => {
        isProgrammaticScroll.value = false
        lastLogScrollTop.value = contentElement.scrollTop
      }, 0)
    })
  }
}

const onToogleAutoscroll = () => {
  executeAutoscroll()
}

const onManualScroll = (event: Event) => {
  if (isProgrammaticScroll.value) {
    return
  }
  const target = event.target as HTMLElement
  const currentScrollTop = target.scrollTop
  const maxScrollTop = target.scrollHeight - target.clientHeight

  if (maxScrollTop - currentScrollTop > scrollMargin) {
    if (autoscrollLogs.value) {
      autoscrollLogs.value = false
    }
  }
  // Benutzer scrollt wieder ganz runter innerhalb der Marge
  else if (maxScrollTop - currentScrollTop <= scrollMargin) {
    if (!autoscrollLogs.value) {
      autoscrollLogs.value = true
    }
  }

  lastLogScrollTop.value = currentScrollTop
}
</script>
