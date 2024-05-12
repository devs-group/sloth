interface User {
    "avatar_url": string
    "email": string
    "first_name": string
    "id": number
    "last_name": string
    "location": string
    "name": string
    "nickname": string
}

interface OAuthUserResponse {
    user: User
}

interface ICreateProjectResponse {
    id: string
}

interface IBaseNavigationItems {
    icon?: string
    to?: string
    click?: () => void
}

interface IDividerNavigationItems extends IBaseNavigationItems {
    divider: true
    label?: string
}

interface INavigationItems extends IBaseNavigationItems  {
    divider?: false
    label: string
}

interface IDialogInjectRef<T> {
    value: {
        data: T
        close: (data?: any) => any
    }
}

interface ICreateProjectDialog {
    name: string
}

interface ICustomConfirmDialog {
    question: string
    confirmText: string
    rejectText: string
}

interface TabItem {
    label: string
    icon?: string
    command?: () => void
    component?: string
    to?: string
    disabled?: boolean
}

interface CreateOrganisationRequest {
    organisation_name: string
}

type NavigationItems = INavigationItems | IDividerNavigationItems