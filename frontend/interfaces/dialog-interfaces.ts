import type { Project, Service } from '~/schema/schema'

export interface IBaseDialog<T> {
  close: () => object
  value: {
    close: () => void
    data: T
  }
}

export interface ILogsDialogData {
  project: Project
  service: Service
}
