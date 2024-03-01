<script setup lang="ts">
definePageMeta({
  layout: "auth",
});

const router = useRouter();

interface UserResponse {
  user: {
    avatar_url: string;
    id: string;
    name: string;
    nickname: string;
  };
}

const config = useRuntimeConfig();
function openGithubPage() {
  window.open(`${config.public.backendHost}/v1/auth/github`, "_self");
}

const { hook } = useNuxtApp();

hook("page:finish", async () => {
  const p = new URLSearchParams(window.location.search);
  const inviteToken = new URLSearchParams(window.location.search).get("invite");
  if (inviteToken) {
    document.cookie = `inviteCode=${inviteToken}; path=/; max-age=86400`;
  }

  const c = p.get("code");
  const s = p.get("state");
  if (c && s) {
    const cbURL = `${config.public.backendHost}/v1/auth/github/callback?code=${c}&state=${s}`;
    const res = await $fetch<UserResponse>(cbURL, {
      credentials: "include",
    }).catch((e) => console.error(e));
    if (res?.user.id) {
      useState("user", () => res.user);
      console.log("User has been logged in... redirecting to /");
      setTimeout(() => {
        router.push("/");
      }, 100);
    }
  }
});
</script>
<template>
  <div class="flex flex-col justify-center items-center h-full space-y-12">
    <p class="text-3xl font-bold">Log in to Sloth</p>
    <UButton @click="openGithubPage" size="xl">
      <img src="~/public/github-mark.svg" alt="GH" height="24" width="24" />
      Login with github
    </UButton>
  </div>
</template>
