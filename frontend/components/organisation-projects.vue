<script lang="ts" setup>
import type { OrganisationProject } from "~/schema/schema";
const confirm = useConfirm();
const toast = useToast();
const route = useRoute();

const organisation_name = route.params.organisation_name;
const g = ref<OrganisationProject[]>();
const isAddGroupProjectModalOpen = ref(false);
const projectUPN = ref("");

const config = useRuntimeConfig();

async function fetchOrganisationProjects() {
  try {
    g.value = await $fetch<OrganisationProject[]>(
      `${config.public.backendHost}/v1/organisation/${organisation_name}/projects`,
      { credentials: "include" }
    );
    console.log(g.value);
  } catch (e) {
    console.error("unable to fetch Organisation", e);
  }
}

onMounted(() => {
  fetchOrganisationProjects();
});

async function addProject() {
  try {
    g.value = await $fetch(
      `${config.public.backendHost}/v1/organisation/project`,
      {
        method: "PUT",
        credentials: "include",
        body: {
          id: organisation_name,
          upn: projectUPN.value,
        },
      }
    );
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Project added to group",
    });
  } catch (e) {
    console.error("unable to invite", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to add Project",
    });
  } finally {
    isAddGroupProjectModalOpen.value = false;
    fetchOrganisationProjects();
  }
}

async function removeProject(upn: string) {
  try {
    g.value = await $fetch(
      `${config.public.backendHost}/v1/organisation/project`,
      {
        method: "DELETE",
        credentials: "include",
        body: {
          organisation_name: organisation_name,
          upn: upn,
        },
      }
    );
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Project removed from group",
    });
  } catch (e) {
    console.error("unable to invite", e);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to remove Project",
    });
  } finally {
    fetchOrganisationProjects();
  }
}
</script>