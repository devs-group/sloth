<script lang="ts" setup>

import {Project, Service} from "~/schema/schema";

const route = useRoute()
const upn = route.params.upn
const config = useRuntimeConfig()
const { showError, showSuccess } = useNotification()

const p = ref<Project>()
const isUpdatingLoading = ref(false)
const isChangeProjectNameModalOpen = ref(false)

try {
  p.value = await $fetch<Project>(
      `${config.public.backendHost}/v1/project/${upn}`,
      { credentials: "include" },
  )
} catch (e) {
  console.error("unable to fetch project", e)
}


async function updateProject() {
  isUpdatingLoading.value = true
  try {
    await $fetch<Project>(
        `${config.public.backendHost}/v1/project/${upn}`,
        {
          method: "PUT",
          credentials: "include",
          body: p.value
        },
    )
    showSuccess("Success", "Project has been updated")
  } catch (e) {
    console.error("unable to update project", e)
    showError("Error", "Unable to update project")
  } finally {
    isUpdatingLoading.value = false
  }
}

function removeService(idx: number) {
  p.value?.services.splice(idx, 1)
}

function addEnv(serviceIdx: number) {
  p.value?.services[serviceIdx].env_vars.push(["",""])
}

function removeEnv(envIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].env_vars.splice(envIdx, 1)
}

function addVolume(serviceIdx: number) {
  p.value?.services[serviceIdx].volumes.push("")
}

function removeVolume(volIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].volumes.splice(volIdx, 1)
}

function addPort(serviceIdx: number) {
  p.value?.services[serviceIdx].ports.push("")
}

function removePort(portIdx: number, serviceIdx: number) {
  p.value?.services[serviceIdx].ports.splice(portIdx, 1)
}

function hookCurlCmd(url: string, accessToken: string) {
  return `curl -X GET "${url}" -H "X-Access-Token: ${accessToken}"`
}
</script>

<template>
  <div>
    <div v-if="p">
      <div class="py-12 px-6 space-y-4">
        <div>
          <p class="text-sm text-gray-500">Project name</p>
          <div class="flex flex-row items-center space-x-4">
            <p>{{ p.name }}</p>
            <UBadge class="cursor-pointer" @click="isChangeProjectNameModalOpen = true">Change</UBadge>
            <UModal v-model="isChangeProjectNameModalOpen">
              <div class="flex flex-row items-center w-full space-x-4 p-6">
                <UInput class="w-full" v-model="p.name"/>
                <UButton icon="i-heroicons-check" @click="isChangeProjectNameModalOpen = false"></UButton>
              </div>
            </UModal>
          </div>
        </div>

        <div>
          <p class="text-sm text-gray-500">Project unique name</p>
          <p>{{ p.upn }}</p>
        </div>

        <div>
          <p class="text-sm text-gray-500">Deployment webhook</p>
          <div class="flex flex-row items-center space-x-2">
            <p class="text-sm text-gray-500">URL:</p>
            <p>{{ p.hook }}</p>
            <CopyButton :string="p.hook as string"></CopyButton>
          </div>
          <div class="flex flex-row items-center space-x-2">
            <p class="text-sm text-gray-500">
              Access Token:
            </p>
            <p>{{ p.access_token }}</p>
            <CopyButton :string="p.access_token as string"></CopyButton>
          </div>
          <div class="flex flex-row items-center space-x-2">
            <code class="text-sm text-gray-500">
              {{ hookCurlCmd(p.hook as string, p.access_token as string) }}
            </code>
            <CopyButton :string="hookCurlCmd(p.hook as string, p.access_token as string)"></CopyButton>
          </div>
        </div>

        <div class="pt-6 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-12">
          <ServiceForm
              v-for="(s, idx) in Object.values(p.services)"
              :service="s as Service"
              :index="idx"
              @add-env="addEnv"
              @remove-env="removeEnv"
              @remove-service="removeService"
              @add-volume="addVolume"
              @remove-volume="removeVolume"
          ></ServiceForm>
        </div>

        <UButton @click="updateProject" :loading="isUpdatingLoading">Save</UButton>
      </div>
    </div>
  </div>
</template>