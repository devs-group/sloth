import type { ToastServiceMethods } from "primevue/toastservice";
import type { User } from "~/config/interfaces";

export function useOrganisationInviation(toaster: ToastServiceMethods) {
  const toast = toaster;
  const user = useState<User>("user");
  const config = useRuntimeConfig();

  function checkInvitation(): string | null {
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
      await $fetch(
        `${config.public.backendHost}/v1/organisation/accept_invitation`,
        {
          method: "POST",
          body: data,
          credentials: "include",
        }
      );
      toast.add({
        severity: "success",
        summary: "Invitation Accepted",
        detail: "Invitation has been accepted",
      });
    } catch (e) {
      console.error("unable to accept invitation", e);
      toast.add({
        severity: "error",
        summary: "Invitation Acception Failed",
        detail: "Can't accept invitation, ask for another invitation link",
      });
    } finally {
      removeInvitationCookie(inviteCode);
    }
  }

  // Decline invitation of a new member to the organisation
  // TODO: Endpoint
  async function withdrawInvitation(email: string, organisationID: number) {
    const data = {
      email: email,
      organisation_id: organisationID
    };

    try {
      await $fetch(
        `${config.public.backendHost}/v1/organisation/withdraw_invitation`,
        {
          method: "DELETE",
          credentials: "include",
          body: data,
        }
      );
      toast.add({
        severity: "success",
        summary: "Invitation Withdrawn",
        detail: "Invitation has been withdrawn",
      });
    } catch (e) {
      console.error("unable to withdraw invite", e);
      toast.add({
        severity: "error",
        summary: "Withdraw Invitation Failed",
        detail: "Unable to withdraw invitation",
      });
    }
  }

  return {
    checkInvitation,
    acceptInvitation,
    withdrawInvitation,
  };
}
