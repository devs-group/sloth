<template>
  <Button
  class="overflow-clip"
  type="button"
    text
    @click="toggle"
  >
  <i class="pi pi-bell text-white" />
  <Badge v-if="notifications && notifications.length > 0" :value="notifications.length > 9 ? '9+' : notifications.length" class="absolute top-0 right-0 bg-red-600	text-white"></Badge>
</Button>

  <OverlayPanel ref="notificationOverlay" class="max-h-80 overflow-auto">
    <div class="flex flex-col w-25rem">
      <div
      v-if="notifications.length > 0"
        v-for="(notification, idx) in notifications"
        :key="idx"
        class="flex justify-between items-center py-4 gap-2"
      >
        <div class="flex flex-row gap-2">
          <Avatar :icon="notification.notification_type === NotificationType.INVITATION ? 'pi pi-envelope' : 'pi pi-info-circle'" class="mr-2" shape="circle" />
          <div>
            <p class="break-all">{{ notification.subject}}</p>
            <p class="text-xs text-prime-secondary-text">
              {{ notification.content }}
            </p>
          </div>
        </div>
        <div v-if="notification.notification_type === NotificationType.INVITATION">
        <Button
          icon="pi pi-check"
          text
          rounded
          aria-label="Filter"
          size="small"
          @click="acceptInvitation()"
        />
        <Button
          icon="pi pi-times"
          severity="danger"
          text
          rounded
          aria-label="Cancel"
          size="small"
          @click="declineInvitation()"
        />
        </div>
      </div>
      <div v-else>
        Nothing in your inbox.
      </div>
    </div>
  </OverlayPanel>
</template>

<script lang="ts" setup>
import { NotificationType } from '~/config/enums';

const notificationOverlay = ref();
const { notifications, getNotifications } = useNotification();

onMounted(async () => {
  await getNotifications();
});

const toggle = (event: any) => {
  notificationOverlay.value.toggle(event);
};

const acceptInvitation = () => {
  // TODO: implement logic
  console.log('accept invitation call')
};

const declineInvitation = () => {
  // TODO: implement logic
  console.log('decline invitation call')
};
</script>
