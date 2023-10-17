<script lang="ts" setup>
import { Project } from "~/components/projects-list.vue";

const route = useRoute()
const upn = route.params.upn
const config = useRuntimeConfig()
const { data, pending, error } = useFetch<Project>(
    `${config.public.backendHost}/v1/project/${upn}`,
    { server: false, lazy: true, credentials: "include" },
)


</script>

<template>
  <div>
    <div v-if="error">Error: {{ error }}</div>
    <div v-else-if="pending">Loading...</div>
    <div v-else>
      <div>{{ data }}</div>
    </div>
  </div>
</template>