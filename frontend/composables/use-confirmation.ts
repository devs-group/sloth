const state = ref({
    show: false,
    title: "",
    description: "",
    hasConfirmed: false,
})

export default function() {
    return {
        state,
        showConfirmation: (title: string, description: string, onConfirm: Function) => {
            state.value.title = title
            state.value.description = description
            state.value.show = true
            const unwatch = watchEffect(() => {
                if (state.value.hasConfirmed) {
                    state.value.hasConfirmed = false
                    state.value.show = false
                    onConfirm()
                    unwatch()
                }
            })
        },
    }
}
