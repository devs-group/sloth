<template>
  <div v-if="props.projects && props.projects.length > 0">
    <template v-for="project in props.projects" :key="project.upn">
      <OrganisationProjectRow :organisation="props.organisation.id" :project="project"/>
    </template>
  </div>
  <div v-else>
    <OverlayProgressSpinner :show="props.isLoading"/>
    <p v-if="!props.isLoading">No projects found.</p> 
  </div>
</template>
<script lang="ts" setup>
import type { Organisation, OrganisationProject } from "~/schema/schema";
import OrganisationProjectRow from "./rows/organisation-project-row.vue";
import OverlayProgressSpinner from "./overlay-progress-spinner.vue";
defineProps({
  props: {
    required: true,
    type: Object as PropType<{ projects: OrganisationProject[], organisation: Organisation, isLoading: boolean }>,
  },
});
</script>
