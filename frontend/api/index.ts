import type { INotification, IServiceState } from '~/config/interfaces'
import type { Invitation, Organisation, OrganisationProject, Project } from '~/schema/schema'

const config = useRuntimeConfig()

export const APIService = {
  /**
   * Fetches all organisations.
   * @returns A promise that resolves with the list of organisations.
   */
  GET_organisations: async () => {
    return useFetch<Organisation[]>(
      `${config.public.backendHost}/v1/organisations`,
      {
        credentials: 'include',
        method: 'GET',
      },
    )
  },

  /**
   * Creates a new organisation with the specified name.
   * @param name - The name of the organisation to be created.
   * @returns A promise that resolves with the created organisation.
   */
  POST_organisation: async (name: string) => {
    return useFetch<Organisation>(
      `${config.public.backendHost}/v1/organisation`,
      {
        credentials: 'include',
        method: 'POST',
        body: {
          organisationName: name,
        },
      },
    )
  },

  /**
   * Fetches invitations by organisation ID.
   * @param organisationID - The ID of the organisation to fetch invitations for.
   * @returns A promise that resolves with the list of invitations.
   */
  GET_invitationsByOrganisationID: async (organisationID: number) => {
    return useFetch<Invitation[]>(
      `${config.public.backendHost}/v1/organisation/${organisationID}/invitations`,
      {
        credentials: 'include',
        method: 'GET',
      },
    )
  },

  /**
   * Fetches projects by organisation ID.
   * @param organisationID - The ID of the organisation to fetch projects for.
   * @returns A promise that resolves with the list of projects.
   */
  GET_projectsByOrganisationID: async (organisationID: number) => {
    return useFetch<OrganisationProject[]>(
      `${config.public.backendHost}/v1/organisation/${organisationID}/projects`,
      {
        credentials: 'include',
        method: 'GET',
      },
    )
  },

  /**
   * Deletes a project from an organisation.
   * @param organisationID - The ID of the organisation.
   * @param upn - The unique project name.
   * @returns A promise indicating the completion of the deletion.
   */
  DELETE_projectFromOrganisation: async (
    organisationID: number | string,
    upn: string,
  ) => {
    return useFetch(`${config.public.backendHost}/v1/organisation/project`, {
      credentials: 'include',
      method: 'DELETE',
      body: {
        organisation_id: organisationID,
        upn: upn,
      },
    })
  },

  /**
   * Adds a project to an organisation.
   * @param organisationID - The ID of the organisation.
   * @param upn - The unique project name.
   * @returns A promise indicating the completion of the addition.
   */
  PUT_projectToOrganisation: async (organisationID: string, upn: string) => {
    return useFetch(`${config.public.backendHost}/v1/organisation/project`, {
      credentials: 'include',
      method: 'PUT',
      body: {
        organisation_id: organisationID,
        upn: upn,
      },
    })
  },

  /**
   * Fetches details of a specific organisation.
   * @param organisationID - The ID of the organisation.
   * @returns A promise that resolves with the organisation details.
   */
  GET_organisationDetails: async (organisationID: number) => {
    return useFetch<Organisation>(
      `${config.public.backendHost}/v1/organisation/${organisationID}`,
      {
        credentials: 'include',
        method: 'GET',
      },
    )
  },

  /**
   * Deletes a member from an organisation.
   * @param organisationID - The ID of the organisation.
   * @param memberID - The ID of the member to delete.
   * @returns A promise indicating the completion of the deletion.
   */
  DELETE_memberFromOrganisation: async (
    organisationID: number,
    memberID: number,
  ) => {
    return useFetch(
      `${config.public.backendHost}/v1/organisation/member/${organisationID}/${memberID}`,
      {
        credentials: 'include',
        method: 'DELETE',
      },
    )
  },

  /**
   * Invites a new member to the organisation.
   * @param organisationID - The ID of the organisation.
   * @param email - The email of the member to invite.
   * @returns A promise indicating the completion of the invitation.
   */
  PUT_inviteMemberToOrganisation: async (
    organisationID: number | string,
    email: string,
  ) => {
    return useFetch(`${config.public.backendHost}/v1/organisation/member`, {
      credentials: 'include',
      method: 'PUT',
      body: {
        organisation_id: organisationID,
        email: email,
      },
    })
  },

  /**
   * Accepts an invitation to join an organisation.
   * @param userID - The ID of the user accepting the invitation.
   * @param inviteCode - The invitation token.
   * @returns A promise indicating the completion of the acceptance.
   */
  POST_acceptInvitation: async (userID: number, inviteCode: string) => {
    return useFetch(
      `${config.public.backendHost}/v1/organisation/accept_invitation`,
      {
        credentials: 'include',
        method: 'POST',
        body: {
          user_id: userID,
          invitation_token: inviteCode,
        },
      },
    )
  },

  /**
   * Withdraws an invitation to join an organisation.
   * @param organisationID - The ID of the organisation.
   * @param email - The email of the invitee to withdraw the invitation from.
   * @returns A promise indicating the completion of the withdrawal.
   */
  DELETE_withdrawInvitation: async (organisationID: number, email: string) => {
    return useFetch(
      `${config.public.backendHost}/v1/organisation/withdraw_invitation`,
      {
        credentials: 'include',
        method: 'DELETE',
        body: {
          email: email,
          organisation_id: organisationID,
        },
      },
    )
  },

  /**
   * Deletes an organisation by its ID.
   * @param organisationID - The ID of the organisation to delete.
   * @returns A promise indicating the completion of the deletion.
   */
  DELETE_organisation: async (organisationID: number) => {
    return useFetch(
      `${config.public.backendHost}/v1/organisation/${organisationID}`,
      {
        credentials: 'include',
        method: 'DELETE',
      },
    )
  },

  /**
   * Fetches all notifications.
   * @returns A promise that resolves with the list of notifications.
   */
  GET_notifications: async () => {
    return useFetch<INotification[]>(
      `${config.public.backendHost}/v1/notifications`,
      {
        credentials: 'include',
        method: 'GET',
      },
    )
  },

  /**
   * Stores a new notification.
   * @param subject - The subject of the notification.
   * @param content - The content of the notification.
   * @param recipient - The recipient of the notification.
   * @returns A promise indicating the completion of the storage.
   */
  PUT_notification: async (
    subject: string,
    content: string,
    recipient: string,
  ) => {
    return useFetch(`${config.public.backendHost}/v1/notifications`, {
      credentials: 'include',
      method: 'PUT',
      body: {
        subject: subject,
        content: content,
        recipient: recipient,
      },
    })
  },

  /**
   * Creates a new project.
   * @param name - The name of the project.
   * @returns A promise that resolves with the response of the created project.
   */
  POST_project: async (name: string) => {
    return useFetch<Project>(`${config.public.backendHost}/v1/project`, {
      method: 'POST',
      body: { name },
      credentials: 'include',
    })
  },

  /**
   * Fetches a project by its ID.
   * @param projectID - The ID of the project to fetch.
   * @returns A promise that resolves with the project details.
   */
  GET_projectByID: async (projectID: number) => {
    return useFetch<Project>(
      `${config.public.backendHost}/v1/project/${projectID}`,
      {
        credentials: 'include',
        method: 'GET',
      },
    )
  },

  /**
   * Updates an existing project.
   * @param project - The project data to update.
   * @returns A promise indicating the completion of the update.
   */
  PUT_updateProject: async (project: Project) => {
    return useFetch<Project>(
      `${config.public.backendHost}/v1/project/${project.id}`,
      {
        credentials: 'include',
        method: 'PUT',
        body: project,
      },
    )
  },

  /**
   * Fetches all projects.
   * @returns A promise that resolves with the list of projects.
   */
  GET_allProjects: async () => {
    return useFetch<Project[]>(`${config.public.backendHost}/v1/projects`, {
      credentials: 'include',
      method: 'GET',
    })
  },

  /**
   * Fetches the states of services for a specific project.
   * @param projectID - The ID of the project to fetch service states for.
   * @returns A promise that resolves with the service states.
   */
  GET_serviceStates: async (projectID: number) => {
    return useFetch<Record<string, IServiceState>>(
      `${config.public.backendHost}/v1/project/state/${projectID}`,
      {
        credentials: 'include',
        method: 'GET',
      },
    )
  },
}
