<script setup lang="ts">
definePageMeta({
    layout: "auth",
})

const router = useRouter()

interface UserResponse {
    user: {
        avatar_url: string
        id: string
        name: string
        nickname: string
    }
}

const config = useRuntimeConfig()
function openGithubPage() {
    window.open(`${config.public.backendHost}/v1/auth/github`, "_self")
}

const { hook } = useNuxtApp()

hook("page:finish", async () => {
    const p = new URLSearchParams(window.location.search)
    const c = p.get("code")
    const s = p.get("state")
    if (c && s) {
        const cbURL = `${config.public.backendHost}/v1/auth/github/callback?code=${c}&state=${s}`
        const res = await $fetch<UserResponse>(cbURL, {credentials: 'include'}).catch((e) => console.error(e))
        if (res?.user.id) {
            useState("user", () => res.user)
            console.log("User has been logged in... redirecting to /")
            setTimeout(() => {
                router.push("/")
            }, 100)
        }
    }
})

</script>
<template>
    <div class="flex flex-col justify-center items-center flex-1 gap-6">
        <p class="text-3xl font-bold">Log in to Sloth</p>
        <Button @click="openGithubPage" class="flex gap-2 items-center">
          <img src="/github-mark.svg" alt="GH" height="24" width="24"/>
          Login with github
        </Button>
    </div>
</template>
