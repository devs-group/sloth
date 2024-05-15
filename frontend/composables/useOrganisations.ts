import { Constants } from "~/config/const";
import { type Invitation, type Organisation } from "~/schema/schema";

export function useOrganisations() {
    const config = useRuntimeConfig();
    const toast = useToast();

    function loadOrganisations() {
        return useFetch<Organisation[]>(`${config.public.backendHost}/v1/organisations`, {
          server: false,
          lazy: true,
          credentials: "include",
        });
    }
      
    function loadInvitations() {
        return useFetch<Invitation[]>(
            `${config.public.backendHost}/v1/organisations/invitations`,
            {
              server: false,
              lazy: true,
              credentials: "include",
            }
        );
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
        deleteOrganisation,
        loadInvitations,
        loadOrganisations,
    }
}