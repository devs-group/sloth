import { ref } from 'vue';
import { useToast } from 'primevue/usetoast';
import { useRouter } from 'vue-router';
import { Constants } from '~/config/const';
import type { Project } from '~/schema/schema';

export function useProject(id: string) {
  const config = useRuntimeConfig();
  const router = useRouter();
  const toast = useToast();

  const project = ref<Project| null>(null);
  const isLoading = ref(false);
  const isUpdatingLoading = ref(false);
  const pageErrorMessage = ref('');

  async function updateProject() {
    isUpdatingLoading.value = true;
    try {
      await $fetch(`${config.public.backendHost}/v1/project/${id}`, {
        method: "PUT",
        credentials: "include",
        body: project.value,
      });
      await fetchProject();
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

  async function fetchProject() {
    isLoading.value = true;
    try {
      const payload = await $fetch<Project | null>(`${config.public.backendHost}/v1/project/${id}`, {
        credentials: "include",
      });
      project.value = payload as Project | null;
    } catch (error) {
      pageErrorMessage.value = "Sorry we can't find this project";
      router.push('/error'); // Redirect or handle errors as needed
      toast.add({
        severity: "error",
        summary: "Error",
        detail: pageErrorMessage.value,
        life: Constants.ToasterDefaultLifeTime,
      });
    } finally {
      isLoading.value = false;
    }
  }

  return {
    project,
    isLoading,
    isUpdatingLoading,
    updateProject,
    fetchProject,
    pageErrorMessage
  };
}
