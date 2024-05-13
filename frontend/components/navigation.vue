<script setup>
const user = useState("user");
const links = ref([]);
const router = useRouter();
const config = useRuntimeConfig();
const toast = useToast();
const confirm = useConfirm();

const { checkInvitation, acceptInvitation } = useOrganisationInviation();

onMounted(() => {
  if (checkInvitation()) {
    confirm.require({
      header: "Accept invitation?",
      message:
        "You were invited to a new Group do you wanna participate to the Group?",
      accept: () => acceptInvitation(),
      acceptLabel: "Accept",
      rejectLabel: "Cancel",
    });
  } 
});

function logOut() {
  $fetch(`${config.public.backendHost}/v1/auth/logout/github`, {
    credentials: "include",
    server: false,
    lazy: true,
  })
    .then(() => {
      router.push("/auth");
    })
    .catch((e) => {
      console.error(e);
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Unable to log out user",
      });
    });
}

watchEffect(() => {
  links.value = [
    {
      label: user.value?.nickname,
      avatar: {
        src: user.value?.avatar_url,
      },
      badge: "github",
    },
    {
      label: "Projects",
      icon: "i-heroicons-home",
      to: "/project",
    },
    {
      label: "Organisations",
      icon: "i-heroicons-user-group",
      to: "/organisation",
    },
    {
      label: "Logout",
      icon: "i-heroicons-arrow-left-on-rectangle",
      click: () => {
        confirm.require({
          header: "Logging out?",
          message: "Are you sure you want to log out from sloth?",
          accept: () => logOut(),
          acceptLabel: "Logout",
          rejectLabel: "Cancel",
        });
      },
    },
  ];
});
</script>
<template>
  <div class="hidden lg:block border border-gray-200 dark:border-gray-700 border-t-0 border-b-0 relative pt-5 px-2 h-full">
    <div class="flex gap-2 items-center pb-4">
      <Avatar :image="user.avatar_url" shape="circle" class="p-1" />
      <span>{{ user.nickname }}</span>
      <Badge value="github" severity="secondary" />
    </div>
    <div v-for="link in links">
      <NuxtLink :to="link.to">
        <Button
          text
          severity="secondary"
          class="flex gap-2 items-center w-full"
          @click="link.click"
        >
          <Icon v-if="link.icon" :icon="link.icon" style="font-size: 20px" />
          <span>{{ link.label }}</span>
        </Button>
      </NuxtLink>
    </div>
  </div>
</template>
