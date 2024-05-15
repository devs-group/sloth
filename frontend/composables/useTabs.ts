import { shallowRef, ref, computed } from 'vue';
import type { TabItem } from '~/config/interfaces';

export function useTabs(initialTabs: TabItem[]) {
  const tabs = shallowRef<TabItem[]>(initialTabs);
  const activeTabComponent = shallowRef(tabs.value[0].component);

  function onChangeTab(idx: number) {
    if (tabs.value[idx].disabled) {
      console.log("This tab is currently disabled.");
      return;
    }
    activeTabComponent.value = tabs.value[idx].component;
  }

  const activeTabLabel = computed(() => 
    tabs.value.find(tab => tab.component === activeTabComponent.value)?.label
  );

  return {
    tabs,
    activeTabComponent,
    onChangeTab,
    activeTabLabel
  };
}
