import type { ToastServiceMethods } from 'primevue/toastservice'
import { Constants } from '~/config/const'
import type { ICreateOrganisationRequest } from '~/config/interfaces'
import type { Invitation, Organisation, OrganisationProject } from '~/schema/schema'
import type { OrganisationPUTFormData, OrganisationResponse } from '~/models/organisation'

export function useOrganisation(
  toaster: ToastServiceMethods,
) {
  const config = useRuntimeConfig()
  const toast = toaster
  const currentOrganisation = shallowRef<Organisation | null>(null)
  const invitations = shallowRef<Invitation[]>([])
  const organisationProjects = shallowRef<OrganisationProject[] | null>(null)

  async function saveOrganisation(orgName: string) {
    if (!orgName.trim().length) {
      return
    }

    try {
      await $fetch(`${config.public.backendHost}/v1/organisation`, {
        method: 'POST',
        body: {
          organisationName: orgName,
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

  async function removeProjectFromOrganisation(upn: string, name: string, organisationID: number) {
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

  async function addProjectToOrganisation(upn: string, organisationID: number) {
    try {
      currentOrganisation.value = await $fetch(
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
      const newID = currentOrganisation.value!.id
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
  const fetchOrganisation = async (organisationID: number) => {
    try {
      const data = await $fetch<Organisation>(`${config.public.backendHost}/v1/organisation/${organisationID}`, {
        method: 'GET',
      })
      currentOrganisation.value = data
    }
    catch {
      return Promise.reject('Fehler beim Laden der Organisation')
    }
  }

  const updateOrganisation = async (organisationID: number, payload: OrganisationPUTFormData) => {
    try {
      await $fetch<OrganisationResponse>(`${config.public.backendHost}/v1/organisation/${organisationID}`, {
        method: 'PUT',
        body: {
          ...payload,
        },
      })
    }
    catch {
      return Promise.reject('Error updating Organisation')
    }
  }

  // Delete a member from the organisation
  async function deleteMember(memberID: number) {
    try {
      await $fetch<OrganisationResponse>(`/v1/organisation/member/${currentOrganisation.value.id}/${memberID}`, {
        method: 'DELETE',
        baseURL: config.public.backendHost,
      })
    }
    catch {
      return Promise.reject('Error removing Member from Organisation')
    }
  }

  // Invite a new member to the organisation
  async function createInvitation(email: string) {
    try {
      await $fetch(`/v1/organisation/member/invitation`, {
        method: 'POST',
        baseURL: config.public.backendHost,
        credentials: 'include',
        body: {
          organisationID: currentOrganisation.value?.id,
          email: email,
        },
      })
    }
    catch {
      return Promise.reject('Unable to send invitation')
    }
  }

  async function deleteInvitation(invitationID: number) {
    try {
      await $fetch(`/v1/organisation/member/invitation/${invitationID}`, {
        method: 'DELETE',
        baseURL: config.public.backendHost,
        // credentials: 'include',
      })
    }
    catch {
      return Promise.reject('Unable to delete invitation')
    }
  }

  const canEditOrganisation = () => currentOrganisation.value
    ? ['owner', 'admin'].includes(currentOrganisation.value.currentRole)
    : false

  const isOwnerOfOrganisation = () => currentOrganisation.value && currentOrganisation.value.currentRole == 'owner'

  return {
    saveOrganisation,
    updateOrganisation,
    fetchOrganisation,
    fetchOrganisationProjects,
    removeProjectFromOrganisation,
    addProjectToOrganisation,
    deleteMember,
    invitations,
    createInvitation,
    deleteInvitation,
    currentOrganisation,
    organisationProjects,
    canEditOrganisation,
    isOwnerOfOrganisation,
  }
}
