<template>
  <div class="flex flex-col gap-2 w-full">
    <TabMenu :model="tabItems" class="w-full"/>
    <div class="flex flex-col gap-2 px-4">
      <p class="text-lg">OrgName Hier Setzen</p>
      <NuxtPage/>
    </div>
  </div>
</template>

<script setup lang="ts">
import {Routes} from "~/config/routes";

const router = useRouter()
const route = useRoute()

const tabItems = [
  {
    label: "Projects",
    command() {
      router.push({name: Routes.ORGANISATION, params: {id: route.params.id}})
    },
  },
  {
    label: "Members",
    command() {
      router.push({name: Routes.ORGANISATION_MEMBERS, params: {id: route.params.id} })
    },
  },
  {
    label: "Invitations",
    command() {
      console.log(Routes.ORGANISATIONS)
    },
  },
  {
    label: "Monitoring (coming soon)",
    disabled: true,
  },
] as TabItem[];

const onTabClick = (tab: TabItem) => {
  console.log(tab.to)
}
</script>