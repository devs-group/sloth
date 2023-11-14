<script setup lang="ts">
import { useRefHistory } from "@vueuse/core";
import {useWebSocket} from "@vueuse/core";
import {Terminal} from "xterm"

const props = defineProps({
    service: {
        type: String,
        required: true
    },
    upn: {
        type: String,
        required: true
    }
})

const config = useRuntimeConfig()
const { showError } = useNotification()

const containerRef = ref<HTMLDivElement>()
const currentLine = ref("")


const term = new Terminal({
    cursorBlink: true
})

const wsBackendHost = config.public.backendHost.replace("http", "ws")
const socket = useWebSocket(`${wsBackendHost}/v1/ws/project/shell/${props.service}/${props.upn}`, {
    immediate: false,
    autoReconnect: {
      retries: 5,
      delay: 1000,
      onFailed() {
        showError("Error", "unable to start shell")
      },
    },
})

function handleKeyPress(key: string, domEvent: KeyboardEvent) {
    switch(domEvent.code) {
        case "Enter":
            currentLine.value += key
            socket.send(currentLine.value)
            currentLine.value = ""
        break;
        case "Backspace":
            if (currentLine.value.length !== 0) {
                currentLine.value = currentLine.value.slice(0,-1)
                term.write("\b \b")
            }
        break;
        default:
            currentLine.value += key
            term.write(key)
    }
}

watchEffect(() => {
    if (socket.data.value) {
        let data = socket.data.value
        console.log(data)
        term.write(data)
    }
})

onMounted(() => {
    if (containerRef.value) {
        term.open(containerRef.value)
        term.clear()      
        socket.open()

        term.onKey((event) => {
            handleKeyPress(event.key, event.domEvent)
        })
    }
})



onUnmounted(() => {
    socket.close()
})

</script>
<template>
    <div class="overflow-scroll" ref="containerRef"></div>
</template>
<style>
    .xterm-helpers > span.xterm-char-measure-element {
        opacity: 0;
    }

    .xterm-helpers > textarea.xterm-helper-textarea {
        opacity: 0;
    }
</style>