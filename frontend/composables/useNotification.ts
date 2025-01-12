import type { INotification } from '~/config/interfaces'

export function useNotification() {
  const config = useRuntimeConfig()
  const toast = useToast()
  const notifications = ref<INotification[]>([])

  async function putNotification(subject: string, content: string, recipient: string) {
    try {
      await $fetch(
        `${config.public.backendHost}/v1/notifications`,
        {
          method: 'PUT',
          credentials: 'include',
          body: {
            subject: subject,
            content: content,
            recipient: recipient,
          },
        },
      )
      toast.add({
        severity: 'success',
        summary: 'Notification Stored',
        detail: 'Notification has been stored',
      })
    }
    catch (err) {
      console.error(err)
      toast.add({
        severity: 'error',
        summary: 'Notification Store Failed',
        detail: 'Failed to store notification, please try again',
      })
    }
  }

  async function getNotifications() {
    try {
      const data: INotification[] = await $fetch(
        `${config.public.backendHost}/v1/notifications`,
        {
          method: 'GET',
          credentials: 'include',
        },
      )
      notifications.value = data ? data : []
    }
    catch (e) {
      console.error(e)
    }
  }

  return {
    notifications,
    putNotification,
    getNotifications,
  }
}
