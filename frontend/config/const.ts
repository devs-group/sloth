export const Constants = {
  ToasterDefaultLifeTime: 4000,
  BASE_LOCALE_CODE: 'en-US',
}

export const DialogProps = {
  BigDialog: {
    contentClass: 'h-full',
    modal: true,
    draggable: false,
    closable: false,
    style: {
      width: '40vw',
    },
    breakpoints: {
      '1440px': '75vw',
      '640px': '96vw',
    },
  },
  SmallDialog: {
    contentClass: 'h-full',
    modal: true,
    draggable: false,
    closable: false,
    style: {
      width: '40vw',
    },
    breakpoints: {
      '1440px': '56vw',
      '640px': '96vw',
    },
  },
}
