import type { AsyncData } from "#app";
import type { ToastServiceMethods } from "primevue/toastservice";
import { Constants } from "~/config/const";
import { type Invitation, type Organisation } from "~/schema/schema";

export function useOrganisations(toaster: ToastServiceMethods) {
    const config = useRuntimeConfig();
    const toast = toaster;

    const invitations = shallowRef<Invitation[] | null>(null);

    function loadOrganisations() {
        return useFetch<Organisation[]>(`${config.public.backendHost}/v1/organisations`, {
          server: false,
          lazy: true,
          credentials: "include",
        });
    }
      
    async function loadInvitations() {
        try {
            invitations.value =  await $fetch<Invitation[]>(
              `${config.public.backendHost}/v1/organisations/invitations`,
              { credentials: "include" }
            );
            return invitations
          } catch (e) {
            console.error("unable to fetch invitations", e);
          }
    }
      
    async function deleteOrganisation(organisation: Organisation): Promise<Organisation[] | null> {
        try {
            await $fetch(`${config.public.backendHost}/v1/organisation/${organisation.id}`, {
                method: "DELETE",
                credentials: "include",
            });
    
            const { data } = await loadOrganisations();
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
    }
}