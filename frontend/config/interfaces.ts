import type DockerCredentialsForm from "~/components/docker-credentials-form.vue"
import type OrganisationInvitationsForm from "~/components/organisation-invitations-form.vue"
import type OrganisationMembersForm from "~/components/organisation-members-form.vue"
import type ServicesForm from "~/components/services-form.vue"

export interface User {
    "avatar_url": string
    "email": string
    "first_name": string
    "id": number
    "last_name": string
    "location": string
    "name": string
    "nickname": string
}

export interface OAuthUserResponse {
    user: User
}

export interface ICreateProjectResponse {
    id: string
}

export interface ICreateOrganisationResponse{
    id: string
}

export interface IAddProjectToOrganisationResponse{
    id: string
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

export interface INavigationItems extends IBaseNavigationItems  {
    divider?: false
    label: string
}

export interface IDialogInjectRef<T> {
    value: {
        data: T
        close: (data?: any) => any
    }
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
    component?: (typeof ServicesForm | 
        typeof DockerCredentialsForm | 
        typeof OrganisationInvitationsForm | 
        typeof OrganisationMembersForm );
    to?: string
    props: any
    disabled?: boolean
}

export interface IServiceState {
    state: string;
    status: string;
}

export interface CreateOrganisationRequest {
    organisation_name: string
}

export type NavigationItems = INavigationItems | IDividerNavigationItems