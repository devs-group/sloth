<template>
    <div class="" v-if="props" v-for="project in props">
        <OrganisationProjectRow :project="project"></OrganisationProjectRow>
    </div>
</template>

<script lang="ts" setup>
import type { OrganisationProject } from "~/schema/schema";
import OrganisationProjectRow from "./organisation-project-row.vue";
 
defineProps({
  props: {
    type: Object as PropType<OrganisationProject[] | null | undefined>,
    required: true,
  }
});
</script>