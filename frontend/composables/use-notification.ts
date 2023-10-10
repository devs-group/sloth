const state = ref({
    show: false,
    title: "",
    message: "",
    id: 1,
    color: "",
    icon: ""
})

export default function() {
    return {
        state,
        showError: (title: string, message: string) => {
            state.value.title = title
            state.value.message = message
            state.value.id = state.value.id + 1
            state.value.icon = "i-heroicons-exclamation-circle"
            state.value.color = "red"
            state.value.show = true
        },
        showSuccess: (title: string, message: string) => {
            state.value.title = title
            state.value.message = message
            state.value.id = state.value.id + 1
            state.value.icon = "i-heroicons-check-badge"
            state.value.color = "primary"
            state.value.show = true
        }
    }
}
