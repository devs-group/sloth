export default defineNuxtRouteMiddleware((to, from) => {
    const router = useRouter()
    const config = useRuntimeConfig()
    
    if (to.path.startsWith("/auth")) {
        return
    }

    $fetch(`${config.public.backendHost}/v1/auth/user`, {credentials: 'include', lazy: true, server: false})
    .then((d) => {
        if (!d.user) {
            return goToAuthPage(router)
        }
        useState("user", () => d.user)
    })
    .catch(() => {
        return goToAuthPage(router)
    })
});

function goToAuthPage(router) {
    console.log("User is not logged in... Redirecting to auth page")
    setTimeout(() => {
        router.push('/auth');
    }, 100)
}