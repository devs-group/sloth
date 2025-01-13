import type { NotificationType } from './enums'
import type DockerCredentialsForm from '~/components/docker-credentials-form.vue'
import type OrganisationInvitationsForm from '~/components/organisation-invitations-form.vue'
import type OrganisationMembersForm from '~/components/organisation-members-form.vue'
import type OrganisationProjectList from '~/components/organisation-project-list.vue'
import type ServicesForm from '~/components/services-form.vue'
import type { OrganisationProject } from '~/schema/schema'

export interface User {
  avatar_url: string
  email: string
  first_name: string
  id: number
  last_name: string
  location: string
  name: string
  nickname: string
}

export interface OAuthUserResponse {
  user: User
}

export interface ICreateProjectResponse {
  id: string
}

export interface IAddProjectToOrganisation {
  upn: string
  organisation_id: number
}

export interface IPutMemberToOrganisation {
  email: string
  organisation_id: number
}

export interface IAddProjectToOrganisationResponse {
  id: string
  upn: string
}

export interface IPutMemberToOrganisationResponse {
  id: string
  upn: string
}

export interface IBaseNavigationItems {
  icon?: string
  to?: string
  click?: () => void
}

export interface IDividerNavigationItems extends IBaseNavigationItems {
  divider: true
  label?: string
}

export interface INavigationItems extends IBaseNavigationItems {
  divider?: false
  label: string
}

export interface IDialogInjectRef<T, V> {
  value: {
    data: T
    close: (data?: V) => V
  }
}

export interface IAddMemberToOrganisationDialog {
  organisation_id: number
}

export interface IAddProjectToOrganisationDialog {
  organisation_id: number
  organisationProjects: OrganisationProject[]
}

export interface IInviteToOrganisationDialog {
  organisation_id: number
}

export interface ICreateProjectDialog {
  name: string
}

export interface ICustomConfirmDialog {
  question: string
  confirmText: string
  rejectText: string
}
export interface TabItem {
  label: string
  icon?: string
  command?: () => void
  component?:
    | typeof ServicesForm
    | typeof DockerCredentialsForm
    | typeof OrganisationInvitationsForm
    | typeof OrganisationMembersForm
    | typeof OrganisationProjectList
  to?: string
  props?: Record<string, unknown>
  disabled?: boolean
}

export interface IServiceState {
  id: string
  state: string
  status: string
}

export interface ICreateOrganisationRequest {
  organisation_name: string
}

export interface IInviteToOrganisation {
  email: string
  organisation_id: number
}

export interface IInviteToOrganisationResponse {
  eMail: string
  id: string
}

export type NavigationItems = INavigationItems | IDividerNavigationItems

export interface INotification {
  id: string
  subject: string
  content: string
  sender: string
  recipient: string
  time_stamp: string
  notification_type: NotificationType
}

export type TLogoutFunction = () => void
