import {Routes} from "~/config/routes";
import type { NavigationItems, OAuthUserResponse, User } from "~/config/interfaces";

const user = ref<User>()
const isAuthenticated = ref(false)

export const useAuth = () => {
    const getUser = async () => {
        const config = useRuntimeConfig()
        const data = await $fetch<OAuthUserResponse>(
            `${config.public.backendHost}/v1/auth/user`, {
                method: 'GET',
                credentials: 'include',
            })
        user.value = data.user
    }
    const verifySession = async () => {
        const config = useRuntimeConfig()
        return await $fetch<OAuthUserResponse>(
            `${config.public.backendHost}/v1/auth/verify-session`, {
                method: 'GET',
                credentials: 'include',
            })
    }
    const logout = () => {
        const config = useRuntimeConfig()
        return $fetch(`${config.public.backendHost}/v1/auth/logout/github`, {
            credentials: "include",
        })
    }

    const getMenuItems = (data: {onLogout: Function}) => {
        return [
            {
                label: "Projects",
                icon: "heroicons-home",
                to: Routes.PROJECTS,
            },
            {
                label: "Organisations",
                icon: "heroicons-user-group",
                to: Routes.ORGANISATIONS,
            },
            {
                divider: true,
            },
            {
                label: "Logout",
                icon: "heroicons-arrow-left-on-rectangle",
                click: data.onLogout,
            },
        ] as NavigationItems[]
    }

    return {
        logout, getUser, verifySession, getMenuItems,
        user, isAuthenticated,
    }
}