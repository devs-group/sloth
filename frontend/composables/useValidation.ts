import { z, ZodArray, ZodObject } from 'zod'
import {  groupBy } from 'lodash-es'
import { ref } from 'vue'
import type { ZodTypeAny, ZodIssue } from 'zod'

function useValidation<T extends ZodTypeAny>(schema: T, data: z.output<typeof schema>) {
  const isValid = ref(true)
  const errors = ref<Map<string, ZodIssue[]>>(new Map())

  const clearErrors = () => {
    errors.value.clear()
  }

  const validateByPath = (path: (string| number)[]): boolean => {
    let subSchema = schema
    let subData = data
    for (let property of path) {
      subSchema = getSubSchema(subSchema, property)
      subData = subData[property]
    }

    let result = subSchema.safeParse(subData)

    isValid.value = result.success
    if (!result.success) {
      errors.value.set(path.join(","), result.error.issues)
    } else {
      errors.value.delete(path.join(','))
    }

    return result.success
  }

  const validate = (...path: (string| number)[]): boolean => {
    if (path.length > 0) {
      return validateByPath(path)
    }

    let result = schema.safeParse(data)
    isValid.value = result.success
    if (!result.success) {
      errors.value = new Map(Object.entries(groupBy(result.error.errors, "path")))
    } else {
      clearErrors()
    }

    return result.success
  }

  const getError = (...path: (string| number)[]) => {
    return errors.value.get(path.join(","))?.[0]
  }

  return { validate, errors, clearErrors, getError }
}

const getSubSchema = (schema: ZodTypeAny, property: string | number) => {
  if (schema instanceof ZodArray) {
    return schema.element
  }
  if (!(schema instanceof ZodObject)) {
    return schema
  }
  return (schema as ZodObject<any>).shape[property]
}


export default useValidation
