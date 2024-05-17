<template>
  <div class="flex flex-1 justify-center items-center">
    <p>Authenticating...</p>
  </div>
</template>

<script setup lang="ts">
import {Routes} from "~/config/routes";
import {Constants} from "~/config/const";

const route = useRoute()
const toast = useToast()
const config = useRuntimeConfig();
const {showGlobalSpinner, hideGlobalSpinner} = useGlobalSpinner()

const provider = route.params.provider as string
const code = route.query.code
const state = route.query.state

if (code && state && provider) {
  showGlobalSpinner()
  const cbURL = `${config.public.backendHost}/v1/auth/${provider}/callback?code=${code}&state=${state}`;
  $fetch<OAuthUserResponse>(cbURL, {
    credentials: "include",
  })
      .then(async payload => {
        if (payload.user.id) {
          await navigateTo({name: Routes.AUTH, replace: true})
          reloadNuxtApp({force: true})
        }
      })
      .catch(async (e) => {
        hideGlobalSpinner()
        await navigateTo({name: Routes.AUTH, replace: true})
        toast.add({
          severity: "error",
          summary: "Error",
          detail: "Sign in has failed, sorry",
          life: Constants.ToasterDefaultLifeTime,
        });
      })
} else {
  await navigateTo({name: Routes.AUTH, replace: true})
  toast.add({
    severity: "error",
    summary: "Error",
    detail: "Invalid auth url, please try again",
  });
}
</script>