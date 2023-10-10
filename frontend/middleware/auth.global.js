export default defineNuxtRouteMiddleware((to, from) => {
    const router = useRouter()
    const config = useRuntimeConfig()
    $fetch(`${config.public.backendHost}/v1/auth/user`, {credentials: 'include', lazy: true, server: false})
    .then((d) => {
        if (!d.user) {
            return router.push('/auth');
        }
        useState("user", () => d.user)
    })
    .catch(() => {
        return router.push('/auth');
    })
});