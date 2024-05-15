import {Routes} from "~/config/routes";

export default defineNuxtRouteMiddleware(async (to) => {
    const {isAuthenticated} = useAuth()

    const isOnPublicRoute = to.matched.some(p => p.path.startsWith(`/${Routes.AUTH}`))
    if (!isOnPublicRoute && !isAuthenticated.value) {
        return navigateTo({name: Routes.AUTH}, {replace: true})
    } else if (isOnPublicRoute && isAuthenticated.value) {
        return navigateTo({name: Routes.PROJECTS}, {replace: true})
    }
});