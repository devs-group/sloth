import { onMounted } from 'vue';
import { Constants } from '~/config/const';
import { type Organisation, type OrganisationProject } from '~/schema/schema';

export function useOrganisation(organisationName: string | string[], config: any, toast: any ) {
    const organisation = shallowRef<Organisation | null>(null);
    const isAddMemberModalOpen = shallowRef(false);
    const memberID = shallowRef("");

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
          fetchOrganisationProjects();
        }
    }

    async function addProjectToOrganisation(upn: string, organisationName: string, modalOpenFlag: Ref) {
        try {
          organisation.value = await $fetch(
              `${config.public.backendHost}/v1/organisation/project`,
              {
                method: "PUT",
                credentials: "include",
                body: {
                  organisation_name: organisationName,
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
        } catch (e) {
          console.error("unable to invite", e);
          toast.add({
            severity: "error",
            summary: "Error",
            detail: "Unable to add Project",
            life: Constants.ToasterDefaultLifeTime,
          });
        } finally {
          modalOpenFlag.value = false;
          fetchOrganisationProjects();
        }
    }

    async function fetchOrganisationProjects() {
        try {
          await $fetch<OrganisationProject[]>(
              `${config.public.backendHost}/v1/organisation/${organisationName}/projects`,
              { credentials: "include" }
          );
        } catch (e) {
          console.error("unable to fetch Organisation", e);
        }
    }

    // Fetch organisation details
    async function fetchOrganisation() {
        try {
            organisation.value = await $fetch<Organisation>(
                `${config.public.backendHost}/v1/organisation/${organisationName}`,
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
                `${config.public.backendHost}/v1/organisation/member/${organisationName}/${memberID}`,
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
    async function inviteMember() {
        try {
            await $fetch(
                `${config.public.backendHost}/v1/organisation/member`,
                {
                    method: "PUT",
                    credentials: "include",
                    body: {
                        id: organisation.value?.id,
                        organisation_name: organisation.value?.organisation_name,
                        email: memberID.value,
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
            isAddMemberModalOpen.value = false;
            fetchOrganisation();
        }
    }

    onMounted(() => {
        fetchOrganisation();
    });

    return {
        organisation,
        isAddMemberModalOpen,
        memberID,
        fetchOrganisation,
        fetchOrganisationProjects,
        removeProjectFromOrganisation,
        addProjectToOrganisation,
        deleteMember,
        inviteMember,
    };
}
