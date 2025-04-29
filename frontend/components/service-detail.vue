<template>
  <div>
    <div v-if="service?.usn && serviceState">
      <div>
        <p
          v-if="serviceState[service.usn]"
          class="text-xs text-prime-secondary-text"
        >
          State: {{ serviceState[service.usn].state }}
        </p>
        <p
          v-if="serviceState[service.usn]"
          class="text-xs text-prime-secondary-text"
        >
          Status: {{ serviceState[service.usn].status }}
        </p>
      </div>
      <div class="flex flex-row items-start gap-2 mt-2">
        <Button
          label="Logs"
          icon-pos="left"
          icon="pi pi-book"
          size="small"
          rounded
          @click="onOpenLogs"
        />
        <Button
          label="Shell"
          icon-pos="left"
          icon="pi pi-code"
          size="small"
          rounded
          @click="openShellModal"
        />
      </div>

      <!-- Shell dialog -->
      <Dialog
        v-model:visible="isShellModalOpen"
        :header="service.name + ' shell'"
        maximizable
        modal
      >
        <ServiceShellDialog
          :data="shellData"
          @send="sendShellData"
        />
      </Dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { type PropType, ref } from 'vue'
import type { IServiceState } from '~/config/interfaces'
import type { Project, Service } from '~/schema/schema'
import ServiceShellDialog from '~/components/dialogs/service-shell-dialog.vue'
import LogsDialog from '~/components/dialogs/LogsDialog.vue'
import { ModalConfig } from '~/config/dialog-props'
import type { ILogsDialogData } from '~/interfaces/dialog-interfaces'

const props = defineProps({
  service: {
    required: true,
    type: Object as PropType<Service>,
  },
  serviceState: {
    required: true,
    type: Object as PropType<Record<string, IServiceState>>,
  },
  project: {
    required: true,
    type: Object as PropType<Project>,
  },
})

const toast = useToast()
const dialog = useDialog()

const isShellModalOpen = ref(false)

const shellData = ref()
let sendShellData: (
  data: string | ArrayBuffer | Blob,
  useBuffer?: boolean,
) => boolean

const { startServiceShell } = useService(ref(props.project))

function onOpenLogs() {
  if (!props.project.upn || !props.service.usn) {
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: 'Unable to stream logs.',
    })
    return
  }

  dialog.open(LogsDialog, {
    props: {
      header: `${props.service.name} logs`,
      ...ModalConfig,
    },
    data: {
      project: props.project,
      service: props.service,
    } as ILogsDialogData,
  })
}

function openShellModal() {
  if (!props.project.id || !props.service.name) {
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: 'Unable to connect to the shell.',
    })
    return
  }
  isShellModalOpen.value = true
  const { data, send } = startServiceShell(
    props.project.id,
    props.service.usn!,
  )
  shellData.value = data
  sendShellData = send
}
</script>
