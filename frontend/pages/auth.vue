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

var socialLogins = {
  github: () => {
    window.open(`${config.public.backendHost}/v1/auth/github`, "_self");
  },
  google: () => {
    window.open(`${config.public.backendHost}/v1/auth/google`, "_self");
  },
};

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
    const cbURL = `${config.public.backendHost}/v1/auth/google/callback?code=${c}&state=${s}`;
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
  <div class="flex flex-col justify-center items-center flex-1 gap-6">
    <p class="text-3xl font-bold">Log in to Sloth</p>
    <Button @click="socialLogins.github()" class="flex gap-2 items-center">
      <img src="/github-mark.svg" alt="GH" height="24" width="24" />
      Login with github
    </Button>
    <Button @click="socialLogins.google()" class="flex gap-2 items-center">
      <img src="/google-mark.svg" alt="Google" height="24" width="24" />
      Login with google
    </Button>
  </div>
</template>
