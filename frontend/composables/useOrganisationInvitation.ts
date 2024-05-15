import type { UserSchema } from '~/schema/schema';

export function useOrganisationInviation() {
  const toast = useToast();
  const user = useState<UserSchema>("user");
  const config = useRuntimeConfig();

  function checkInvitation(): string | null  {
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

  return { checkInvitation, acceptInvitation };
}
