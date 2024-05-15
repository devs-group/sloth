<template>
  <div></div>
</template>

<script lang="ts" setup>
import {Routes} from "~/config/routes";
// Simply redirect to the Projects
navigateTo({name: Routes.PROJECTS}, {replace: true})
</script>
