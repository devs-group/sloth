<script lang="ts" setup>
import type {DockerCredentialSchema} from "~/schema/schema";

const props = defineProps({
  credentials: {
    required: true,
    type: Object as PropType<DockerCredentialSchema[]>
  }
})

defineEmits<{
  (event: 'addCredential'): void,
  (event: 'removeCredential', index: number): void
}>()

const config = useRuntimeConfig()
</script>

<template>
  <div class="flex flex-col flex-1">
    <div class="flex items-center gap-4 py-6">
        <p class="text-prime-secondary-text">Docker registry credentials</p>
        <IconButton icon="heroicons:plus" @click="$emit('addCredential')" outlined/>
    </div>
    <div class="flex gap-12 overflow-auto flex-1">
      <div v-for="credential,cIdx in props.credentials" class="flex flex-col gap-6 w-[28em]">
        <div class="flex flex-col gap-1">
          <Label label="Username"/>
          <InputText v-model="credential.username"/>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Password"/>
          <InputText v-model="credential.password"/>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Registry"/>
          <InputText v-model="credential.registry"/>
        </div>
        <div>
          <Button 
            severity="danger"
            outlined
            @click="$emit('removeCredential', cIdx)" 
            label="Remove credential"
            class="w-full flex justify-center"
          />
        </div>
      </div>
    </div>
  </div>
</template>