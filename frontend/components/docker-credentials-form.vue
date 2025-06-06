<script lang="ts" setup>
import { z } from 'zod'
import type { DockerCredentialSchema } from '~/schema/schema'
import { dockerCredentialSchema } from '~/schema/schema'

const props = defineProps<{
  credentials: DockerCredentialSchema[]
  submitted: boolean
}>()

defineEmits<{
  (event: 'addCredential'): void
  (event: 'removeCredential', index: number): void
}>()

const { validate, getError } = useValidation(z.array(dockerCredentialSchema), props.credentials)

const validateInputFields = () => {
  props.credentials.forEach((_credential, index) => {
    validate(index, 'username')
    validate(index, 'password')
    validate(index, 'registry')
  })
}

onMounted(() => {
  if (props.submitted) {
    validateInputFields()
  }
})

watch(() => props.submitted, validateInputFields)
</script>

<template>
  <div class="flex flex-col flex-1">
    <div class="flex items-center gap-4 py-6">
      <p class="text-prime-secondary-text">
        Docker registry credentials
      </p>
      <IconButton
        icon="heroicons:plus"
        outlined
        @click="$emit('addCredential')"
      />
    </div>
    <div class="flex gap-12 overflow-auto flex-1">
      <div
        v-for="(credential, cIdx) of props.credentials"
        :key="credential.id"
        class="flex flex-col gap-6 max-w-[28em]"
      >
        <div class="flex flex-col gap-1">
          <Label label="Username" />
          <InputText
            v-model="credential.username"
            @blur="validate(cIdx, 'username')"
          />
          <small class="text-prime-danger">{{ getError(cIdx, 'username')?.message }}</small>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Password" />
          <InputText
            v-model="credential.password"
            @blur="validate(cIdx, 'password')"
          />
          <small class="text-prime-danger">{{ getError(cIdx, 'password')?.message }}</small>
        </div>
        <div class="flex flex-col gap-1">
          <Label label="Registry" />
          <InputText
            v-model="credential.registry"
            @blur="validate(cIdx, 'registry')"
          />
          <small class="text-prime-danger">{{ getError(cIdx, 'registry')?.message }}</small>
        </div>
        <div>
          <Button
            severity="danger"
            outlined
            label="Remove credential"
            class="w-full flex justify-center"
            @click="$emit('removeCredential', cIdx)"
          />
        </div>
      </div>
    </div>
  </div>
</template>
