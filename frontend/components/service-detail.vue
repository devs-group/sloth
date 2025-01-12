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
          @click="openLogsModal"
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
      <!-- Logs dialog -->
      <Dialog
        v-model:visible="isLogsModalOpen"
        :header="service.name + ' logs'"
        maximizable
        modal
      >
        <div class="overflow-auto h-[80vh]">
          <code
            v-for="line in logLines"
            :key="line"
            class="text-xs"
          >
            <span>{{ line }}</span>
          </code>
        </div>
      </Dialog>

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
const isLogsModalOpen = ref(false)
const isShellModalOpen = ref(false)

const shellData = ref()
const logLines = ref<string[]>([])
let sendShellData: (
  data: string | ArrayBuffer | Blob,
  useBuffer?: boolean,
) => boolean

const { streamServiceLogs, startServiceShell } = useService(ref(props.project))

function openLogsModal() {
  if (!props.project.upn || !props.service.usn) {
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: 'Unable to stream logs.',
    })
    return
  }
  isLogsModalOpen.value = true
  const { data } = streamServiceLogs(props.project.upn!, props.service.usn!)
  watch(data, (value) => {
    if (value) {
      logLines.value.push(value)
    }
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
