import { onMounted } from 'vue';
import { Constants } from '~/config/const';
import type { CreateOrganisationRequest } from '~/config/interfaces';
import { Routes } from '~/config/routes';
import { type Organisation, type OrganisationProject } from '~/schema/schema';

export function useOrganisation(organisationID: number ) {
    const config = useRuntimeConfig();
    const toast = useToast();
    const organisation = shallowRef<Organisation | null>(null);
    const organisationProjects = shallowRef<OrganisationProject[] | null>(null);

    async function saveOrganisation(orgName: string) {
      if (!orgName.trim().length) {
        return
      }

      try {
        await $fetch(`${config.public.backendHost}/v1/organisation`, {
          method: "POST",
          body: {
            organisation_name: orgName,
          } as CreateOrganisationRequest,
          credentials: "include",
        });
        toast.add({
          severity: "success",
          summary: "Success",
          detail: "Your Organisation has been created successfully",
          life: Constants.ToasterDefaultLifeTime,
        });
      } catch (e) {
        console.error(e);
        toast.add({
          severity: "error",
          summary: "Error",
          detail: "Something went wrong",
          life: Constants.ToasterDefaultLifeTime,
        });
        throw e;
      }
    }

    async function removeProjectFromOrganisation(upn: string) {
        try {
          organisation.value = await $fetch(
              `${config.public.backendHost}/v1/organisation/project`,
              {
                method: "DELETE",
                credentials: "include",
                body: {
                  id: organisation.value?.id,
                  upn: upn,
                },
              }
          );
          toast.add({
            severity: "success",
            summary: "Success",
            detail: "Project removed from organisation",
            life: Constants.ToasterDefaultLifeTime,
          });
        } catch (e) {
          console.error("unable to invite", e);
          toast.add({
            severity: "error",
            summary: "Error",
            detail: "Unable to remove Project",
            life: Constants.ToasterDefaultLifeTime,
          });
        } finally {
          fetchOrganisationProjects(organisationID);
        }
    }

    async function fetchOrganisationProjects(organisationID: number) {
        try {
          return await $fetch<OrganisationProject[]>(
            `${config.public.backendHost}/v1/organisation/${organisationID}/projects`,
            { credentials: "include" }
          );
        } catch (e) {
          console.error("unable to fetch Organisation", e);
        }
    }
    
    async function addProjectToOrganisation(upn: string) {
        console.log(upn)
        try {
          organisation.value = await $fetch(
              `${config.public.backendHost}/v1/organisation/project`,
              {
                method: "PUT",
                credentials: "include",
                body: {
                  organisation_id: organisationID,
                  upn: upn,
                },
              }
          );
          toast.add({
            severity: "success",
            summary: "Success",
            detail: "Project added to organisation",
            life: Constants.ToasterDefaultLifeTime,
          });
          const newID = parseInt(organisation.value?.id ?? "0",10)
          fetchOrganisationProjects(newID);
        } catch (e) {
          console.error("unable to invite", e);
          toast.add({
            severity: "error",
            summary: "Error",
            detail: "Unable to add Project",
            life: Constants.ToasterDefaultLifeTime,
          });
        } 
    }

    // Fetch organisation details
    async function fetchOrganisation() {
        try {
            organisation.value = await $fetch<Organisation>(
                `${config.public.backendHost}/v1/organisation/${organisationID}`,
                { credentials: "include" }
            );
        } catch (e) {
            console.error("unable to fetch organisation", e);
            toast.add({
                severity: "error",
                summary: "Fetch Failed",
                detail: "Unable to fetch organisation details"
            });
        }
    }

    // Delete a member from the organisation
    async function deleteMember(memberID: string) {
        try {
            await $fetch(
                `${config.public.backendHost}/v1/organisation/member/${organisationID}/${memberID}`,
                {
                    method: "DELETE",
                    credentials: "include",
                }
            );
            fetchOrganisation(); // Refresh data
            toast.add({
                severity: "success",
                summary: "Deleted",
                detail: "Member successfully removed"
            });
        } catch (e) {
            console.error("unable to delete member", e);
            toast.add({
                severity: "error",
                summary: "Deletion Failed",
                detail: "Unable to delete member"
            });
        }
    }

    // Invite a new member to the organisation
    async function inviteMember(email: String) {
        try {
            await $fetch(
                `${config.public.backendHost}/v1/organisation/member`,
                {
                    method: "PUT",
                    credentials: "include",
                    body: {
                        id: organisationID,
                        email: email,
                    },
                }
            );
            toast.add({
                severity: "success",
                summary: "Invitation Sent",
                detail: "Invitation has been sent successfully"
            });
        } catch (e) {
            console.error("unable to invite", e);
            toast.add({
                severity: "error",
                summary: "Invitation Failed",
                detail: "Unable to send invitation"
            });
        } finally {
            fetchOrganisation();
        }
    }

    return {
        saveOrganisation,
        fetchOrganisation,
        fetchOrganisationProjects,
        removeProjectFromOrganisation,
        addProjectToOrganisation,
        deleteMember,
        inviteMember,
        organisation,
        organisationProjects,
    };
}
