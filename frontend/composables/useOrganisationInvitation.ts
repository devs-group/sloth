import { useRouter } from 'vue-router';
import type { UserSchema } from '~/schema/schema';

export function useOrganisationInviation() {
  const router = useRouter();
  const toast = useToast();
  const confirm = useConfirm();
  const user = useState<UserSchema>("user");
  const config = useRuntimeConfig();

  function checkInvitation() {
    const cookies = document.cookie.split("; ");
    const inviteCookie = cookies.find((cookie) =>
      cookie.startsWith("inviteCode=")
    );

    if (inviteCookie) {
      return inviteCookie.split("=")[1];
    }
    return null;
  }

  function removeInvitationCookie(inviteCode: string) {
    // TODO
    console.log("Remove cookie logic here for", inviteCode);
  }

  async function acceptInvitation(inviteCode: string) {
    
    const data = {
      user_id: user.value?.id,
      invitation_token: inviteCode,
    };

    try {
      await $fetch(`${config.public.backendHost}/v1/organisation/accept_invitation`, {
        method: "POST",
        body: data,
        credentials: "include",
      });
      toast.add({
        severity: "success",
        summary: "Success",
        detail: "Successfully accepted invitation",
      });
    } catch (e) {
      console.error(e);
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Can't accept invitation, ask for another invitation link",
      });
    } finally {
      removeInvitationCookie(inviteCode);
    }
  }

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

  return { checkInvitation, acceptInvitation, logOut, confirm, user, toast };
}
