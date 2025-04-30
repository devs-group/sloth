import type { ToastServiceMethods } from 'primevue/toastservice'
import { Constants } from '~/config/const'
import type { ICreateOrganisationRequest } from '~/config/interfaces'
import type { Organisation, OrganisationProject } from '~/schema/schema'

export function useOrganisation(
  organisationID: number | string,
  toaster: ToastServiceMethods,
) {
  const config = useRuntimeConfig()
  const toast = toaster
  const organisation = useState<Organisation>('organisation', () => null) // shallowRef<Organisation | null>(null)
  const organisationProjects = shallowRef<OrganisationProject[] | null>(null)

  async function saveOrganisation(orgName: string) {
    if (!orgName.trim().length) {
      return
    }

    try {
      await $fetch(`${config.public.backendHost}/v1/organisation`, {
        method: 'POST',
        body: {
          organisation_name: orgName,
        } as ICreateOrganisationRequest,
        credentials: 'include',
      })
      toast.add({
        severity: 'success',
        summary: 'Success',
        detail: 'Your Organisation has been created successfully',
        life: Constants.ToasterDefaultLifeTime,
      })
    }
    catch (e) {
      console.error(e)
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: 'Something went wrong',
        life: Constants.ToasterDefaultLifeTime,
      })
      throw e
    }
  }

  async function removeProjectFromOrganisation(upn: string, name: string) {
    try {
      await $fetch(`${config.public.backendHost}/v1/organisation/project`, {
        method: 'DELETE',
        credentials: 'include',
        body: {
          organisation_id: organisationID,
          upn: upn,
        },
      })
      toast.add({
        severity: 'success',
        summary: 'Success',
        detail: `Project "${name}" has been deleted successfully`,
        life: Constants.ToasterDefaultLifeTime,
      })
      fetchOrganisationProjects(organisationID)
    }
    catch (e) {
      console.error('unable to invite', e)
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: `Failed to delete project "${name}"`,
        life: Constants.ToasterDefaultLifeTime,
      })
    }
  }

  async function fetchOrganisationProjects(organisationID: number | string) {
    try {
      organisationProjects.value = await $fetch<OrganisationProject[]>(
        `${config.public.backendHost}/v1/organisation/${organisationID}/projects`,
        { credentials: 'include' },
      )
      return organisationProjects
    }
    catch (e) {
      console.error('unable to fetch Organisation', e)
    }
  }

  async function addProjectToOrganisation(upn: string, organisationID: string) {
    try {
      organisation.value = await $fetch(
        `${config.public.backendHost}/v1/organisation/project`,
        {
          method: 'PUT',
          credentials: 'include',
          body: {
            organisation_id: organisationID,
            upn: upn,
          },
        },
      )
      toast.add({
        severity: 'success',
        summary: 'Success',
        detail: 'Project added to organisation',
        life: Constants.ToasterDefaultLifeTime,
      })
      const newID = organisation.value!.id
      fetchOrganisationProjects(newID)
    }
    catch (e) {
      console.error('unable to invite', e)
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: 'Unable to add Project',
        life: Constants.ToasterDefaultLifeTime,
      })
    }
  }

  // Fetch organisation details
  async function fetchOrganisation() {
    try {
      organisation.value = await $fetch<Organisation>(
        `${config.public.backendHost}/v1/organisation/${organisationID}`,
        { credentials: 'include' },
      )
    }
    catch (e) {
      console.error('unable to fetch organisation', e)
      toast.add({
        severity: 'error',
        summary: 'Fetch Failed',
        detail: 'Unable to fetch organisation details',
      })
    }
  }

  // Delete a member from the organisation
  async function deleteMember(memberID: number) {
    try {
      await $fetch(
        `${config.public.backendHost}/v1/organisation/member/${organisationID}/${memberID}`,
        {
          method: 'DELETE',
          credentials: 'include',
        },
      )
      fetchOrganisation() // Refresh data
      toast.add({
        severity: 'success',
        summary: 'Deleted',
        detail: 'Member successfully removed',
      })
    }
    catch (e) {
      console.info('unable to delete member', e)
      toast.add({
        severity: 'error',
        summary: 'Deletion Failed',
        detail: 'Unable to delete member',
      })
    }
  }

  // Invite a new member to the organisation
  async function inviteMember(email: string) {
    try {
      await $fetch(`${config.public.backendHost}/v1/organisation/member`, {
        method: 'PUT',
        credentials: 'include',
        body: {
          organisation_id: organisationID,
          email: email,
        },
      })
      toast.add({
        severity: 'success',
        summary: 'Invitation Sent',
        detail: 'Invitation has been sent successfully',
      })
    }
    catch (e) {
      console.error('unable to invite', e)
      toast.add({
        severity: 'error',
        summary: 'Invitation Failed',
        detail: 'Unable to send invitation',
      })
    }
    finally {
      fetchOrganisation()
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
  }
}
