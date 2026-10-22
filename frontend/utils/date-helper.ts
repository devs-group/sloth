import { Constants } from '~/config/const'

export const DateDiffInMilliseconds = (d1: Date, d2: Date) => {
  return d1.getTime() - d2.getTime()
}

export const DateDiffInDays = (d1: string | Date, d2: string | Date) => {
  const diffInMs = DateDiffInMilliseconds(d1 instanceof Date ? d1 : new Date(d1), d2 instanceof Date ? d2 : new Date(d2))
  return Math.round(diffInMs / (1000 * 60 * 60 * 24))
}

export const DateAddDays = (date: Date, days: number): Date => {
  const result = new Date(date)
  result.setDate(result.getDate() + days)
  return result
}

export const DateStringToFormattedDate = (date: string | Date, asUtc: boolean = true) => {
  const fmt = Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, { day: '2-digit', month: '2-digit', year: 'numeric', timeZone: asUtc ? 'UTC' : undefined })
  return fmt.format(date instanceof Date ? date : new Date(date))
}

export const DateStringToFormattedDateTime = (date: string | Date, asUtc: boolean = true) => {
  const fmt = Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, {
    day: '2-digit', month: '2-digit', year: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
    timeZone: asUtc ? 'UTC' : undefined,
  })
  return fmt.format(date instanceof Date ? date : new Date(date))
}

export const DateToUTCDate = (date: string | Date) => {
  const dateToFormat = date instanceof Date ? date : new Date(date)
  return new Date(dateToFormat.toLocaleDateString('en-US', { timeZone: 'UTC' }))
}

export const DateToEuropeZurichDate = (date: string | Date) => {
  const dateToFormat = date instanceof Date ? date : new Date(date)
  return new Date(dateToFormat.toLocaleDateString('en-US', { timeZone: 'Europe/Zurich' }))
}

export const DateToApiFormat = (date: string | Date) => {
  const dateToFormat = date instanceof Date ? date : new Date(date)

  const year = dateToFormat.getUTCFullYear() // Use UTC methods to avoid time zone shift
  const month = (dateToFormat.getUTCMonth() + 1).toString().padStart(2, '0') // Ensure two digits for the month
  const day = dateToFormat.getUTCDate().toString().padStart(2, '0') // Ensure two digits for the day

  return `${year}-${month}-${day}`
}
