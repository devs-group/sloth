<script setup>
const user = useState("user")
const links = ref([])
const router = useRouter()
const config = useRuntimeConfig()
const { showError } = useNotification()

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
      icon: 'i-heroicons-home',
      click: () => {
        $fetch(`${config.public.backendHost}/v1/auth/logout/github`)
        .then(() => {
          router.push("/auth")
        })
        .catch((e) => {
          console.error(e)
          showError("Error", "Unable to log out user")
        })
      }
    },
 ]
})
</script>
<template>
    <div class="hidden lg:block h-[100vh] border border-gray-200 dark:border-gray-700 border-t-0 border-b-0 h-full relative pt-5 px-2">
        <UVerticalNavigation :links="links" />
    </div>
  
</template>