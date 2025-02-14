import type { NavigationItems, TLogoutFunction } from '~/config/interfaces'
import { Routes } from '~/config/routes'

export const useMenu = () => {
  const getMainMenuItems = (data: { onLogout: TLogoutFunction }) => {
    return [
      {
        label: 'Projects',
        icon: 'heroicons-home',
        to: Routes.PROJECTS,
      },
      {
        label: 'Organisations',
        icon: 'heroicons-user-group',
        to: Routes.ORGANISATIONS,
      },
      {
        divider: true,
      },
      {
        label: 'Logout',
        icon: 'heroicons-arrow-left-on-rectangle',
        click: data.onLogout,
      },
    ] as NavigationItems[]
  }

  return {
    getMainMenuItems,
  }
}
