<script setup lang="ts">
definePageMeta({
    layout: "auth"
})

const config = useRuntimeConfig()
function openGithubPage() {
    window.open(`${config.public.backendHost}/v1/auth/github`, "_self")
}

const router = useRouter()

interface UserResponse {
    user: {
        avatar_url: string
        id: string
        name: string
        nickname: string
    }
}

onMounted(async () => {
    if (typeof window !== "undefined" && window.document) {
        const p = new URLSearchParams(window.location.search)
        const c = p.get("code")
        const s = p.get("state")

        const cbURL = `${config.public.backendHost}/v1/auth/github/callback?code=${c}&state=${s}`
        if (c && s) {
            const res = await $fetch<UserResponse>(cbURL, {credentials: 'include'}).catch((e) => console.error(e))
            if (res?.user.id) {
                useState("user", () => res.user)
                router.push("/")
            }
        }
    }
})

</script>
<template>
    <div class="flex flex-row justify-center items-center h-full">
        <UButton @click="openGithubPage">
          Login with github
        </UButton>
    </div>
</template>