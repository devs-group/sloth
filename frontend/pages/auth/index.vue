<template>
  <div class="flex flex-col justify-center items-center flex-1 gap-6">
    <p class="text-3xl font-bold">Log in to Sloth</p>
    <Button @click="socialLogins.github()" class="flex gap-2 items-center">
      <img src="~/assets/svg/github-mark.svg" alt="GH" height="24" width="24"/>
      Login with Github
    </Button>
    <Button @click="socialLogins.google()" class="flex gap-2 items-center">
      <img src="~/assets/svg/google-mark.svg" alt="Google" height="24" width="24"/>
      Login with Google
    </Button>
  </div>
</template>

<script setup lang="ts">
import {Constants} from "~/config/const";

const route = useRoute();
const toast = useToast()
const config = useRuntimeConfig();
const {showGlobalSpinner, hideGlobalSpinner} = useGlobalSpinner()

const socialLogins = {
  github: () => {
    window.open(`${config.public.backendHost}/v1/auth/github`, "_self");
  },
  google: () => {
    window.open(`${config.public.backendHost}/v1/auth/google`, "_self");
  },
};

const provider = route.params.provider as string
const code = route.query.code
const state = route.query.state
const invitationToken = route.query.invite

if (code && state && provider) {
  showGlobalSpinner()
  const cbURL = `${config.public.backendHost}/v1/auth/${provider}/callback?code=${code}&state=${state}`;
  $fetch<OAuthUserResponse>(cbURL, {
    credentials: "include",
  })
      .then(payload => {
        if (payload.user.id) {
          reloadNuxtApp({force: true})
        }
      })
      .catch((e) => {
        hideGlobalSpinner()
        toast.add({
          severity: "error",
          summary: "Error",
          detail: "Login failed, try another provider or try again",
          life: Constants.ToasterDefaultLifeTime,
        });
      })
} else if (invitationToken) {
  document.cookie = `inviteCode=${invitationToken}; path=/; max-age=86400`;
}
</script>
