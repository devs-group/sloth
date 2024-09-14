import type { ToastServiceMethods } from "primevue/toastservice";
import { APIService } from "~/api";
import { Constants } from "~/config/const";
import { type Invitation, type Organisation } from "~/schema/schema";

export function useOrganisations(toaster: ToastServiceMethods) {
  const config = useRuntimeConfig();
  const toast = toaster;

  const invitations = shallowRef<Invitation[] | null>(null);

  function loadOrganisations() {
    return APIService.GET_organisations();
  }

  async function loadInvitations(organisationId: number) {
    return APIService.GET_invitationsByOrganisationID(organisationId);
  }

  async function deleteOrganisation(
    organisation: Organisation
  ): Promise<Organisation[] | null> {
    try {
      await $fetch(
        `${config.public.backendHost}/v1/organisation/${organisation.id}`,
        {
          method: "DELETE",
          credentials: "include",
        }
      );

      const { data, error } = await loadOrganisations();
      if (error.value) {
        throw new Error("unable to load organisations");
      }

      toast.add({
        severity: "success",
        summary: "Success",
        detail: `Organisation "${organisation.organisation_name}" has been removed successfully`,
        life: Constants.ToasterDefaultLifeTime,
      });

      return data.value;
    } catch (e) {
      console.error("Failed to delete organisation:", e);
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Failed to delete organisation. Please try again.",
        life: Constants.ToasterDefaultLifeTime,
      });

      return null;
    }
  }

  return {
    invitations,
    deleteOrganisation,
    loadInvitations,
    loadOrganisations,
  };
}
