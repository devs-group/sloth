<template>
  <div class="flex flex-col gap-2">
    <div class="flex flex-wrap lg:flex-nowrap justify-between items-center gap-2">
      <div class="flex flex-col gap-1 max-w-full">
        <p class="text-sm text-prime-secondary-text">Project Name</p>
        <p class="break-all">{{ project.name }}</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <Button
          :loading="isUpdatingLoading"
          :disabled="isUpdatingAndRestartingLoading"
          label="Save"
          @click="updateProject"
        />
        <Button
          :loading="isUpdatingAndRestartingLoading"
          :disabled="isUpdatingLoading"
          label="Save & Restart"
          @click="updateAndRestartProject"
        />
      </div>
    </div>

    <div class="flex flex-col gap-1 max-w-full">
      <p class="text-sm text-prime-secondary-text">Project Unique Name</p>
      <div class="flex items-center gap-1">
        <p>{{ project.upn }}</p>
        <CopyButton v-if="project.upn" :string="project.upn" />
      </div>
    </div>

    <div v-if="project.services && project.services.find(s => s.public.enabled)" class="flex flex-col gap-1 max-w-full">
      <p class="text-sm text-prime-secondary-text">Public URLs</p>
      <div v-for="service in project.services.filter(s => s.public.enabled)" :key="service.usn">
        <template v-if="service.public.hosts.some(url => url.trim().length > 0)">
          <div v-for="host in service.public.hosts" :key="host" class="flex items-center gap-1">
            <Icon icon="heroicons:link" />
            <a :href="`//${host}`" target="_blank">{{ host }}</a>
            <CopyButton :string="host" />
          </div>
        </template>
        <p v-else>No public URLs yet.</p>
      </div>
    </div>

    <div class="flex flex-col gap-1 max-w-full">
      <p class="text-sm text-prime-secondary-text">Deployment Webhook</p>
      <div class="flex flex-wrap gap-1 items-center">
        <p class="break-words">{{ project.hook }}</p>
        <CopyButton :string="project.hook" />
      </div>

      <p class="text-sm text-prime-secondary-text">Deployment Access Token</p>
      <div class="flex gap-1 items-center">
        <p class="break-words">{{ project.access_token }}</p>
        <CopyButton :string="project.access_token" />
      </div>

      <p class="text-sm text-prime-secondary-text">Example Command</p>
      <div class="flex items-center">
        <code class="text-sm text-prime-secondary-text">
          {{ hookCurlCmd(project.hook, project.access_token) }}
        </code>
        <CopyButton :string="hookCurlCmd(project.hook, project.access_token)" />
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import type { PropType } from "vue";
import type { ProjectSchema} from "~/schema/schema";

defineProps({
  project: {
    type: Object as PropType<ProjectSchema>,
    required: true,
  },
  isUpdatingLoading: {
    type: Boolean,
    required: true,
    default: false,
  },
  isUpdatingAndRestartingLoading: {
    type: Boolean,
    required: true,
    default: false,
  }
})

function hookCurlCmd(url: string, accessToken: string) {
    return `curl -X GET "${url}" -H "X-Access-Token: ${accessToken}"`;
}

const emit = defineEmits<{
  (event: "updateProject"): void;
  (event: "updateAndRestartProject"): void;
}>();
function updateProject(){
  emit("updateProject")
}
function updateAndRestartProject(){
  emit('updateAndRestartProject')
}
</script>
