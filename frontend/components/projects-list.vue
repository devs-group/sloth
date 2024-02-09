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
                <Button severity="secondary" class="flex gap-2">
                  <Icon icon="heroicons:pencil-square" style="font-size: 24px;"/>
                  <span>New Project</span>
                </Button>
              </NuxtLink>
        </div>

        <div>
            <div v-for="d in data as Project[]" class="p-6 flex flex-row flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700">
                <div class="flex flex-row items-center">
                    <UAvatar :alt="d.name" size="sm" class="mr-3"/>
                    <div class="w-2/3">
                        <p>{{ d.name }}</p>
                        <p class="text-xs text-gray-400">UPN: {{ d.upn }}</p>
                        <p class="text-xs text-gray-400">Hook URL: {{ d.hook }}</p>
                        <p class="text-xs text-gray-400">Access token: {{ d.access_token }}</p>
                        <div class="relative">
                            <UPopover class="mt-2">
                                <UButton color="white" :label="`${d.services.length} services`" trailing-icon="i-heroicons-chevron-down-20-solid" />
                                <template #panel>
                                <div class="w-full p-4">
                                    <div v-for="s in d.services" class="p-4">
                                        <p class="text-sm text-gray-400">Service: {{ s.name }}</p>
                                        <p class="text-sm text-gray-400">Image: {{ s.image }}</p>
                                        <p class="text-sm text-gray-400">Ports: {{ s.ports.join(", ") }}</p>
                                        <div v-if="s.env_vars?.length > 0" class="text-sm text-gray-400">
                                          Env variables:
                                          <p v-for="e in s.env_vars">- {{ `${e[0]}: ${e[1]}`  }}</p>
                                        </div>

                                        <hr class="mt-4" />
                                    </div>
                                </div>
                                </template>
                            </UPopover>
                        </div>
                    </div>
                </div>
                <div class="space-x-4 flex flex-row items-center">
                  <Button
                      :loading="state[d.id]?.isRemoving"
                      text
                      severity="danger"
                      @click="
                        () => confirm.require({
                          header: 'Remove the project?',
                          message: 'After you you have removed the project, you won\'t be able to restore it.',
                          accept: () => remove(d.id as number, d.upn as string),
                          acceptLabel: 'Remove',
                          rejectLabel: 'Cancel'
                        })
                  ">
                  <Icon icon="heroicons:trash" style="font-size: 24px;"/>
                  </Button>
                    <NuxtLink :to="'project/' + d.upn">
                      <Button>
                        <Icon icon="heroicons:arrow-right-on-rectangle" style="font-size: 24px;"/> 
                      </Button>
                    </NuxtLink>
                    <Button
                        aria-label="Deploy"
                        :loading="state[d.id]?.isDeploying"
                        @click="deploy(d.id as number, d.hook as string, d.access_token as string)"
                        class="flex gap-2"
                    >
                    <Icon icon="heroicons:rocket-launch" style="font-size: 24px;"/> 
                    <span>Deploy</span>
                    </Button>
                </div>
            </div>
        </div>
    </div>
</template>