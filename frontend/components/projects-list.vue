<script lang="ts" setup>
interface Service {
    image: string
    name: string
    ports: string[]
    env_vars: string[] 
    state: string
    status: string
}

interface Project {
    id: number
    name: string
    upn: string
    access_token: string
    hook: string
    services: Record<string, Service>
}
const config = useRuntimeConfig()
const { data } = useFetch<Project[]>(`${config.public.backendHost}/v1/projects`, { server: false, lazy: true, credentials: "include" })

interface ProjectState {
    isDeploying: boolean
}
const state = ref<Record<number, ProjectState>>({})
const { showError, showSuccess } = useNotification()

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
        showSuccess("Success", "Project has been deployed succesfully")
    })
    .catch((e) => {
        console.error(e)
        showError("Error", "Failed to deploy project")
    })
    .finally(() => state.value[id].isDeploying = false)
}
</script>

<template>
    <div>
        <div class="p-6 flex flex-row items-end justify-between">
            <div>
                <h1 class="text-2xl">Projects</h1>
                <p class="text-sm text-gray-400">{{ data?.length }} Projects in your organisation</p>
            </div>
            <UButton icon="i-heroicons-pencil-square" size="sm" color="gray" variant="solid" :trailing="false">
                <NuxtLink to="/project/new">New Project</NuxtLink>
            </UButton>
        </div>

        <div>
            <div v-for="d in data" class="p-6 flex flex-row flex-1 items-center justify-between border border-1 border-x-0 border-gray-200 dark:border-gray-700">
                <div class="flex flex-row items-center">
                    <UAvatar :alt="d.name" size="sm" class="mr-3"/>
                    <div>
                        <p>{{ d.name }}</p>
                        <p class="text-xs text-gray-400">UPN: {{ d.upn }}</p>
                        <p class="text-xs text-gray-400">Hook URL: {{ d.hook }}</p>
                        <p class="text-xs text-gray-400">Access token: {{ d.access_token }}</p>
                        <div class="relative">
                            <UPopover class="mt-2">
                                <UButton color="white" :label="`${Object.keys(d.services).length} services`" trailing-icon="i-heroicons-chevron-down-20-solid" />
                                <template #panel>
                                <div class="w-full p-4">
                                    <div v-for="s in d.services" class="p-4">
                                        <p class="text-sm text-gray-400">Service: {{ s.name }}</p>
                                        <p class="text-sm text-gray-400">Image: {{ s.image }}</p>
                                        <p class="text-sm text-gray-400">Ports: {{ s.ports.join(", ") }}</p>
                                        <p class="text-sm text-gray-400">State: {{ s.state }}</p>
                                        <p class="text-sm text-gray-400">Status: {{ s.status }}</p>
                                        <p v-if="s.env_vars?.length > 0" class="text-sm text-gray-400">Env variables: {{ s.env_vars.join(", ") }}</p>
                                        <hr class="mt-4" />
                                    </div>
                                </div>
                                </template>
                            </UPopover>
                        </div>
                    </div>
                </div>
                <div>
                    <UButton icon="i-heroicons-rocket-launch" :loading="state[d.id]?.isDeploying" @click="deploy(d.id, d.hook, d.access_token)">Deploy</UButton>
                </div>
            </div>
        </div>
    </div>
</template>