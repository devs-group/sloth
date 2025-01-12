const globalSpinner = ref(false)
const globalText = ref('')

export const useGlobalSpinner = () => {
  const showGlobalSpinner = (text?: string) => {
    globalSpinner.value = true
    globalText.value = text ?? ''
  }
  const hideGlobalSpinner = () => {
    globalSpinner.value = false
    globalText.value = ''
  }

  return {
    showGlobalSpinner, hideGlobalSpinner,
    globalSpinner, globalText,
  }
}
