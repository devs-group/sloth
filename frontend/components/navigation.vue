<script setup>
const user = useState("user")
const links = ref([])
const router = useRouter()
const config = useRuntimeConfig()
const toast = useToast()
const confirm = useConfirm()

function logOut() {
  $fetch(`${config.public.backendHost}/v1/auth/logout/github`, {
    credentials: "include",
    server: false,
    lazy: true,
  })
  .then(() => {
    router.push("/auth")
  })
  .catch((e) => {
    console.error(e)
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "Unable to log out user"
    })
  })
}

watchEffect(() => {
  links.value = [
    {
      label: 'Projects',
      icon: 'heroicons:home',
      click: () => navigateTo("/")
    },
    {
      label: 'Logout',
      icon: 'heroicons:arrow-left-on-rectangle',
      click: () => {
        confirm.require({
          header: "Logging out?",
          message: "Are you sure you want to proceed?",
          acceptLabel: "Confirm",
          rejectLabel: "Cancel",
          accept: () => logOut()
        })
      }
    },
 ]
})
</script>
<template>
    <div class="hidden lg:block border border-gray-200 dark:border-gray-700 border-t-0 border-b-0 relative pt-5 px-2 h-full">
        <div class="flex gap-2 items-center pb-4">
          <Avatar :image="user.avatar_url" shape="circle" class="p-1"/>
          <span>{{ user.nickname }}</span>
          <Badge value="github" severity="secondary"/>
        </div>
        <div v-for="link in links">
          <Button text severity="secondary" class="flex gap-2 items-center w-full" @click="link.click">
            <Icon v-if="link.icon" :icon="link.icon" style="font-size: 20px;" />
            <span>{{ link.label }}</span>
          </Button>
        </div>
    </div>  
  
</template>