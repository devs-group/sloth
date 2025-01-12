import type { z, ZodIssue, ZodTypeAny } from 'zod'
import { ZodArray, ZodObject, ZodTuple } from 'zod'
import { groupBy } from 'lodash-es'
import { ref } from 'vue'

function useValidation<T extends ZodTypeAny>(schema: T, data: z.output<typeof schema>) {
  const isValid = ref(true)
  const errors = ref<Map<string, ZodIssue[]>>(new Map())

  const clearErrors = () => {
    errors.value.clear()
  }

  const validateByPath = (path: (string | number)[]): boolean => {
    let subSchema = schema
    let subData = data
    for (const property of path) {
      subSchema = getSubSchema(subSchema, property)
      subData = subData[property]
    }
    const result = subSchema.safeParse(subData)
    isValid.value = result.success
    if (!result.success) {
      errors.value.set(path.join(','), result.error.issues)
    }
    else {
      errors.value.delete(path.join(','))
    }

    return result.success
  }

  const validate = (...path: (string | number)[]): boolean => {
    if (path.length > 0) {
      return validateByPath(path)
    }

    const result = schema.safeParse(data)
    isValid.value = result.success
    if (!result.success) {
      errors.value = new Map(Object.entries(groupBy(result.error.errors, 'path')))
    }
    else {
      clearErrors()
    }

    return result.success
  }

  const getError = (...path: (string | number)[]) => {
    return errors.value.get(path.join(','))?.[0]
  }

  return { validate, errors, clearErrors, getError }
}

const getSubSchema = (schema: ZodTypeAny, property: string | number) => {
  if (schema instanceof ZodArray) {
    return schema.element
  }
  if (schema instanceof ZodTuple) {
    return schema.items[property]
  }
  if (!(schema instanceof ZodObject)) {
    return schema
  }
  return schema.shape[property]
}

export default useValidation
