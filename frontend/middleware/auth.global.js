export default defineNuxtRouteMiddleware((to, from) => {
    const router = useRouter()
    const config = useRuntimeConfig()
    
    if (to.path.startsWith("/auth")) {
        return
    }

    useState("global.isLoading", () => true)
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
    .finally(() => {
        const isLoading = useState("global.isLoading")
        isLoading.value = false
    })
});

function goToAuthPage(router) {
    console.log("User is not logged in... Redirecting to auth page")
    setTimeout(() => {
        router.push('/auth');
    }, 100)
}