<script lang="ts" setup>
import { Project } from "~/components/projects-list.vue";

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
  p.value.services.splice(idx, 1)
}

function addEnv(serviceIdx: number) {
  p.value.services[serviceIdx].env_vars.push(["",""])
}

function removeEnv(serviceIdx: number, envIdx: number) {
  p.value.services[serviceIdx].env_vars.splice(envIdx, 1)
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
            <CopyButton :string="p.hook"></CopyButton>
          </div>
          <div class="flex flex-row items-center space-x-2">
            <p class="text-sm text-gray-500">
              Access Token:
            </p>
            <p>{{ p.access_token }}</p>
            <CopyButton :string="p.access_token"></CopyButton>
          </div>
          <div class="flex flex-row items-center space-x-2">
            <code class="text-sm text-gray-500">
              {{ hookCurlCmd(p.hook, p.access_token) }}
            </code>
            <CopyButton :string="hookCurlCmd(p.hook, p.access_token)"></CopyButton>
          </div>
        </div>

        <div class="pt-6 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-12">
          <div v-for="(s, idx) in Object.values(p.services)" class="space-y-4 py-3">
            <UFormGroup label="Name" :name="`services[${idx}].name`">
              <UInput v-model="s.name" type="text" />
            </UFormGroup>
            <UFormGroup label="Port" :name="`services[${idx}].port`">
              <UInput v-model="s.ports[0]" type="text" />
            </UFormGroup>
            <UFormGroup label="Image" :name="`services[${idx}].image`">
              <UInput v-model="s.image" type="text" />
            </UFormGroup>
            <UFormGroup label="Image tag" :name="`services[${idx}].image_tag`">
              <UInput v-model="s.image_tag" type="text" />
            </UFormGroup>
            <UFormGroup>
              <div class="flex flex-row justify-between items-center">
                <p class="text-sm">Publicly exposed</p>
                <UToggle v-model="s.public.enabled" />
              </div>
            </UFormGroup>
            <div v-if="s.public.enabled" class="space-y-4">
              <UFormGroup label="Host" :name="`services[${idx}].public.host`">
                <UInput v-model="s.public.host" type="text" />
              </UFormGroup>
              <UFormGroup>
                <div class="flex flex-row justify-between items-center">
                  <p class="text-sm">SSL</p>
                  <UToggle v-model="s.public.ssl" />
                </div>
              </UFormGroup>
              <UFormGroup>
                <div class="flex flex-row justify-between items-center">
                  <p class="text-sm">Compress</p>
                  <UToggle v-model="s.public.compress" />
                </div>
              </UFormGroup>
            </div>
            <UFormGroup label="Environment variables" class="pt-4">
              <div class="flex flex-col space-y-2">
                <div v-for="(env, envIdx) in s.env_vars as string[]" class="flex space-x-2">
                  <UInput placeholder="Key" v-model="env[0]"></UInput>
                  <UInput placeholder="Value" v-model="env[1]"></UInput>
                  <UButton
                      v-if="envIdx === (s.env_vars as string[]).length-1"
                      icon="i-heroicons-plus"
                      variant="ghost"
                      :ui="{ rounded: 'rounded-full' }"
                      @click="() => addEnv(idx)"
                      :disabled="env[0] === '' || env[1] === ''"
                  />
                  <UButton
                      v-else
                      icon="i-heroicons-minus"
                      variant="ghost"
                      color="red"
                      :ui="{ rounded: 'rounded-full' }"
                      @click="() => removeEnv(idx, envIdx)"
                  />
                </div>
              </div>
            </UFormGroup>
            <div>
              <p class="text-xs text-red-400 cursor-pointer p-2 text-center" @click="removeService(idx)">Remove</p>
            </div>
          </div>
        </div>

        <UButton @click="updateProject" :loading="isUpdatingLoading">Save</UButton>
      </div>
    </div>
  </div>
</template>