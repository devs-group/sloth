// One time initial call to check if the user is verified or not. This is run before everything else.
export default defineNuxtPlugin(async (_nuxtApp) => {
  const { isAuthenticated, verifySession } = useAuth()

  try {
    await verifySession()
    isAuthenticated.value = true
  }
  catch {
    isAuthenticated.value = false
  }
})
