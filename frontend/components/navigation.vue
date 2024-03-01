<script setup>
const user = useState("user");
const links = ref([]);
const router = useRouter();
const config = useRuntimeConfig();
const { showSuccess } = useNotification();
const { showError } = useNotification();
const { showConfirmation } = useConfirmation();

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
      showError("Error", "Unable to log out user");
    });
}

function checkInvitation() {
  const cookies = document.cookie.split("; ");
  const inviteCookie = cookies.find((cookie) =>
    cookie.startsWith("inviteCode=")
  );

  if (inviteCookie) {
    const inviteCode = inviteCookie.split("=")[1];
    console.log(`Invite code found: ${inviteCode}`);
    return inviteCode;
  } else {
    console.log("No invite code cookie found.");
    return null;
  }
}

function removeInvitationCookie(link) {
  console.log(link);
}

async function acceptInvitation() {
  const data = {
    user_id: user.value?.id,
    invitation_token: inviteCode,
  };

  try {
    await $fetch(`${config.public.backendHost}/v1/group/accept_invitation`, {
      method: "POST",
      body: data,
      credentials: "include",
    });
    showSuccess("Success", "Your Group has been created successfully");
  } catch (e) {
    console.error(e);
    showError(
      "Error",
      "Can't accept invitation, ask for another invitation link"
    );
  } finally {
    removeInvitationCookie(inviteCode);
  }
}

const inviteCode = checkInvitation();
if (inviteCode && user.value?.nickname) {
  showConfirmation(
    "Accept invitation?",
    "You were invited to a new Group do you wanna participate to the Group?",
    () => acceptInvitation()
  );
  console.log("user logged in");
} else {
  console.log("user not logged in");
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
      label: "Groups",
      icon: "i-heroicons-user-group",
      to: "/group",
    },
    {
      label: "Logout",
      icon: "i-heroicons-arrow-left-on-rectangle",
      click: () => {
        showConfirmation(
          "Logging out?",
          "Are you sure you want to log out from sloth?",
          () => logOut()
        );
      },
    },
  ];
});
</script>
<template>
  <div
    class="hidden lg:block border border-gray-200 dark:border-gray-700 border-t-0 border-b-0 relative pt-5 px-2"
  >
    <UVerticalNavigation :links="links" />
  </div>
</template>
