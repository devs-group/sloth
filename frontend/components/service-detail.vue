<template>
  <div>
    <Tag
      v-if="!service?.usn"
      severity="secondary"
      value="Not deployed yet"
    />
    <div v-else>
      <div class="flex items-center gap-2">
        <Tag
          :severity="badge.severity"
          :value="badge.label"
        />
        <span
          v-if="currentState"
          class="text-xs text-prime-secondary-text"
        >{{ currentState.status }}</span>
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
        <Button
          v-if="publicUrl"
          as="a"
          :href="publicUrl"
          target="_blank"
          rel="noopener"
          label="Open"
          icon-pos="left"
          icon="pi pi-external-link"
          size="small"
          rounded
          severity="secondary"
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
import { computed, type PropType, ref } from 'vue'
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
  deploying: {
    type: Boolean,
    default: false,
  },
})

const toast = useToast()
const dialog = useDialog()

const currentState = computed(() => props.service.usn ? props.serviceState[props.service.usn] : undefined)

// Docker reports state (running, exited, restarting, created, ...) and a status text like "Up 5 seconds (healthy)".
const badge = computed(() => {
  const s = currentState.value
  if (props.deploying) return { severity: 'info', label: 'Deploying' }
  if (!s) return { severity: 'secondary', label: 'Stopped' }
  if (s.state === 'running') {
    if (s.status.includes('(health: starting)')) return { severity: 'warn', label: 'Starting' }
    if (s.status.includes('(unhealthy)')) return { severity: 'danger', label: 'Unhealthy' }
    return { severity: 'success', label: 'Running' }
  }
  if (s.state === 'restarting' || s.state === 'created') return { severity: 'warn', label: 'Starting' }
  return { severity: 'danger', label: s.state.charAt(0).toUpperCase() + s.state.slice(1) }
})

// Same host rule as the backend: custom hosts win, otherwise <usn>.<backend host>.
const config = useRuntimeConfig()
const publicUrl = computed(() => {
  const s = props.service
  if (!s.public.enabled || !s.usn) return null
  const customHost = s.public.hosts.find(h => h.trim())
  const host = customHost ?? `${s.usn}.${new URL(config.public.backendHost).hostname}`
  return `${window.location.protocol}//${host}`
})

const isShellModalOpen = ref(false)

const shellData = ref()
let sendShellData: (
  data: string | ArrayBuffer | Blob,
  useBuffer?: boolean,
) => boolean = () => false

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
