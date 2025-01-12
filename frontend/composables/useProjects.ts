import type { Project } from '~/schema/schema'
import { Constants } from '~/config/const'

export function useProjects() {
  const toast = useToast()
  const config = useRuntimeConfig()

  const isLoading = ref(false)

  async function loadProjects() {
    isLoading.value = true
    try {
      return $fetch<Project[]>(`${config.public.backendHost}/v1/projects`, { credentials: 'include' })
    }
    catch {
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: 'Sorry we can\'t find any projects',
        life: Constants.ToasterDefaultLifeTime,
      })
    }
    finally {
      isLoading.value = false
    }
  }

  return {
    isLoading,
    loadProjects,
  }
}
