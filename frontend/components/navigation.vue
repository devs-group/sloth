<script setup>
const links = [{
  label: 'Profile',
  avatar: {
    src: 'https://avatars.githubusercontent.com/u/739984?v=4'
  },
  badge: 100
}, {
  label: 'Projects',
  icon: 'i-heroicons-home',
  to: '/'
}]
</script>
<template>
    <div class="border border-gray-200 dark:border-gray-700 border-t-0 border-b-0 h-full relative pt-5 px-2">
        <UVerticalNavigation :links="links" />
    </div>
  
</template>