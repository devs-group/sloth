import type { DialogProps } from 'primevue/dialog'

export const ModalConfig = {
  closable: true,
  closeOnEscape: true,
  blockScroll: true,
  keepInViewPort: true,
  maximizable: false,
  contentStyle: { 'display': 'flex', 'flex-direction': 'column', 'overflow-y': 'hidden' },
  style: {
    width: '96vw',
    height: '96vw',
  },
  pt: {
    root: {
      class: 'max-dialog-maximized-mobile',
    },
  },
  modal: true,
  draggable: false,
} as DialogProps
