import { computed, ref } from 'vue'
import type { TabItem } from '~/config/interfaces'

export function useTabs(tabItems: Ref<TabItem[]>) {
  const activeTabIndex = ref(0)
  const activeTabComponent = computed(() => tabItems.value[activeTabIndex.value].component)
  const activeTabProps = computed(() => tabItems.value[activeTabIndex.value].props)

  function onChangeTab(idx: number) {
    activeTabIndex.value = idx
  }

  const activeTabLabel = computed(() =>
    tabItems.value[activeTabIndex.value].label,
  )

  return {
    activeTabComponent,
    activeTabProps,
    onChangeTab,
    activeTabLabel,
  }
}
