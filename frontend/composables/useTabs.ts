import { shallowRef, ref, computed } from 'vue';
import type { TabItem } from '~/config/interfaces';

export function useTabs(initialTabs: TabItem[]) {
  const tabs = ref<TabItem[]>(initialTabs);
  const activeTabComponent = ref(tabs.value[0].component);

  function onChangeTab(idx: number) {
    console.log("CHANGED TAB")
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
