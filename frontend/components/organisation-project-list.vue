<template>
  <OrganisationHeader :props="{ organisation_name: props.organisation.organisation_name, button: props.button }" />

  <div class="flex flex-col gap-2 px">
    <div v-if="props.projects && props.projects.length > 0">
      <template
        v-for="project in props.projects"
        :key="project.upn"
      >
        <OrganisationProjectRow
          :organisation-id="props.organisation.id"
          :project="project"
          @on-delete="props.emits.onDelete()"
        />
      </template>
    </div>
    <div
      v-else
      class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-4 p-6 border-t border-gray-200 dark:border-gray-700"
    >
      <OverlayProgressSpinner :show="props.isLoading" />
      <p v-if="!props.isLoading">
        No projects found.
      </p>
    </div>
  </div>
</template>

<script lang="ts" setup>
import OrganisationProjectRow from './rows/organisation-project-row.vue'
import OverlayProgressSpinner from './overlay-progress-spinner.vue'
import type { Organisation, OrganisationProject } from '~/schema/schema'

defineProps({
  props: {
    required: true,
    type: Object as PropType<{ projects: OrganisationProject[], organisation: Organisation, isLoading: boolean, button: { label: string, icon: string, onClick: () => void }, emits: { onDelete: () => void } }>,
  },
})
</script>
