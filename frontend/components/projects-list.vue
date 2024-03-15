<script lang="ts" setup>
import type {Project} from "~/schema/schema";

const config = useRuntimeConfig()
const { data } = loadProjects()
const toast = useToast()
const confirm = useConfirm()

interface ProjectState {
    isDeploying?: boolean
    isRemoving?: boolean
}
const state = ref<Record<number, ProjectState>>({})

function loadProjects() {
  return useFetch<Project[]>(`${config.public.backendHost}/v1/projects`, { server: false, lazy: true, credentials: "include" })
}

function deploy(id: number, hook: string, accessToken: string) {
    state.value[id] = {
      isDeploying: true
    }
    $fetch(hook, {
        method: "GET",
        headers: {
            "X-Access-Token": accessToken
        }
    })
    .then(() => {
      toast.add({
        severity: "success",
        summary: "Success",
        detail: "Project has been deployed successfully",
        life: 3000
      })
    })
    .catch((e) => {
        console.error(e)
        toast.add({
          severity: "error",
          summary: "Error",
          detail: "Failed to deploay project",
          life: 3000
        })
    })
    .finally(() => state.value[id].isDeploying = false)
}

function remove(id: number, upn: string) {
  state.value[id] = {
    isRemoving: true
  }
  $fetch(`${config.public.backendHost}/v1/project/${upn}`, {
    method: "DELETE",
    credentials: "include"
  })
  .then(() => {
    // Re-fetch projects after delete
    const { data: d } = loadProjects()
    data.value = d.value
    
    toast.add({
          severity: "success",
          summary: "Success",
          detail: "Project has been removed successfully",
          life: 3000
    })
  })
  .catch((e) => {
    console.error(e)
    toast.add({
          severity: "error",
          summary: "Error",
          detail: "Failed to delete project",
          life: 3000
        })
  })
  .finally(() => state.value[id].isRemoving = false)
}
</script>

<template>
    <div>
        <div class="p-6 flex flex-row items-end justify-between">
            <div>
                <h1 class="text-2xl">Projects</h1>
                <p class="text-sm text-gray-400">{{ data?.length }} Projects in your organisation</p>
            </div>
              <NuxtLink to="/project/new">
                <IconButton severity="secondary" icon="heroicons:pencil-square" label="New Project"/>
              </NuxtLink>
        </div>

        <div>
            <div v-for="d in data as Project[]" class="p-6 flex flex-row flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700">
                <div class="flex flex-row items-center">
                    <div class="w-2/3">
                        <p>{{ d.name }}</p>
                        <p class="text-xs text-prime-secondary-text">UPN: {{ d.upn }}</p>
                        <p class="text-xs text-prime-secondary-text">Hook URL: {{ d.hook }}</p>
                        <p class="text-xs text-prime-secondary-text">Access token: {{ d.access_token }}</p>
                    </div>
                </div>
                <div class="space-x-4 flex flex-row items-center">
                  <IconButton
                    :loading="state[d.id]?.isRemoving"
                    text
                    severity="danger"
                    icon="heroicons:trash"
                    @click="
                        () => confirm.require({
                          header: 'Remove the project?',
                          message: 'After you you have removed the project, you won\'t be able to restore it.',
                          accept: () => remove(d.id as number, d.upn as string),
                          acceptLabel: 'Remove',
                          rejectLabel: 'Cancel'

                    })"
                  />
                    <NuxtLink :to="'project/' + d.upn">
                      <IconButton icon="heroicons:arrow-right-on-rectangle"/>
                    </NuxtLink>
                    <IconButton
                        label="Deploy"
                        icon="heroicons:rocket-launch"
                        aria-label="Deploy"
                        :loading="state[d.id]?.isDeploying"
                        @click="deploy(d.id as number, d.hook as string, d.access_token as string)"
                    />
                </div>
            </div>
        </div>
    </div>
</template>