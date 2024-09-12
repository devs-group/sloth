<template>
        <div class="w-full">
            <div ref="containerRef"></div>
        </div>
</template>
<script setup lang="ts">
import {Terminal} from "@xterm/xterm"
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

import type { IDialogInjectRef, IServiceShellDialog } from "~/config/interfaces";

const dialogRef = inject<IDialogInjectRef<IServiceShellDialog>>('dialogRef');
const containerRef = ref<HTMLDivElement>()
const currentLine = ref("")


const terminal = new Terminal({
    cursorBlink: true,

});
const fitAddon = new FitAddon();

function handleKeyPress(key: string, domEvent: KeyboardEvent) {
    switch(domEvent.code) {
        case "Enter":
            currentLine.value += key
            dialogRef?.value.data.send(currentLine.value)

            terminal.write(backspace(currentLine.value.length -1))  
            currentLine.value = ""
        break;
        case "Backspace":
            if (currentLine.value.length !== 0) {
                currentLine.value = currentLine.value.slice(0,-1)
                terminal.write(backspace(1))
            }
        break;
        case "ArrowUp":
        break;
        case "ArrowDown":
        break;
        default:
            currentLine.value += key
            terminal.write(key)
    }   
}

function backspace(n: number) {
    const backspaces = '\b'.repeat(n);
    const spaces = ' '.repeat(n);
    const resetBackspaces = '\b'.repeat(n);

    return backspaces+spaces+resetBackspaces

}

onMounted(() => {
    if (containerRef.value) {
        terminal.loadAddon(fitAddon);   
        fitAddon.fit();
        terminal.open(containerRef.value)
        terminal.clear()

        terminal.onKey((event) => {
            handleKeyPress(event.key, event.domEvent)
        })
    }

    window.addEventListener("resize", fitAddon.fit)
})

onBeforeUnmount(() => {
    if (terminal) {
        terminal.dispose()
    }

    window.removeEventListener("resize", fitAddon.fit)
})

watchEffect(() => {
    let data = dialogRef?.value.data.data
    if(data) {
        terminal.write(data)
    }
})

</script>