import { ref } from 'vue';
import { useToast } from 'primevue/usetoast';
import { Constants } from '~/config/const';
import type { Project, ProjectSchema } from '~/schema/schema';

export function useProject() {
  const config = useRuntimeConfig();
  const toast = useToast();

  const isLoading = ref(false);
  const isUpdatingLoading = ref(false);
  const pageErrorMessage = ref('');

  async function updateProject(project: Project) {
    isUpdatingLoading.value = true;
    try {
      await $fetch(`${config.public.backendHost}/v1/project/${project.id}`, {
        method: "PUT",
        credentials: "include",
        body: project,
      });
      await fetchProject(project.id);
      toast.add({
        severity: "success",
        summary: "Success",
        detail: "Project has been updated",
        life: Constants.ToasterDefaultLifeTime,
      });
    } catch (e) {
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Unable to update project",
        life: Constants.ToasterDefaultLifeTime,
      });
    } finally {
      isUpdatingLoading.value = false;
    }
  }

  async function fetchProject(id: number) : Promise<ProjectSchema | null> {
    isLoading.value = true;
    try {
        return await $fetch<Project|null>(`${config.public.backendHost}/v1/project/${id}`, {
        credentials: "include",
      });
    } catch (error) {
      pageErrorMessage.value = "Sorry we can't find this project";
      toast.add({
        severity: "error",
        summary: "Error",
        detail: pageErrorMessage.value,
        life: Constants.ToasterDefaultLifeTime,
      });
      return null
    } finally {
      isLoading.value = false;
    }
  }

  return {
    isLoading,
    isUpdatingLoading,
    updateProject,
    fetchProject,
    pageErrorMessage
  };
}
