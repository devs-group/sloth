import { Routes } from '~/config/routes'
import { Constants } from '~/config/const'

// This plugin extends the $fetch used in the app by checking responses for 401 status
// except for the verify-session call. It then makes a final call to verify-session.
// If that call fails it means the users session has ended while he was active,
// and we redirect to auth with a hint
export default defineNuxtPlugin(async (_nuxtApp) => {
  const { isAuthenticated, verifySession } = useAuth()
  const toast = useToast()

  globalThis.$fetch = $fetch.create({
    retry: false,
    onResponseError({ response }) {
      if (response.status == 401 && !response.url.includes('/auth/verify-session')) {
        verifySession().catch(async () => {
          if (isAuthenticated.value) {
            // TODO: Implement session refresh?
            isAuthenticated.value = false
            await navigateTo({ name: Routes.AUTH }, { replace: true })
            toast.add({
              severity: 'error',
              summary: 'Error',
              detail: 'Sorry, your session has ended. Please login again',
              life: Constants.ToasterDefaultLifeTime,
            })
          }
        })
      }
    },
  })
})
