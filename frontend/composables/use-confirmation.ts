const state = ref({
    show: false,
    title: "",
    description: "",
    hasConfirmed: false,
    params: Object,
    confirm: () => {
        state.value.hasConfirmed = true
    }
})

export default function() {
    return {
        state,
        showConfirmation: (title: string, description: string, params: any = {}) => {
            state.value.title = title
            state.value.description = description
            state.value.show = true
            state.value.params = params
        },
        onConfirm: (cb: Function) => {
            watchEffect(() => {
                if (state.value.hasConfirmed) {
                    state.value.hasConfirmed = false
                    state.value.show = false
                    cb(state.value.params)
                }
            })
        }
    }
}
