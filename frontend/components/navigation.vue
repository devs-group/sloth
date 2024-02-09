<script setup>
const user = useState("user")
const links = ref([])
const router = useRouter()
const config = useRuntimeConfig()
const toast = useToast()
const { showConfirmation } = useConfirmation()

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
      label: user.value?.nickname,
      avatar: {
        src: user.value?.avatar_url
      },
      badge: "github"
    },
    {
      label: 'Projects',
      icon: 'i-heroicons-home',
      to: '/'
    },
    {
      label: 'Logout',
      icon: 'i-heroicons-arrow-left-on-rectangle',
      click: () => {
        showConfirmation(
            "Logging out?",
            "Are you sure you want to log out from sloth?",
            () => logOut(),
        )
      }
    },
 ]
})
</script>
<template>
    <div class="hidden lg:block border border-gray-200 dark:border-gray-700 border-t-0 border-b-0 relative pt-5 px-2">
        <UVerticalNavigation :links="links" />
    </div>
  
</template>